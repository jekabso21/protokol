package schema

// MethodType indicates the streaming behavior of a method.
type MethodType int

// MethodType constants for different streaming modes.
const (
	MethodUnary       MethodType = iota // MethodUnary is a simple request-response method.
	MethodServerStream                   // MethodServerStream sends multiple responses for one request.
	MethodClientStream                   // MethodClientStream receives multiple requests for one response.
	MethodBidirectional                  // MethodBidirectional supports streaming in both directions.
)

// Method represents a single RPC method.
type Method struct {
	Name        string
	Input       Type
	Output      Type
	Type        MethodType
	Description string
	HTTPMethod  string
	HTTPPath    string
	Options     map[string]any
}

// IsStreaming returns true if the method uses any form of streaming.
func (m Method) IsStreaming() bool {
	return m.Type != MethodUnary
}

// IsServerStreaming returns true if the server can send multiple responses.
func (m Method) IsServerStreaming() bool {
	return m.Type == MethodServerStream || m.Type == MethodBidirectional
}

// IsClientStreaming returns true if the client can send multiple requests.
func (m Method) IsClientStreaming() bool {
	return m.Type == MethodClientStream || m.Type == MethodBidirectional
}
