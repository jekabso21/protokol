package protokol

import "errors"

var (
	ErrAlreadyRunning        = errors.New("protokol: already running")
	ErrServiceNotFound       = errors.New("protokol: service not found")
	ErrMethodNotFound        = errors.New("protokol: method not found")
	ErrBackendNotFound       = errors.New("protokol: backend not found")
	ErrStreamingNotSupported = errors.New("protokol: streaming not supported")
)
