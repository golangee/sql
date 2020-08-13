package ddl

import (
	"github.com/golangee/reflectplus/src"
)

func (s *Table) golangRow() (table *src.TypeBuilder, enums []*src.TypeBuilder) {
	table = src.NewStruct(snakeCaseToCamelCase(s.name)).SetDoc("... represents a single row of the SQL table " + s.name + ".\n" + s.doc)
	for _, column := range s.columns {
		typeDecl := golangToTypeDecl(column.typeName)
		if typeDecl == nil {
			//oops, we need an enum
			enumType := src.NewStringEnum(snakeCaseToCamelCase(column.name), column.enumValues...)
			enumType.SetDoc("...is based on the sql enum definition '" + column.mysqlType() + "'")
			enums = append(enums, enumType)
			typeDecl = src.NewTypeDecl(src.Qualifier(enumType.Name()))
		}

		table.AddFields(src.NewField(snakeCaseToCamelCase(column.name), typeDecl).
			AddTag("db", column.name).
			SetDoc("...represents the column '" + column.name + " " + column.mysqlType() + "'.\n" + column.doc),
		)
	}
	return
}

// goToTypeDecl inspects the given dataType and converts it into the according go type.
// It returns nil, if a new type needs to be generated to represent the data type. This currently only
// happens for enums.
func golangToTypeDecl(kind dataType) *src.TypeDecl {
	switch kind {
	case "int8":
		return src.NewTypeDecl("int8")
	case "uint8":
		return src.NewTypeDecl("uint8")
	case "int16":
		return src.NewTypeDecl("int16")
	case "uint16":
		return src.NewTypeDecl("uint16")
	case "int24":
		return src.NewTypeDecl("int32")
	case "uint24":
		return src.NewTypeDecl("uint32")
	case "int32":
		return src.NewTypeDecl("int32")
	case "uint32":
		return src.NewTypeDecl("uint32")
	case "int64":
		return src.NewTypeDecl("int64")
	case "uint64":
		return src.NewTypeDecl("uint64")
	case "float32":
		return src.NewTypeDecl("float32")
	case "float64":
		return src.NewTypeDecl("float64")
	case "varchar":
		return src.NewTypeDecl("string")
	case "char":
		return src.NewTypeDecl("string")
	case "text":
		return src.NewTypeDecl("string")
	case "blob":
		return src.NewSliceDecl(src.NewTypeDecl("byte"))
	case "uuid":
		//return src.NewArrayDecl(16, src.NewTypeDecl("byte"))
		return src.NewTypeDecl("github.com/golangee/uuid.UUID")
	case "binary":
		return src.NewSliceDecl(src.NewTypeDecl("byte"))
	case "timestamp":
		return src.NewTypeDecl("int64") //TODO introduce custom type like uuid with build-in support
	case "duration":
		return src.NewTypeDecl("time.Duration")
	case "enum":
		return nil
	default:
		panic("not yet implemented: " + string(kind))
	}

}
