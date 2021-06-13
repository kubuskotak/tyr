package tyr

import "strconv"

// UpdateStmt builds `UPDATE ...`.
type UpdateStmt struct {
	Dialect

	raw

	Table        string
	Value        map[string]interface{}
	WhereCond    []Builder
	ReturnColumn []string
	LimitCount   int64
	comments     Comments
}

type UpdateBuilder = UpdateStmt

func (b *UpdateStmt) ToSql(d Dialect, buf Buffer) error {
	i := interpolator{
		Buffer:       buf,
		Dialect:      d,
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(b, true)
	if err != nil {
		return err
	}
	return nil
}

func (b *UpdateStmt) Build(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(d, buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Value) == 0 {
		return ErrColumnNotSpecified
	}

	err := b.comments.Build(d, buf)
	if err != nil {
		return err
	}

	buf.WriteString("UPDATE ")
	buf.WriteString(d.QuoteIdent(b.Table))
	buf.WriteString(" SET ")

	i := 0
	for col, v := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(d.QuoteIdent(col))
		buf.WriteString(" = ")
		buf.WriteString(placeholder)

		buf.WriteValue(v)

		i++
	}

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}

	if len(b.ReturnColumn) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.ReturnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(d.QuoteIdent(col))
		}
	}

	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
	}

	return nil
}

// Update creates an UpdateStmt.
func Update(table string) *UpdateStmt {
	return &UpdateStmt{
		Table:      table,
		Value:      make(map[string]interface{}),
		LimitCount: -1,
	}
}

// UpdateBySql creates an UpdateStmt with raw query.
func UpdateBySql(query string, value ...interface{}) *UpdateStmt {
	return &UpdateStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
		Value:      make(map[string]interface{}),
		LimitCount: -1,
	}
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *UpdateStmt) Where(query interface{}, value ...interface{}) *UpdateStmt {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Returning specifies the returning columns for postgres.
func (b *UpdateStmt) Returning(column ...string) *UpdateStmt {
	b.ReturnColumn = column
	return b
}

// Set updates column with value.
func (b *UpdateStmt) Set(column string, value interface{}) *UpdateStmt {
	b.Value[column] = value
	return b
}

// SetMap specifies a map of (column, value) to update in bulk.
func (b *UpdateStmt) SetMap(m map[string]interface{}) *UpdateStmt {
	for col, val := range m {
		b.Set(col, val)
	}
	return b
}

// IncrBy increases column by value
func (b *UpdateStmt) IncrBy(column string, value interface{}) *UpdateStmt {
	b.Value[column] = Expr("? + ?", I(column), value)
	return b
}

// DecrBy decreases column by value
func (b *UpdateStmt) DecrBy(column string, value interface{}) *UpdateStmt {
	b.Value[column] = Expr("? - ?", I(column), value)
	return b
}

func (b *UpdateStmt) Limit(n uint64) *UpdateStmt {
	b.LimitCount = int64(n)
	return b
}

func (b *UpdateStmt) Comment(comment string) *UpdateStmt {
	b.comments = b.comments.Append(comment)
	return b
}
