package tyr

import (
	"reflect"
	"strings"

	"github.com/kubuskotak/tyr/dialect"
)

// InsertStmt builds `INSERT INTO ...`.
type InsertStmt struct {
	Dialect

	raw

	Table        string
	Column       []string
	Value        [][]interface{}
	Ignored      bool
	ReturnColumn []string
	RecordID     *int64
	comments     Comments
}

type InsertBuilder = InsertStmt

func (b *InsertStmt) ToSQL(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.ToSQL(d, buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	err := b.comments.ToSQL(d, buf)
	if err != nil {
		return err
	}

	if b.Ignored {
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}

	buf.WriteString(d.QuoteIdent(b.Table))

	var placeholderBuf strings.Builder
	placeholderBuf.WriteString("(")
	buf.WriteString(" (")
	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(",")
			placeholderBuf.WriteString(",")
		}
		buf.WriteString(d.QuoteIdent(col))
		placeholderBuf.WriteString(placeholder)
	}
	buf.WriteString(")")

	if d == dialect.MSSQL && len(b.ReturnColumn) > 0 {
		buf.WriteString(" OUTPUT ")
		for i, col := range b.ReturnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("INSERTED." + d.QuoteIdent(col))
		}
	}

	buf.WriteString(" VALUES ")
	placeholderBuf.WriteString(")")
	placeholderStr := placeholderBuf.String()

	for i, tuple := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(placeholderStr)

		buf.WriteValue(tuple...)
	}

	if d != dialect.MSSQL && len(b.ReturnColumn) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.ReturnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(d.QuoteIdent(col))
		}
	}

	return nil
}

// InsertInto creates an InsertStmt.
func InsertInto(table string) *InsertStmt {
	return &InsertStmt{
		Table: table,
	}
}

// InsertBySql creates an InsertStmt from raw query.
func InsertBySql(query string, value ...interface{}) *InsertStmt {
	return &InsertStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
	}
}

func (b *InsertStmt) Columns(column ...string) *InsertStmt {
	b.Column = column
	return b
}

// Comment adds a comment to prepended. All multi-line sql comment characters are stripped
func (b *InsertStmt) Comment(comment string) *InsertStmt {
	b.comments = b.comments.Append(comment)
	return b
}

// Ignore any insertion errors
func (b *InsertStmt) Ignore() *InsertStmt {
	b.Ignored = true
	return b
}

// Values adds a tuple to be inserted.
// The order of the tuple should match Columns.
func (b *InsertStmt) Values(value ...interface{}) *InsertStmt {
	b.Value = append(b.Value, value)
	return b
}

// Record adds a tuple for columns from a struct.
//
// If there is a field called "Id" or "ID" in the struct,
// it will be set to LastInsertId.
func (b *InsertStmt) Record(structValue interface{}) *InsertStmt {
	v := reflect.Indirect(reflect.ValueOf(structValue))

	if v.Kind() == reflect.Struct {
		found := make([]interface{}, len(b.Column)+1)
		// ID is recommended by golint here
		s := newTagStore()
		s.findValueByName(v, append(b.Column, "id"), found, false)

		value := found[:len(found)-1]
		for i, v := range value {
			if v != nil {
				value[i] = v.(reflect.Value).Interface()
			}
		}

		if v.CanSet() {
			switch idField := found[len(found)-1].(type) {
			case reflect.Value:
				if idField.Kind() == reflect.Int64 {
					b.RecordID = idField.Addr().Interface().(*int64)
				}
			}
		}
		b.Values(value...)
	}
	return b
}

// Returning specifies the returning columns for postgres/mssql.
func (b *InsertStmt) Returning(column ...string) *InsertStmt {
	b.ReturnColumn = column
	return b
}

// Pair adds (column, value) to be inserted.
// It is an error to mix Pair with Values and Record.
func (b *InsertStmt) Pair(column string, value interface{}) *InsertStmt {
	b.Column = append(b.Column, column)
	switch len(b.Value) {
	case 0:
		b.Values(value)
	case 1:
		b.Value[0] = append(b.Value[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}
