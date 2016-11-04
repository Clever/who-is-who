package client

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	discovery "github.com/Clever/discovery-go"
	"github.com/Clever/who-is-who/gen-go/models"
	"github.com/afex/hystrix-go/hystrix"
)

var _ = json.Marshal
var _ = strings.Replace
var _ = strconv.FormatInt
var _ = bytes.Compare

// WagClient is used to make requests to the who-is-who service.
type WagClient struct {
	basePath    string
	requestDoer doer
	transport   *http.Transport
	timeout     time.Duration
	// Keep the retry doer around so that we can set the number of retries
	retryDoer *retryDoer
	// Keep the circuit doer around so that we can turn it on / off
	circuitDoer    *circuitBreakerDoer
	defaultTimeout time.Duration
}

var _ Client = (*WagClient)(nil)

// New creates a new client. The base path and http transport are configurable.
func New(basePath string) *WagClient {
	base := baseDoer{}
	tracing := tracingDoer{d: base}
	retry := retryDoer{d: tracing, defaultRetries: 1}
	circuit := &circuitBreakerDoer{
		d:     &retry,
		debug: true,
		// one circuit for each service + url pair
		circuitName: fmt.Sprintf("who-is-who-%s", shortHash(basePath)),
	}
	circuit.init()
	client := &WagClient{requestDoer: circuit, retryDoer: &retry, circuitDoer: circuit, defaultTimeout: 10 * time.Second,
		transport: &http.Transport{}, basePath: basePath}
	client.SetCircuitBreakerSettings(DefaultCircuitBreakerSettings)
	return client
}

// NewFromDiscovery creates a client from the discovery environment variables. This method requires
// the three env vars: SERVICE_WHO_IS_WHO_HTTP_(HOST/PORT/PROTO) to be set. Otherwise it returns an error.
func NewFromDiscovery() (*WagClient, error) {
	url, err := discovery.URL("who-is-who", "default")
	if err != nil {
		url, err = discovery.URL("who-is-who", "http") // Added fallback to maintain reverse compatibility
		if err != nil {
			return nil, err
		}
	}
	return New(url), nil
}

// WithRetries returns a new client that retries all GET operations until they either succeed or fail the
// number of times specified.
func (c *WagClient) WithRetries(retries int) *WagClient {
	c.retryDoer.defaultRetries = retries
	return c
}

// SetCircuitBreakerDebug puts the circuit
func (c *WagClient) SetCircuitBreakerDebug(b bool) {
	c.circuitDoer.debug = b
}

// CircuitBreakerSettings are the parameters that govern the client's circuit breaker.
type CircuitBreakerSettings struct {
	// MaxConcurrentRequests is the maximum number of concurrent requests
	// the client can make at the same time. Default: 100.
	MaxConcurrentRequests int
	// RequestVolumeThreshold is the minimum number of requests needed
	// before a circuit can be tripped due to health. Default: 20.
	RequestVolumeThreshold int
	// SleepWindow how long, in milliseconds, to wait after a circuit opens
	// before testing for recovery. Default: 5000.
	SleepWindow int
	// ErrorPercentThreshold is the threshold to place on the rolling error
	// rate. Once the error rate exceeds this percentage, the circuit opens.
	// Default: 90.
	ErrorPercentThreshold int
}

// DefaultCircuitBreakerSettings describes the default circuit parameters.
var DefaultCircuitBreakerSettings = CircuitBreakerSettings{
	MaxConcurrentRequests:  100,
	RequestVolumeThreshold: 20,
	SleepWindow:            5000,
	ErrorPercentThreshold:  90,
}

// SetCircuitBreakerSettings sets parameters on the circuit breaker. It must be
// called on application startup.
func (c *WagClient) SetCircuitBreakerSettings(settings CircuitBreakerSettings) {
	hystrix.ConfigureCommand(c.circuitDoer.circuitName, hystrix.CommandConfig{
		// redundant, with the timeout we set on the context, so set
		// this to something high and irrelevant
		Timeout:                100 * 1000,
		MaxConcurrentRequests:  settings.MaxConcurrentRequests,
		RequestVolumeThreshold: settings.RequestVolumeThreshold,
		SleepWindow:            settings.SleepWindow,
		ErrorPercentThreshold:  settings.ErrorPercentThreshold,
	})
}

// WithTimeout returns a new client that has the specified timeout on all operations. To make a single request
// have a timeout use context.WithTimeout as described here: https://godoc.org/golang.org/x/net/context#WithTimeout.
func (c *WagClient) WithTimeout(timeout time.Duration) *WagClient {
	c.defaultTimeout = timeout
	return c
}

// JoinByFormat joins a string array by a known format:
//	 csv: comma separated value (default)
//	 ssv: space separated value
//	 tsv: tab separated value
//	 pipes: pipe (|) separated value
func JoinByFormat(data []string, format string) string {
	if len(data) == 0 {
		return ""
	}
	var sep string
	switch format {
	case "ssv":
		sep = " "
	case "tsv":
		sep = "\t"
	case "pipes":
		sep = "|"
	default:
		sep = ","
	}
	return strings.Join(data, sep)
}

// HealthCheck makes a GET request to /_health.
// Checks if the service is healthy
func (c *WagClient) HealthCheck(ctx context.Context) error {
	path := c.basePath + "/_health"
	urlVals := url.Values{}
	var body []byte

	path = path + "?" + urlVals.Encode()

	client := &http.Client{Transport: c.transport}
	req, err := http.NewRequest("GET", path, bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	// Add the opname for doers like tracing
	ctx = context.WithValue(ctx, opNameCtx{}, "healthCheck")
	req = req.WithContext(ctx)
	// Don't add the timeout in a "doer" because we don't want to call "defer.cancel()"
	// until we've finished all the processing of the request object. Otherwise we'll cancel
	// our own request before we've finished it.
	if c.defaultTimeout != 0 {
		ctx, cancel := context.WithTimeout(req.Context(), c.defaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	resp, err := c.requestDoer.Do(client, req)

	if err != nil {
		return models.DefaultInternalError{Msg: err.Error()}
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return nil
	case 400:
		var output models.DefaultBadRequest

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return models.DefaultInternalError{Msg: err.Error()}
		}

		return output
	case 500:
		var output models.DefaultInternalError

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return models.DefaultInternalError{Msg: err.Error()}
		}

		return output

	default:
		return models.DefaultInternalError{Msg: "Unknown response"}
	}
}

// GetUserByAlias makes a GET request to /alias/{alias_type}/{alias_value}.
// Returns info for a user based on an alias
func (c *WagClient) GetUserByAlias(ctx context.Context, i *models.GetUserByAliasInput) (*models.User, error) {
	path := c.basePath + "/alias/{alias_type}/{alias_value}"
	urlVals := url.Values{}
	var body []byte

	path = strings.Replace(path, "{alias_type}", i.AliasType, -1)
	path = strings.Replace(path, "{alias_value}", i.AliasValue, -1)
	path = path + "?" + urlVals.Encode()

	client := &http.Client{Transport: c.transport}
	req, err := http.NewRequest("GET", path, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	// Add the opname for doers like tracing
	ctx = context.WithValue(ctx, opNameCtx{}, "getUserByAlias")
	req = req.WithContext(ctx)
	// Don't add the timeout in a "doer" because we don't want to call "defer.cancel()"
	// until we've finished all the processing of the request object. Otherwise we'll cancel
	// our own request before we've finished it.
	if c.defaultTimeout != 0 {
		ctx, cancel := context.WithTimeout(req.Context(), c.defaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	resp, err := c.requestDoer.Do(client, req)

	if err != nil {
		return nil, models.DefaultInternalError{Msg: err.Error()}
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var output models.User

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return &output, nil
	case 400:
		var output models.DefaultBadRequest

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return nil, output
	case 500:
		var output models.DefaultInternalError

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return nil, output

	default:
		return nil, models.DefaultInternalError{Msg: "Unknown response"}
	}
}

// List makes a GET request to /list.
// Returns info for all user
func (c *WagClient) List(ctx context.Context) (*models.UserList, error) {
	path := c.basePath + "/list"
	urlVals := url.Values{}
	var body []byte

	path = path + "?" + urlVals.Encode()

	client := &http.Client{Transport: c.transport}
	req, err := http.NewRequest("GET", path, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	// Add the opname for doers like tracing
	ctx = context.WithValue(ctx, opNameCtx{}, "list")
	req = req.WithContext(ctx)
	// Don't add the timeout in a "doer" because we don't want to call "defer.cancel()"
	// until we've finished all the processing of the request object. Otherwise we'll cancel
	// our own request before we've finished it.
	if c.defaultTimeout != 0 {
		ctx, cancel := context.WithTimeout(req.Context(), c.defaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	resp, err := c.requestDoer.Do(client, req)

	if err != nil {
		return nil, models.DefaultInternalError{Msg: err.Error()}
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var output models.UserList

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return &output, nil
	case 400:
		var output models.DefaultBadRequest

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return nil, output
	case 500:
		var output models.DefaultInternalError

		if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
			return nil, models.DefaultInternalError{Msg: err.Error()}
		}

		return nil, output

	default:
		return nil, models.DefaultInternalError{Msg: "Unknown response"}
	}
}

func shortHash(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))[0:6]
}
