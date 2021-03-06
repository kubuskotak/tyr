package tyr

import (
	"fmt"
	"strconv"

	"github.com/kubuskotak/tyr/dialect"
)

// SelectStmt builds `SELECT ...`.
type SelectStmt struct {
	Dialect

	raw

	IsDistinct bool

	Column    []interface{}
	Table     interface{}
	JoinTable []Builder

	WhereCond  []Builder
	Group      []Builder
	HavingCond []Builder
	Order      []Builder
	Suffixes   []Builder

	LimitCount  int64
	OffsetCount int64

	comments Comments
}

func (b *SelectStmt) ToSQL(d Dialect, i Buffer) error {
	builder := NewBuffer()
	_ = b.Build(d, builder)
	return interpolateSql(d, i, builder.String(), builder.Value())
}

func (b *SelectStmt) Build(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(d, buf)
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	err := b.comments.Build(d, buf)
	if err != nil {
		return err
	}

	_, _ = buf.WriteString("SELECT ")

	if b.IsDistinct {
		_, _ = buf.WriteString("DISTINCT ")
	}

	for i, col := range b.Column {
		if i > 0 {
			_, _ = buf.WriteString(", ")
		}
		switch col := col.(type) {
		case string:
			// FIXME: no quote ident
			_, _ = buf.WriteString(col)
		default:
			_, _ = buf.WriteString(placeholder)
			_ = buf.WriteValue(col)
		}
	}

	if b.Table != nil {
		_, _ = buf.WriteString(" FROM ")
		switch table := b.Table.(type) {
		case string:
			// FIXME: no quote ident
			_, _ = buf.WriteString(table)
		default:
			_, _ = buf.WriteString(placeholder)
			_ = buf.WriteValue(table)
		}
		if len(b.JoinTable) > 0 {
			for _, join := range b.JoinTable {
				err := join.Build(d, buf)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(b.WhereCond) > 0 {
		_, _ = buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}

	if len(b.Group) > 0 {
		_, _ = buf.WriteString(" GROUP BY ")
		for i, group := range b.Group {
			if i > 0 {
				_, _ = buf.WriteString(", ")
			}
			err := group.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	if len(b.HavingCond) > 0 {
		_, _ = buf.WriteString(" HAVING ")
		err := And(b.HavingCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}

	if len(b.Order) > 0 {
		_, _ = buf.WriteString(" ORDER BY ")
		for i, order := range b.Order {
			if i > 0 {
				_, _ = buf.WriteString(", ")
			}
			err := order.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	if d == dialect.MSSQL {
		b.addMSSQLLimits(buf)
	} else {
		if b.LimitCount >= 0 {
			_, _ = buf.WriteString(" LIMIT ")
			_, _ = buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
		}

		if b.OffsetCount >= 0 {
			_, _ = buf.WriteString(" OFFSET ")
			_, _ = buf.WriteString(strconv.FormatInt(b.OffsetCount, 10))
		}
	}

	if len(b.Suffixes) > 0 {
		for _, suffix := range b.Suffixes {
			_, _ = buf.WriteString(" ")
			err := suffix.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// https://docs.microsoft.com/en-us/previous-versions/sql/sql-server-2012/ms188385(v=sql.110)
func (b *SelectStmt) addMSSQLLimits(buf Buffer) {
	limitCount := b.LimitCount
	offsetCount := b.OffsetCount
	if limitCount < 0 && offsetCount < 0 {
		return
	}
	if offsetCount < 0 {
		offsetCount = 0
	}

	if len(b.Order) == 0 {
		// ORDER is required for OFFSET / FETCH
		_, _ = buf.WriteString(" ORDER BY ")
		col := b.Column[0]
		switch col := col.(type) {
		case string:
			// FIXME: no quote ident
			_, _ = buf.WriteString(col)
		default:
			_, _ = buf.WriteString(placeholder)
			_ = buf.WriteValue(col)
		}
	}

	_, _ = buf.WriteString(" OFFSET ")
	_, _ = buf.WriteString(strconv.FormatInt(offsetCount, 10))
	_, _ = buf.WriteString(" ROWS ")

	if limitCount >= 0 {
		_, _ = buf.WriteString(" FETCH FIRST ")
		_, _ = buf.WriteString(strconv.FormatInt(limitCount, 10))
		_, _ = buf.WriteString(" ROWS ONLY ")
	}
}

// Select creates a SelectStmt.
func Select(column ...interface{}) *SelectStmt {
	return &SelectStmt{
		Column:      column,
		LimitCount:  -1,
		OffsetCount: -1,
	}
}

// SelectBySql creates a SelectStmt from raw query.
func SelectBySql(query string, value ...interface{}) *SelectStmt {
	return &SelectStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
		LimitCount:  -1,
		OffsetCount: -1,
	}
}

// From specifies table to select from.
// table can be Builder like SelectStmt, or string.
func (b *SelectStmt) From(table interface{}) *SelectStmt {
	b.Table = table
	return b
}

func (b *SelectStmt) Distinct() *SelectStmt {
	b.IsDistinct = true
	return b
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectStmt) Where(query interface{}, value ...interface{}) *SelectStmt {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Having adds a having condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectStmt) Having(query interface{}, value ...interface{}) *SelectStmt {
	switch query := query.(type) {
	case string:
		b.HavingCond = append(b.HavingCond, Expr(query, value...))
	case Builder:
		b.HavingCond = append(b.HavingCond, query)
	}
	return b
}

// GroupBy specifies columns for grouping.
func (b *SelectStmt) GroupBy(col ...string) *SelectStmt {
	for _, group := range col {
		b.Group = append(b.Group, Expr(group))
	}
	return b
}

func (b *SelectStmt) OrderAsc(col string) *SelectStmt {
	b.Order = append(b.Order, order(col, asc))
	return b
}

func (b *SelectStmt) OrderDesc(col string) *SelectStmt {
	b.Order = append(b.Order, order(col, desc))
	return b
}

// OrderBy specifies columns for ordering.
func (b *SelectStmt) OrderBy(col string) *SelectStmt {
	b.Order = append(b.Order, Expr(col))
	return b
}

func (b *SelectStmt) Limit(n uint64) *SelectStmt {
	b.LimitCount = int64(n)
	return b
}

func (b *SelectStmt) Offset(n uint64) *SelectStmt {
	b.OffsetCount = int64(n)
	return b
}

// Suffix adds an expression to the end of the query. This is useful to add dialect-specific clauses like FOR UPDATE
func (b *SelectStmt) Suffix(suffix string, value ...interface{}) *SelectStmt {
	b.Suffixes = append(b.Suffixes, Expr(suffix, value...))
	return b
}

// Seek fetches a page in key set way for a large set of data.
func (b *SelectStmt) Seek(col string, cursor, limit uint64) *SelectStmt {
	if limit == 0 { // default limit
		limit = 10
	}
	nextCursor := cursor + limit
	b.Limit(limit)
	b.WhereCond = append(b.WhereCond, Expr(fmt.Sprintf("%s > ? AND %s <= ?", col, col), cursor, nextCursor))
	return b
}

// Paginate fetches a page in a naive way for a small set of data.
func (b *SelectStmt) Paginate(page, perPage uint64) *SelectStmt {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// OrderDir is a helper for OrderAsc and OrderDesc.
func (b *SelectStmt) OrderDir(col string, isAsc bool) *SelectStmt {
	if isAsc {
		b.OrderAsc(col)
	} else {
		b.OrderDesc(col)
	}
	return b
}

func (b *SelectStmt) Comment(comment string) *SelectStmt {
	b.comments = b.comments.Append(comment)
	return b
}

// Join add inner-join.
// on can be Builder or string.
func (b *SelectStmt) Join(table, on interface{}) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(inner, table, on))
	return b
}

// LeftJoin add left-join.
// on can be Builder or string.
func (b *SelectStmt) LeftJoin(table, on interface{}) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(left, table, on))
	return b
}

// RightJoin add right-join.
// on can be Builder or string.
func (b *SelectStmt) RightJoin(table, on interface{}) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(right, table, on))
	return b
}

// FullJoin add full-join.
// on can be Builder or string.
func (b *SelectStmt) FullJoin(table, on interface{}) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(full, table, on))
	return b
}

// As creates alias for select statement.
func (b *SelectStmt) As(alias string) Builder {
	return as(b, alias)
}
