package tyr

type joinType uint8

const (
	inner joinType = iota
	left
	right
	full
)

func join(t joinType, table, on interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		_, _ = buf.WriteString(" ")
		switch t {
		case left:
			_, _ = buf.WriteString("LEFT ")
		case right:
			_, _ = buf.WriteString("RIGHT ")
		case full:
			_, _ = buf.WriteString("FULL ")
		}
		_, _ = buf.WriteString("JOIN ")
		switch table := table.(type) {
		case string:
			_, _ = buf.WriteString(d.QuoteIdent(table))
		default:
			_, _ = buf.WriteString(placeholder)
			_ = buf.WriteValue(table)
		}
		_, _ = buf.WriteString(" ON ")
		switch on := on.(type) {
		case string:
			_, _ = buf.WriteString(on)
		case Builder:
			_, _ = buf.WriteString(placeholder)
			_ = buf.WriteValue(on)
		}
		return nil
	})
}
