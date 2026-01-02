package protokol

import "errors"

// Sentinel errors returned by protokol operations.
var (
	// ErrAlreadyRunning is returned when Run is called on an already running instance.
	ErrAlreadyRunning = errors.New("protokol: already running")
	// ErrServiceNotFound is returned when a requested service does not exist.
	ErrServiceNotFound = errors.New("protokol: service not found")
	// ErrMethodNotFound is returned when a requested method does not exist.
	ErrMethodNotFound = errors.New("protokol: method not found")
	// ErrBackendNotFound is returned when a backend is not registered.
	ErrBackendNotFound = errors.New("protokol: backend not found")
	// ErrStreamingNotSupported is returned when streaming is not supported by a backend.
	ErrStreamingNotSupported = errors.New("protokol: streaming not supported")
)
