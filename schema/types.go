// Package schema provides types for defining API schemas.
package schema

// Kind represents the fundamental type category.
type Kind int

// Kind constants for primitive and composite types.
const (
	KindInvalid  Kind = iota // KindInvalid represents an uninitialized or invalid type.
	KindBool                 // KindBool represents a boolean type.
	KindInt32                // KindInt32 represents a 32-bit signed integer.
	KindInt64                // KindInt64 represents a 64-bit signed integer.
	KindFloat32              // KindFloat32 represents a 32-bit floating point number.
	KindFloat64              // KindFloat64 represents a 64-bit floating point number.
	KindString               // KindString represents a UTF-8 string.
	KindBytes                // KindBytes represents a byte slice.
	KindMessage              // KindMessage represents a structured message type.
	KindEnum                 // KindEnum represents an enumeration type.
	KindMap                  // KindMap represents a key-value map type.
	KindRepeated             // KindRepeated represents a repeated/array type.
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

// IsScalar returns true if the type is a primitive scalar type.
func (t Type) IsScalar() bool {
	switch t.Kind {
	case KindBool, KindInt32, KindInt64, KindFloat32, KindFloat64, KindString, KindBytes:
		return true
	default:
		return false
	}
}

// IsComposite returns true if the type is a composite type (message, map, or repeated).
func (t Type) IsComposite() bool {
	switch t.Kind {
	case KindMessage, KindMap, KindRepeated:
		return true
	default:
		return false
	}
}
