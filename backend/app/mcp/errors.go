package mcpserver

import "maps"

// ToolError is a structured error for MCP tool handlers. It carries a stable
// sentinel code and optional context data. Use errors.Is() to match by code.
// Always use With() or Wrap() to derive a copy — never mutate a sentinel.
type ToolError struct {
	code  string
	data  map[string]any
	cause error
}

// Sentinel errors for MCP tool operations.
var (
	ErrInvalidFormat = &ToolError{code: "invalid format"}
	ErrParseFailure  = &ToolError{code: "parse failure"}
)

func (e *ToolError) Error() string {
	if e.cause != nil {
		return e.code + ": " + e.cause.Error()
	}
	return e.code
}

func (e *ToolError) Is(target error) bool {
	t, ok := target.(*ToolError)
	if !ok {
		return false
	}
	return e.code == t.code
}

func (e *ToolError) Unwrap() error {
	return e.cause
}

// Code returns the stable error code string.
func (e *ToolError) Code() string { return e.code }

// Data returns the optional structured context data.
func (e *ToolError) Data() map[string]any { return e.data }

// With returns a copy with the given data merged into any existing data.
func (e *ToolError) With(data map[string]any) *ToolError {
	cp := *e
	if e.data == nil {
		cp.data = data
	} else {
		cp.data = make(map[string]any, len(e.data)+len(data))
		maps.Copy(cp.data, e.data)
		maps.Copy(cp.data, data)
	}
	return &cp
}

// Wrap returns a copy with the given cause, preserving existing data.
func (e *ToolError) Wrap(cause error) *ToolError {
	cp := *e
	cp.cause = cause
	return &cp
}
