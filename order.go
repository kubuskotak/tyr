package tyr

type direction bool

// order by directions
// most databases by default use asc
const (
	asc  direction = false
	desc           = true
)

func order(column string, dir direction) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		// FIXME: no quote ident
		buf.WriteString(column)
		switch dir {
		case asc:
			buf.WriteString(" ASC")
		case desc:
			buf.WriteString(" DESC")
		}
		return nil
	})
}
