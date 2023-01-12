package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// ValidateResponse validates whether the given response object conforms
// to the response schema (schema.Fields). It cycles through the data map and
// validates conversions in the schema. In "strict" mode, this function will
// also ensure that the data map has all schema-required fields and does not
// have any fields outside of the schema.
//
// This function is inefficient and is intended to be used in tests only.
func ValidateResponse(schema *framework.Response, response *logical.Response, strict bool) error {
	if response != nil {
		return ValidateResponseData(schema, response.Data, strict)
	}

	return ValidateResponseData(schema, nil, strict)
}

// ValidateResponseData validates whether the given response data map conforms
// to the response schema (schema.Fields). It cycles through the data map and
// validates conversions in the schema. In "strict" mode, this function will
// also ensure that the data map has all schema-required fields and does not
// have any fields outside of the schema.
//
// This function is inefficient and is intended to be used in tests only.
func ValidateResponseData(schema *framework.Response, data map[string]interface{}, strict bool) error {
	// nothing to validate
	if schema == nil {
		return nil
	}

	// Marshal the data to JSON and back to convert the map's values into
	// JSON strings expected by Validate() and ValidateStrict(). This is
	// not efficient and is done for testing purposes only.
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to convert input to json: %w", err)
	}

	var dataWithStringValues map[string]interface{}
	if err := json.Unmarshal(
		jsonBytes,
		&dataWithStringValues,
	); err != nil {
		return fmt.Errorf("failed to unmashal data: %w", err)
	}

	// Validate
	fd := framework.FieldData{
		Raw:    dataWithStringValues,
		Schema: schema.Fields,
	}

	if strict {
		return fd.ValidateStrict()
	}

	return fd.Validate()
}

// FindResponseSchema is a test helper to extract the response schema from a given framework path / operation
func FindResponseSchema(t *testing.T, paths []*framework.Path, pathIdx int, operation logical.Operation) *framework.Response {
	t.Helper()

	if pathIdx >= len(paths) {
		t.Fatalf("path index %d is out of range", pathIdx)
	}

	schemaPath := paths[pathIdx]

	schemaOperation, ok := schemaPath.Operations[operation]
	if !ok {
		t.Fatalf(
			"could not find response schema: %s: %q operation does not exist",
			schemaPath.Pattern,
			operation,
		)
	}

	var schemaResponses []framework.Response

	for _, status := range []int{
		http.StatusOK,
		http.StatusNoContent,
	} {
		schemaResponses, ok = schemaOperation.Properties().Responses[status]
		if ok {
			break
		}
	}

	if len(schemaResponses) == 0 {
		t.Fatalf(
			"could not find response schema: %s: %q operation: no responses found",
			schemaPath.Pattern,
			operation,
		)
	}

	return &schemaResponses[0]
}