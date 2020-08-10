package ddl

import (
	"fmt"
	"io"
)

type DataKind string

const (
	Int32     DataKind = "int32"
	Int64              = "int64"
	Bool               = "bool"
	String             = "string"
	Timestamp          = "timestamp"
	UUID               = "uuid"
)

type Column struct {
	doc      string
	name     string
	kind     DataKind
	len      int
	nullable bool
}

func NewColumn(name string, kind DataKind) *Column {
	c := &Column{name: name, kind: kind}
	return c
}

func (c *Column) AsMySQL(w io.Writer) (err error) {
	CheckPrintf(w, &err, "`%s` ", c.name)
	switch c.kind {
	case Int64:
		CheckPrint(w, &err, "BIGINT")
		if c.len != 0 {
			return fmt.Errorf("kind does not allow a length")
		}
	case String:
		CheckPrint(w, &err, "VARCHAR")
		if c.len != 0 {
			CheckPrintf(w, &err, "(%d)", c.len)
		}
	case Timestamp:
		CheckPrint(w, &err, "TIMESTAMP")
	default:
		panic("not yet implemented: " + string(c.kind))
	}

	if !c.nullable {
		CheckPrint(w, &err, " NOT NULL ")
	}

	if c.doc != "" {
		CheckPrintf(w, &err, " COMMENT '%s'", c.doc)
	}

	return
}

func (c *Column) AsPlantUML(w io.Writer) error {
	panic("implement me")
}

func (c *Column) Doc(doc string) *Column {
	c.doc = doc
	return c
}

func (c *Column) Len(i int) *Column {
	c.len = i
	return c
}

func (c *Column) Nullable() *Column {
	c.nullable = true
	return c
}

func (c *Column) NotNull() *Column {
	c.nullable = false
	return c
}

func (c *Column) Validate() error {
	if !isSafeName(c.name) {
		return fmt.Errorf("column name '%s' is not a valid identifier", c.name)
	}

	if !isSafeComment(c.doc) {
		return fmt.Errorf("column '%s' comment is not valid", c.doc)
	}

	return nil
}
