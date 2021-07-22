package tyr

type union struct {
	builder []Builder
	all     bool
}

// Union builds `... UNION ...`.
func Union(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
	}
}

// UnionAll builds `... UNION ALL ...`.
func UnionAll(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
		all:     true,
	}
}

func (u *union) ToSQL(d Dialect, buf Buffer) error {
	return u.Build(d, buf)
}

func (u *union) Build(d Dialect, buf Buffer) error {
	for i, b := range u.builder {
		if i > 0 {
			_, _ = buf.WriteString(" UNION ")
			if u.all {
				_, _ = buf.WriteString("ALL ")
			}
		}
		err := b.Build(d, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *union) As(alias string) Builder {
	return as(u, alias)
}
