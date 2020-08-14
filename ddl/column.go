package ddl

import (
	"fmt"
)

// DataType contains our own hardcoded types. It reads like a mixture of go type names and mysql type names.
type DataType string

const (
	Int8      DataType = "int8"
	Uint8     DataType = "uint8"
	Int16     DataType = "int16"
	Uint16    DataType = "uint16"
	Int24     DataType = "int24"
	Uint24    DataType = "uint24"
	Int32     DataType = "int32"
	Uint32    DataType = "uint32"
	Int64     DataType = "int64"
	Uint64    DataType = "uint64"
	Float32   DataType = "float32"
	Float64   DataType = "float64"
	Varchar   DataType = "varchar"
	Char      DataType = "char"
	Text      DataType = "text"
	Blob      DataType = "blob"
	UUID      DataType = "uuid"
	Binary    DataType = "binary"
	Timestamp DataType = "timestamp"
	Duration  DataType = "duration"
	Enum      DataType = "enum"
	Bool      DataType = "bool"
	Json      DataType = "json"
)

// Column is a model to describe a strongly typed field or column in a relational database like MariaDB or postgreSQL.
type Column struct {
	doc        string
	name       string
	len        int
	nullable   bool
	enumValues []string
	typeName   DataType
}

// NewColumn create a new but still invalid column. You need to define at the kind.
func NewColumn(name string) *Column {
	c := &Column{name: name}
	return c
}

// Int8 declares a 1 byte integer storage type. Range is -128...127.
func (c *Column) Int8() *Column {
	c.typeName = Int8
	return c
}

// Uint8 declares a 1 byte integer storage type. Range is 0...255.
func (c *Column) Uint8() *Column {
	c.typeName = Uint8
	return c
}

// TinyInt is an alias for Int8
func (c *Column) TinyInt() *Column {
	return c.Int8()
}

// Int16 declares a 2 byte integer storage type. Range is  -32768...32767.
func (c *Column) Int16() *Column {
	c.typeName = Int16
	return c
}

// Uint16 declares a 2 byte integer storage type. Range is 0...65535.
func (c *Column) Uint16() *Column {
	c.typeName = Uint16
	return c
}

// SmallInt is an alias for Int16.
func (c *Column) SmallInt() *Column {
	return c.Int16()
}

// Int24 declares a 3 byte integer storage type. Range is  -8388608...8388607.
func (c *Column) Int24() *Column {
	c.typeName = Int24 // this type does not exist and is mapped to int32 in code generator
	return c
}

// Uint24 declares a 3 byte integer storage type. Range is 0...16777215.
func (c *Column) Uint24() *Column {
	c.typeName = Uint24 // this type does not exist and is mapped to uint32 in code generator
	return c
}

// MediumInt is an Alias for Int24
func (c *Column) MediumInt() *Column {
	return c.Int24()
}

// Int32 declares a 4 byte integer storage type. Range is  -2147483648...2147483647.
func (c *Column) Int32() *Column {
	c.typeName = Int32
	return c
}

// Uint32 declares a 4 byte integer storage type. Range is 0...4294967295.
func (c *Column) Uint32() *Column {
	c.typeName = Uint32
	return c
}

// Int declares a 4 byte integer storage type. Range is either -2147483648...2147483647 or 0...4294967295.
func (c *Column) Int() *Column {
	return c
}

// Int64 declares an 8 byte integer storage type. Range is -9.223.372.036.854.775.808...9.223.372.036.854.775.807.
func (c *Column) Int64() *Column {
	c.typeName = Int64
	return c
}

// Uint64 declares an 8 byte integer storage type. Range is 0...18.446.744.073.709.551.615.
func (c *Column) Uint64() *Column {
	c.typeName = Uint64
	return c
}

// BigInt is an alias for Int64.
func (c *Column) BigInt() *Column {
	return c
}

// Float32 represents a single precision 32bit float.
func (c *Column) Float32() *Column {
	c.typeName = Float32
	return c
}

// Float is an alias for Float32
func (c *Column) Float() *Column {
	return c.Float32()
}

// Float64 represents a double precision 32bit float.
func (c *Column) Float64() *Column {
	c.typeName = Float32
	return c
}

// Double is an alias for Float64
func (c *Column) Double() *Column {
	return c.Float64()
}

// Varchar saves a string of at most max bytes within the row. What max can be is subject to the actual database.
// You should keep max in the range of at most a few thousand, better hundreds or below. The encoding is UTF-8 and
// the collation (sorting and index order) should be as close to an actual unicode standard as possible.
func (c *Column) Varchar(max int) *Column {
	c.len = max
	c.typeName = Varchar
	return c
}

// Char saves a string and pads it with spaces to match the length. See also Varchar and Text.
func (c *Column) Char(length int) *Column {
	c.len = length
	c.typeName = Char
	return c
}

// Text defines a usually out-of-row storage for an "arbitrary" amount of text. However this all depends on the
// actual database. For postgreSQL this is technically the same as Varchar but for MySQL it is less efficient.
func (c *Column) Text() *Column {
	c.typeName = Text
	return c
}

// Blob declares an usually uninterpreted and arbitrary amount of bytes to store. The implementation will expect
// that it fits still into memory. Usually the data is stored outside of a row, but keep in mind, that databases
// usually really poor when handling large blobs.
func (c *Column) Blob() *Column {
	c.typeName = Blob
	return c
}

// Binary is similar to Char but contains uninterpreted bytes instead utf-8 chars and fixed length.
func (c *Column) Binary(length int) *Column {
	c.typeName = Binary
	c.len = length
	return c
}

// UUID represents a UUID if the database supports that format natively. Otherwise the best fitting datatype is used.
func (c *Column) UUID() *Column {
	c.typeName = UUID
	return c
}

// Timestamp does not map to any build-in sql type, because the behavior is not even close to a common standard.
// The implementation always uses a signed 8 byte integer to represent a unix epoch timestamp in millisecond resolution.
func (c *Column) Timestamp() *Column {
	c.typeName = Timestamp
	return c
}

// Duration is a span of time in nanoseconds, so overflows after 290 years.
func (c *Column) Duration() *Column {
	c.typeName = Duration
	return c
}

// Enum declares a custom type, which can only hold one of the given values. The advantage is, that the database
// does the consistency check and can optimize storage.
func (c *Column) Enum(values ...string) *Column {
	c.typeName = Enum
	c.enumValues = values
	return c
}

func (c *Column) Bool() *Column {
	c.typeName = Bool
	return c
}

// Comment adds a note about the purpose of the column. If the database supports it, the comment will be issued into
// the concrete DDL.
func (c *Column) Comment(doc string) *Column {
	c.doc = doc
	return c
}

// Nullable is for most sql based systems the default setting but not for us: by default, any column is not nullable.
func (c *Column) Nullable() *Column {
	c.nullable = true
	return c
}

// NotNull is the default setting.
func (c *Column) NotNull() *Column {
	c.nullable = false
	return c
}

// Validate checks various conditions and returns a descriptive error for the first violation which has been found.
func (c *Column) Validate() error {
	if c.typeName == "" {
		return fmt.Errorf("column '%s' has no type", c.name)
	}

	if !isSafeName(c.name) {
		return fmt.Errorf("column name '%s' is not a valid identifier", c.name)
	}

	if !isSafeComment(c.doc) {
		return fmt.Errorf("column '%s' comment is not valid", c.doc)
	}

	return nil
}
