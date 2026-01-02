package schema

// Primitive type helpers for common scalar types.
var (
	String  = Type{Kind: KindString}  // String is a UTF-8 string type.
	Int32   = Type{Kind: KindInt32}   // Int32 is a 32-bit signed integer type.
	Int64   = Type{Kind: KindInt64}   // Int64 is a 64-bit signed integer type.
	Float32 = Type{Kind: KindFloat32} // Float32 is a 32-bit floating point type.
	Float64 = Type{Kind: KindFloat64} // Float64 is a 64-bit floating point type.
	Bool    = Type{Kind: KindBool}    // Bool is a boolean type.
	Bytes   = Type{Kind: KindBytes}   // Bytes is a byte slice type.
)

// Repeated creates a repeated (array/slice) type with the given element type.
func Repeated(elem Type) Type {
	return Type{Kind: KindRepeated, Elem: &elem}
}

// Map creates a map type with the given key and value types.
func Map(key, value Type) Type {
	return Type{Kind: KindMap, Key: &key, Elem: &value}
}

// TypeBuilder provides a fluent API for building message types.
type TypeBuilder struct {
	t Type
}

// Message creates a new TypeBuilder for a message type with the given name.
func Message(name string) *TypeBuilder {
	return &TypeBuilder{
		t: Type{Kind: KindMessage, Name: name},
	}
}

// Field adds an optional field to the message type.
func (b *TypeBuilder) Field(name string, typ Type) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:   name,
		Type:   typ,
		Number: len(b.t.Fields) + 1,
	})
	return b
}

// RequiredField adds a required field to the message type.
func (b *TypeBuilder) RequiredField(name string, typ Type) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:     name,
		Type:     typ,
		Number:   len(b.t.Fields) + 1,
		Required: true,
	})
	return b
}

// FieldWithDefault adds a field with a default value to the message type.
func (b *TypeBuilder) FieldWithDefault(name string, typ Type, defaultVal any) *TypeBuilder {
	b.t.Fields = append(b.t.Fields, Field{
		Name:    name,
		Type:    typ,
		Number:  len(b.t.Fields) + 1,
		Default: defaultVal,
	})
	return b
}

// Build returns the constructed Type.
func (b *TypeBuilder) Build() Type {
	return b.t
}

// EnumBuilder provides a fluent API for building enum types.
type EnumBuilder struct {
	t      Type
	names  map[string]struct{}
	numbers map[int]struct{}
}

// Enum creates a new EnumBuilder for an enum type with the given name.
func Enum(name string) *EnumBuilder {
	return &EnumBuilder{
		t:       Type{Kind: KindEnum, Name: name},
		names:   make(map[string]struct{}),
		numbers: make(map[int]struct{}),
	}
}

// Value adds an enum value with the given name and number.
// Duplicate names or numbers are silently ignored.
func (b *EnumBuilder) Value(name string, number int) *EnumBuilder {
	if _, exists := b.names[name]; exists {
		return b
	}
	if _, exists := b.numbers[number]; exists {
		return b
	}
	b.names[name] = struct{}{}
	b.numbers[number] = struct{}{}
	b.t.Values = append(b.t.Values, EnumValue{Name: name, Number: number})
	return b
}

// Build returns the constructed enum Type.
func (b *EnumBuilder) Build() Type {
	return b.t
}
