package models

import (
	"encoding/json"
	"strconv"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// These imports may not be used depending on the input parameters
var _ = json.Marshal
var _ = strconv.FormatInt
var _ = validate.Maximum
var _ = strfmt.NewFormats

// HealthCheckInput holds the input parameters for a healthCheck operation.
type HealthCheckInput struct {
}

// Validate returns an error if any of the HealthCheckInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i HealthCheckInput) Validate() error {
	return nil
}

// GetUserByAliasInput holds the input parameters for a getUserByAlias operation.
type GetUserByAliasInput struct {
	AliasType  string
	AliasValue string
}

// Validate returns an error if any of the GetUserByAliasInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetUserByAliasInput) Validate() error {
	return nil
}

// ListInput holds the input parameters for a list operation.
type ListInput struct {
}

// Validate returns an error if any of the ListInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i ListInput) Validate() error {
	return nil
}
