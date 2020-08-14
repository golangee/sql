package ddl

import (
	"github.com/golangee/reflectplus/src"
)

// golangRow creates a type which is used to scan an entire single row. It also returns a bunch of enum types,
// if those have been defined at all. The emitted struct does intentionally never use sql.* types like sql.NullString
// because without it allows models which can directly be used in the SPI repository layer in DDD architectures.
// Any sql data type would be a violation of that principle. The downside is, that we have another layer of
// indirection, because we need to represent NULL values as nil pointer, so we have 3 layers of indirection for a
// string: a pointer to a string  struct, which itself holds a pointer to the actual byte data.
// A sql.NullString would be a bit more efficient because it represents the NULL information as a bool value
// and not with another heap pointer. Another reason to avoid sql.Null* is seamless json support.
func (s *Table) golangRow() (table *src.TypeBuilder, enums []*src.TypeBuilder) {
	table = src.NewStruct(snakeCaseToCamelCase(s.name)).SetDoc("... represents a single row of the SQL table " + s.name + ".\n" + s.doc)
	for _, column := range s.columns {
		typeDecl := golangToTypeDecl(column.typeName)
		if typeDecl == nil {
			switch column.typeName {
			case Enum:
				enumType := src.NewStringEnum(snakeCaseToCamelCase(column.name), column.enumValues...)
				enumType.SetDoc("...is based on the sql enum definition '" + column.mysqlType() + "'")
				enums = append(enums, enumType)
				typeDecl = src.NewTypeDecl(src.Qualifier(enumType.Name()))
			case Binary:
				typeDecl = src.NewArrayDecl(int64(column.len), src.NewTypeDecl("byte"))
			default:
				panic("illegal state")
			}
		}

		if column.nullable {
			typeDecl = src.NewPointerDecl(typeDecl)
		}
		table.AddFields(src.NewField(snakeCaseToCamelCase(column.name), typeDecl).
			AddTag("db", column.name).
			SetDoc("...represents the column '" + column.name + " " + column.mysqlType() + "'.\n" + column.doc),
		)
	}
	return
}

// goToTypeDecl inspects the given DataType and converts it into the according go type.
// It returns nil, if a new type needs to be generated to represent the data type. This currently only
// happens for enums and binary.
func golangToTypeDecl(kind DataType) *src.TypeDecl {
	switch kind {
	case Int8:
		return src.NewTypeDecl("int8")
	case Uint8:
		return src.NewTypeDecl("uint8")
	case Int16:
		return src.NewTypeDecl("int16")
	case Uint16:
		return src.NewTypeDecl("uint16")
	case Int24:
		return src.NewTypeDecl("int32")
	case Uint24:
		return src.NewTypeDecl("uint32")
	case Int32:
		return src.NewTypeDecl("int32")
	case Uint32:
		return src.NewTypeDecl("uint32")
	case Int64:
		return src.NewTypeDecl("int64")
	case Uint64:
		return src.NewTypeDecl("uint64")
	case Float32:
		return src.NewTypeDecl("float32")
	case Float64:
		return src.NewTypeDecl("float64")
	case Varchar:
		return src.NewTypeDecl("string")
	case Char:
		return src.NewTypeDecl("string")
	case Text:
		return src.NewTypeDecl("string")
	case Blob:
		return src.NewSliceDecl(src.NewTypeDecl("byte"))
	case UUID:
		//return src.NewArrayDecl(16, src.NewTypeDecl("byte"))
		return src.NewTypeDecl("github.com/golangee/uuid.UUID")
	case Binary:
		return nil
	case Timestamp:
		return src.NewTypeDecl("int64") //TODO introduce custom type like uuid with build-in support
	case Duration:
		return src.NewTypeDecl("time.Duration")
	case Enum:
		return nil
	case Bool:
		return src.NewTypeDecl("bool")
	default:
		panic("not yet implemented: " + string(kind))
	}

}
