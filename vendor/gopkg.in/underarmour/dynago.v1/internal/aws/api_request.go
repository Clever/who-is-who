package aws

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const DynamoTargetPrefix = "DynamoDB_20120810." // This is the Dynamo API version we support
var MaxResponseSize int64 = 25 * 1024 * 1024    // 25MB maximum response
var MaxResponseError = errors.New("Exceeded maximum response size of 25MB")

type Signer interface {
	SignRequest(*http.Request, []byte)
}

/*
RequestMaker is the default AwsRequester used by Dynago.

The RequestMaker has its properties exposed as public to allow easier
construction. Directly modifying properties on the RequestMaker after
construction is not goroutine-safe so it should be avoided except for in
special cases (testing, mocking).
*/
type RequestMaker struct {
	// These are required to be set
	Endpoint   string
	Signer     Signer
	BuildError func(*http.Request, []byte, *http.Response, []byte) error

	// These can be optionally set
	Caller         http.Client
	DebugRequests  bool
	DebugResponses bool
	DebugFunc      func(string, ...interface{})
}

func (r *RequestMaker) MakeRequest(target string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", r.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if !strings.Contains(target, ".") {
		target = DynamoTargetPrefix + target
	}
	req.Header.Add("x-amz-target", target)
	req.Header.Add("content-type", "application/x-amz-json-1.0")
	req.Header.Set("Host", req.URL.Host)
	r.Signer.SignRequest(req, body)
	if r.DebugRequests {
		r.DebugFunc("Request:%#v\n\nRequest Body: %s\n\n", req, body)
	}
	response, err := r.Caller.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := responseBytes(response)
	if r.DebugResponses {
		r.DebugFunc("Response: %#v\nBody:%s\n", response, respBody)
	}
	if response.StatusCode != http.StatusOK {
		err = r.BuildError(req, body, response, respBody)
	}
	return respBody, err
}

func responseBytes(response *http.Response) (output []byte, err error) {
	if response.ContentLength != 0 {
		var buffer bytes.Buffer
		reader := io.LimitReader(response.Body, MaxResponseSize)
		if response.ContentLength > 0 {
			buffer.Grow(int(response.ContentLength)) // avoid a ton of allocations
		}
		var n int64
		n, err = io.Copy(&buffer, reader)
		if n >= MaxResponseSize {
			err = MaxResponseError
		} else if err == nil {
			output = buffer.Bytes()
			err = response.Body.Close()
		}
	}
	return
}

func FixEndpointUrl(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	if u.Scheme == "https" && !strings.Contains(u.Host, ":") {
		u.Host += ":443"
	}
	return u.String()
}
