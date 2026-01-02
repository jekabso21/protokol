package schema

// Kind represents the fundamental type category.
type Kind int

const (
	KindInvalid Kind = iota
	KindBool
	KindInt32
	KindInt64
	KindFloat32
	KindFloat64
	KindString
	KindBytes
	KindMessage
	KindEnum
	KindMap
	KindRepeated
)

// Type represents a field's type information.
type Type struct {
	Kind   Kind
	Name   string
	Elem   *Type       // For repeated/map: element type
	Key    *Type       // For map: key type
	Fields []Field     // For message: nested fields
	Values []EnumValue // For enum: possible values
}

// Field represents a single field in a message.
type Field struct {
	Name     string
	Type     Type
	Number   int
	Required bool
	Default  any
}

// EnumValue represents a single enum option.
type EnumValue struct {
	Name   string
	Number int
}

func (t Type) IsScalar() bool {
	switch t.Kind {
	case KindBool, KindInt32, KindInt64, KindFloat32, KindFloat64, KindString, KindBytes:
		return true
	default:
		return false
	}
}

func (t Type) IsComposite() bool {
	switch t.Kind {
	case KindMessage, KindMap, KindRepeated:
		return true
	default:
		return false
	}
}
