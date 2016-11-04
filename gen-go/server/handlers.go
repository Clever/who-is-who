package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Clever/who-is-who/gen-go/models"
	"github.com/go-errors/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/gorilla/mux"
	"gopkg.in/Clever/kayvee-go.v5/logger"
)

var _ = strconv.ParseInt
var _ = strfmt.Default
var _ = swag.ConvertInt32
var _ = errors.New
var _ = mux.Vars
var _ = bytes.Compare
var _ = ioutil.ReadAll

var formats = strfmt.Default
var _ = formats

// convertBase64 takes in a string and returns a strfmt.Base64 if the input
// is valid base64 and an error otherwise.
func convertBase64(input string) (strfmt.Base64, error) {
	temp, err := formats.Parse("byte", input)
	if err != nil {
		return strfmt.Base64{}, err
	}
	return *temp.(*strfmt.Base64), nil
}

// convertDateTime takes in a string and returns a strfmt.DateTime if the input
// is a valid DateTime and an error otherwise.
func convertDateTime(input string) (strfmt.DateTime, error) {
	temp, err := formats.Parse("date-time", input)
	if err != nil {
		return strfmt.DateTime{}, err
	}
	return *temp.(*strfmt.DateTime), nil
}

// convertDate takes in a string and returns a strfmt.Date if the input
// is a valid Date and an error otherwise.
func convertDate(input string) (strfmt.Date, error) {
	temp, err := formats.Parse("date", input)
	if err != nil {
		return strfmt.Date{}, err
	}
	return *temp.(*strfmt.Date), nil
}

func jsonMarshalNoError(i interface{}) string {
	bytes, err := json.Marshal(i)
	if err != nil {
		// This should never happen
		return ""
	}
	return string(bytes)
}

// statusCodeForHealthCheck returns the status code corresponding to the returned
// object. It returns -1 if the type doesn't correspond to anything.
func statusCodeForHealthCheck(obj interface{}) int {

	switch obj.(type) {

	case models.DefaultBadRequest:
		return 400
	case models.DefaultInternalError:
		return 500
	default:
		return -1
	}
}

func (h handler) HealthCheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	err := h.HealthCheck(ctx)

	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		if btErr, ok := err.(*errors.Error); ok {
			logger.FromContext(ctx).AddContext("stacktrace", string(btErr.Stack()))
		}
		statusCode := statusCodeForHealthCheck(err)
		if statusCode != -1 {
			http.Error(w, err.Error(), statusCode)
		} else {
			http.Error(w, jsonMarshalNoError(models.DefaultInternalError{Msg: err.Error()}), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(""))

}

// newHealthCheckInput takes in an http.Request an returns the input struct.
func newHealthCheckInput(r *http.Request) (*models.HealthCheckInput, error) {
	var input models.HealthCheckInput

	var err error
	_ = err

	return &input, nil
}

// statusCodeForGetUserByAlias returns the status code corresponding to the returned
// object. It returns -1 if the type doesn't correspond to anything.
func statusCodeForGetUserByAlias(obj interface{}) int {

	switch obj.(type) {

	case *models.User:
		return 200

	case models.User:
		return 200

	case models.DefaultBadRequest:
		return 400
	case models.DefaultInternalError:
		return 500
	default:
		return -1
	}
}

func (h handler) GetUserByAliasHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	input, err := newGetUserByAliasInput(r)
	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		http.Error(w, jsonMarshalNoError(models.DefaultBadRequest{Msg: err.Error()}), http.StatusBadRequest)
		return
	}

	err = input.Validate()
	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		http.Error(w, jsonMarshalNoError(models.DefaultBadRequest{Msg: err.Error()}), http.StatusBadRequest)
		return
	}

	resp, err := h.GetUserByAlias(ctx, input)

	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		if btErr, ok := err.(*errors.Error); ok {
			logger.FromContext(ctx).AddContext("stacktrace", string(btErr.Stack()))
		}
		statusCode := statusCodeForGetUserByAlias(err)
		if statusCode != -1 {
			http.Error(w, err.Error(), statusCode)
		} else {
			http.Error(w, jsonMarshalNoError(models.DefaultInternalError{Msg: err.Error()}), http.StatusInternalServerError)
		}
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		http.Error(w, jsonMarshalNoError(models.DefaultInternalError{Msg: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCodeForGetUserByAlias(resp))
	w.Write(respBytes)

}

// newGetUserByAliasInput takes in an http.Request an returns the input struct.
func newGetUserByAliasInput(r *http.Request) (*models.GetUserByAliasInput, error) {
	var input models.GetUserByAliasInput

	var err error
	_ = err

	aliasTypeStr := mux.Vars(r)["alias_type"]
	if len(aliasTypeStr) == 0 {
		return nil, errors.New("Parameter must be specified")
	}
	if len(aliasTypeStr) != 0 {
		var aliasTypeTmp string
		aliasTypeTmp, err = aliasTypeStr, error(nil)
		if err != nil {
			return nil, err
		}
		input.AliasType = aliasTypeTmp

	}
	aliasValueStr := mux.Vars(r)["alias_value"]
	if len(aliasValueStr) == 0 {
		return nil, errors.New("Parameter must be specified")
	}
	if len(aliasValueStr) != 0 {
		var aliasValueTmp string
		aliasValueTmp, err = aliasValueStr, error(nil)
		if err != nil {
			return nil, err
		}
		input.AliasValue = aliasValueTmp

	}

	return &input, nil
}

// statusCodeForList returns the status code corresponding to the returned
// object. It returns -1 if the type doesn't correspond to anything.
func statusCodeForList(obj interface{}) int {

	switch obj.(type) {

	case *models.UserList:
		return 200

	case models.UserList:
		return 200

	case models.DefaultBadRequest:
		return 400
	case models.DefaultInternalError:
		return 500
	default:
		return -1
	}
}

func (h handler) ListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	resp, err := h.List(ctx)

	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		if btErr, ok := err.(*errors.Error); ok {
			logger.FromContext(ctx).AddContext("stacktrace", string(btErr.Stack()))
		}
		statusCode := statusCodeForList(err)
		if statusCode != -1 {
			http.Error(w, err.Error(), statusCode)
		} else {
			http.Error(w, jsonMarshalNoError(models.DefaultInternalError{Msg: err.Error()}), http.StatusInternalServerError)
		}
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.FromContext(ctx).AddContext("error", err.Error())
		http.Error(w, jsonMarshalNoError(models.DefaultInternalError{Msg: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCodeForList(resp))
	w.Write(respBytes)

}

// newListInput takes in an http.Request an returns the input struct.
func newListInput(r *http.Request) (*models.ListInput, error) {
	var input models.ListInput

	var err error
	_ = err

	return &input, nil
}
