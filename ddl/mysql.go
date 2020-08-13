package ddl

import (
	"io"
	"strconv"
	"strings"
)



// mysqlType returns the specific DDL type, e.g. VARCHAR(20)
func (c *Column) mysqlType() string {
	switch c.typeName {
	case "int8":
		fallthrough
	case "uint8":
		return "TINYINT"
	case "int16":
		fallthrough
	case "uint16":
		return "SMALLINT"
	case "int24":
		fallthrough
	case "uint24":
		return "MEDIUMINT"
	case "int32":
		fallthrough
	case "uint32":
		return "INT"
	case "int64":
		fallthrough
	case "uint64":
		return "BIGINT"
	case "float32":
		return "FLOAT"
	case "float64":
		return "FLOAT64"
	case "varchar":
		return "VARCHAR(" + strconv.Itoa(c.len) + ")"
	case "char":
		return "CHAR(" + strconv.Itoa(c.len) + ")"
	case "text":
		return "LONGTEXT"
	case "blob":
		return "LONGBLOB"
	case "uuid":
		return "BINARY(16)"
	case "binary":
		return "BINARY(" + strconv.Itoa(c.len) + ")"
	case "timestamp":
		return "BIGINT"
	case "duration":
		return "BIGINT"
	case "enum":
		return "ENUM('" + strings.Join(c.enumValues, "', '") + "')"
	default:
		panic("not yet implemented: " + string(c.typeName))
	}
}

// mysql returns the full column DDL, including name, type and comment, e.g.
// `my_col` VARCHAR(20) NOT NULL COMMENT 'A required column to do better.'
func (c *Column) mysql(wr io.Writer) error {
	w := strWriter{Writer: wr}
	w.Print("`")
	w.Print(c.name)
	w.Print("` ")
	w.Print(c.mysqlType())

	if !c.nullable {
		w.Print(" NOT NULL")
	}

	if c.doc != "" {
		w.Print(" COMMENT '")
		w.Print(c.doc)
		w.Print("'")
	}

	return w.Err
}
