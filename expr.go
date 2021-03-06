package tyr

type raw struct {
	Query string
	Value []interface{}
}

// Expr allows raw expression to be used when current SQL syntax is
// not supported by gocraft/dbr.
func Expr(query string, value ...interface{}) Builder {
	return &raw{Query: query, Value: value}
}

func (raw *raw) ToSQL(_ Dialect, buf Buffer) error {
	_, _ = buf.WriteString(raw.Query)
	_ = buf.WriteValue(raw.Value...)
	return nil
}

func (raw *raw) Build(_ Dialect, buf Buffer) error {
	_, _ = buf.WriteString(raw.Query)
	_ = buf.WriteValue(raw.Value...)
	return nil
}
