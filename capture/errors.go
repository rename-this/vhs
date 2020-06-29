package capture

import (
	"bytes"
	"fmt"
	"sync"
)

// Error represents multiple errors that could
// occur while listening on a set of interfaces.
type Error struct {
	mu              sync.RWMutex
	InterfaceErrors []*InterfaceError
}

// Append appends an error to this Error. Nil errors are ignored.
func (e *Error) Append(err *InterfaceError) {
	if err == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.InterfaceErrors = append(e.InterfaceErrors, err)
}

// Error returns an error string.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var buf bytes.Buffer
	for _, err := range e.InterfaceErrors {
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// NewInterfaceError returns a new InterfaceError that wraps err.
// If err is nil then NewInterfaceError returns nil.
func NewInterfaceError(name string, err error) *InterfaceError {
	if err == nil {
		return nil
	}

	return &InterfaceError{
		Name: name,
		Err:  err,
	}
}

// InterfaceError represents an error on a specific interface
type InterfaceError struct {
	Name string
	Err  error
}

// Error returns an error string.
func (e *InterfaceError) Error() string {
	return fmt.Sprintf(fmt.Sprintf("failed to listen on %s: %v\n", e.Name, e.Err))
}
