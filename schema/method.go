package schema

// MethodType indicates the streaming behavior of a method.
type MethodType int

const (
	MethodUnary MethodType = iota
	MethodServerStream
	MethodClientStream
	MethodBidirectional
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

func (m Method) IsStreaming() bool {
	return m.Type != MethodUnary
}

func (m Method) IsServerStreaming() bool {
	return m.Type == MethodServerStream || m.Type == MethodBidirectional
}

func (m Method) IsClientStreaming() bool {
	return m.Type == MethodClientStream || m.Type == MethodBidirectional
}
