package tyr

import (
	"strconv"
)

// DeleteStmt builds `DELETE ...`.
type DeleteStmt struct {
	Dialect

	raw

	Table      string
	WhereCond  []Builder
	LimitCount int64

	comments Comments
}

type DeleteBuilder = DeleteStmt

func (b *DeleteStmt) ToSQL(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.ToSQL(d, buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	err := b.comments.ToSQL(d, buf)
	if err != nil {
		return err
	}

	buf.WriteString("DELETE FROM ")
	buf.WriteString(d.QuoteIdent(b.Table))

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).ToSQL(d, buf)
		if err != nil {
			return err
		}
	}
	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
	}
	return nil
}

// DeleteFrom creates a DeleteStmt.
func DeleteFrom(table string) *DeleteStmt {
	return &DeleteStmt{
		Table:      table,
		LimitCount: -1,
	}
}

// DeleteBySql creates a DeleteStmt from raw query.
func DeleteBySql(query string, value ...interface{}) *DeleteStmt {
	return &DeleteStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
		LimitCount: -1,
	}
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *DeleteStmt) Where(query interface{}, value ...interface{}) *DeleteStmt {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

func (b *DeleteStmt) Limit(n uint64) *DeleteStmt {
	b.LimitCount = int64(n)
	return b
}

func (b *DeleteStmt) Comment(comment string) *DeleteStmt {
	b.comments = b.comments.Append(comment)
	return b
}