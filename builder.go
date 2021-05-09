package tyr

// Builder builds SQL in Dialect like MySQL, and PostgreSQL.
// The raw SQL and values are stored in Buffer.
//
// The core of tyr is interpolation, which can expand ? with arbitrary SQL.
// If you need a feature that is not currently supported, you can build it
// on your own (or use Expr).
//
// To do that, the value that you wish to be expanded with ? needs to
// implement Builder.
type Builder interface {
	ToSQL(Dialect, Buffer) error
}

// BuildFunc implements Builder.
type BuildFunc func(Dialect, Buffer) error

// ToSQL calls itself to build SQL.
func (b BuildFunc) ToSQL(d Dialect, buf Buffer) error {
	return b(d, buf)
}
