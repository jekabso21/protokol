package schema

// MethodBuilder provides a fluent API for defining methods.
type MethodBuilder struct {
	method Method
}

// Unary creates a new MethodBuilder for a unary (request-response) method.
func Unary(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodUnary},
	}
}

// ServerStream creates a new MethodBuilder for a server-streaming method.
func ServerStream(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodServerStream},
	}
}

// ClientStream creates a new MethodBuilder for a client-streaming method.
func ClientStream(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodClientStream},
	}
}

// Bidirectional creates a new MethodBuilder for a bidirectional streaming method.
func Bidirectional(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodBidirectional},
	}
}

// Input sets the input type for the method.
func (b *MethodBuilder) Input(t Type) *MethodBuilder {
	b.method.Input = t
	return b
}

// Output sets the output type for the method.
func (b *MethodBuilder) Output(t Type) *MethodBuilder {
	b.method.Output = t
	return b
}

// HTTP sets the HTTP method and path for REST exposure.
func (b *MethodBuilder) HTTP(method, path string) *MethodBuilder {
	b.method.HTTPMethod = method
	b.method.HTTPPath = path
	return b
}

// Description sets a human-readable description for the method.
func (b *MethodBuilder) Description(desc string) *MethodBuilder {
	b.method.Description = desc
	return b
}

// Option sets a custom option on the method.
func (b *MethodBuilder) Option(key string, value any) *MethodBuilder {
	if b.method.Options == nil {
		b.method.Options = make(map[string]any)
	}
	b.method.Options[key] = value
	return b
}

// Build returns the constructed Method.
func (b *MethodBuilder) Build() Method {
	return b.method
}
