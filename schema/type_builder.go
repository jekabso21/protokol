package schema

// Primitive type helpers.
var (
	String  = Type{Kind: KindString}
	Int32   = Type{Kind: KindInt32}
	Int64   = Type{Kind: KindInt64}
	Float32 = Type{Kind: KindFloat32}
	Float64 = Type{Kind: KindFloat64}
	Bool    = Type{Kind: KindBool}
	Bytes   = Type{Kind: KindBytes}
)

func Repeated(elem Type) Type {
	return Type{Kind: KindRepeated, Elem: &elem}
}

func Map(key, value Type) Type {
	return Type{Kind: KindMap, Key: &key, Elem: &value}
}

// TypeBuilder provides a fluent API for building message types.
type TypeBuilder struct {
	t Type
}

func Message(name string) *TypeBuilder {
	return &TypeBuilder{
		t: Type{Kind: KindMessage, Name: name},
	}
}

func (b *TypeBuilder) Field(name string, typ Type) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:   name,
		Type:   typ,
		Number: len(b.t.Fields) + 1,
	})
	return b
}

func (b *TypeBuilder) RequiredField(name string, typ Type) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:     name,
		Type:     typ,
		Number:   len(b.t.Fields) + 1,
		Required: true,
	})
	return b
}

func (b *TypeBuilder) FieldWithDefault(name string, typ Type, defaultVal any) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:    name,
		Type:    typ,
		Number:  len(b.t.Fields) + 1,
		Default: defaultVal,
	})
	return b
}

func (b *TypeBuilder) Build() Type {
	return b.t
}

// EnumBuilder provides a fluent API for building enum types.
type EnumBuilder struct {
	t Type
}

func Enum(name string) *EnumBuilder {
	return &EnumBuilder{
		t: Type{Kind: KindEnum, Name: name},
	}
}

func (b *EnumBuilder) Value(name string, number int) *EnumBuilder {
	b.t.Values = append(b.t.Values, EnumValue{Name: name, Number: number})
	return b
}

func (b *EnumBuilder) Build() Type {
	return b.t
}
