package toolapi

import (
	"context"
)

// Tool defines the generic interface for all OpenShannon tools
type Tool interface {
	// Name returns the identifier of the tool (e.g., "Read")
	Name() string

	// Description returns a brief explanation of what the tool does
	Description() string

	// Execute performs the tool's logic given the context and input arguments
	// The implementation must parse the input as per its schema
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)

	// InputSchema returns the expected input JSON schema (or struct mapping)
	InputSchema() map[string]interface{}
}
