package ddl

import (
	"io"
	"strconv"
	"strings"
)

// mysqlType returns the specific DDL type, e.g. VARCHAR(20)
func (c *Column) mysqlType() string {
	switch c.typeName {
	case Int8:
		fallthrough
	case Uint8:
		return "TINYINT"
	case Int16:
		fallthrough
	case Uint16:
		return "SMALLINT"
	case Int24:
		fallthrough
	case Uint24:
		return "MEDIUMINT"
	case Int32:
		fallthrough
	case Uint32:
		return "INT"
	case Int64:
		fallthrough
	case Uint64:
		return "BIGINT"
	case Float32:
		return "FLOAT"
	case Float64:
		return "FLOAT64"
	case Varchar:
		return "VARCHAR(" + strconv.Itoa(c.len) + ")"
	case Char:
		return "CHAR(" + strconv.Itoa(c.len) + ")"
	case Text:
		return "LONGTEXT"
	case Blob:
		return "LONGBLOB"
	case UUID:
		return "BINARY(16)"
	case Binary:
		return "BINARY(" + strconv.Itoa(c.len) + ")"
	case Timestamp:
		return "BIGINT"
	case Duration:
		return "BIGINT"
	case Enum:
		return "ENUM('" + strings.Join(c.enumValues, "', '") + "')"
	case Bool:
		return "BOOLEAN"
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
