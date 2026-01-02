package schema

// MethodBuilder provides a fluent API for defining methods.
type MethodBuilder struct {
	method Method
}

func Unary(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodUnary},
	}
}

func ServerStream(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodServerStream},
	}
}

func ClientStream(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodClientStream},
	}
}

func Bidirectional(name string) *MethodBuilder {
	return &MethodBuilder{
		method: Method{Name: name, Type: MethodBidirectional},
	}
}

func (b *MethodBuilder) Input(t Type) *MethodBuilder {
	b.method.Input = t
	return b
}

func (b *MethodBuilder) Output(t Type) *MethodBuilder {
	b.method.Output = t
	return b
}

func (b *MethodBuilder) HTTP(method, path string) *MethodBuilder {
	b.method.HTTPMethod = method
	b.method.HTTPPath = path
	return b
}

func (b *MethodBuilder) Description(desc string) *MethodBuilder {
	b.method.Description = desc
	return b
}

func (b *MethodBuilder) Option(key string, value any) *MethodBuilder {
	if b.method.Options == nil {
		b.method.Options = make(map[string]any)
	}
	b.method.Options[key] = value
	return b
}

func (b *MethodBuilder) Build() Method {
	return b.method
}
