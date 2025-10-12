package errors

import "fmt"

// ServiceUnavailableError indicates a service is unavailable (circuit breaker open)
type ServiceUnavailableError struct {
	Service string
	Err     error
}

func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service unavailable: %s - %v", e.Service, e.Err)
}

// DatabaseUnavailableError indicates database is unavailable
type DatabaseUnavailableError struct {
	Err error
}

func (e *DatabaseUnavailableError) Error() string {
	return fmt.Sprintf("database unavailable: %v", e.Err)
}

// ValidationError indicates invalid input
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NotFoundError indicates a resource was not found
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.Resource, e.ID)
}

// ConflictError indicates a resource conflict
type ConflictError struct {
	Resource string
	Message  string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Message)
}
