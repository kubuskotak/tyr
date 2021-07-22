package tyr

import (
	"testing"

	"github.com/kubuskotak/tyr/dialect"
	"github.com/stretchr/testify/require"
)

func TestSelectStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Select("a", "b").
		From(Select("a").From("table")).
		LeftJoin("table2", "table.a1 = table.a2").
		Distinct().
		Where(Eq("c", 1)).
		GroupBy("d").
		Having(Eq("e", 2)).
		OrderAsc("f").
		Limit(3).
		Offset(4).
		Suffix("FOR UPDATE").
		Comment("SELECT TEST")

	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "/* SELECT TEST */\nSELECT DISTINCT a, b FROM ? LEFT JOIN `table2` ON table.a1 = table.a2 WHERE (`c` = ?) GROUP BY d HAVING (`e` = ?) ORDER BY f ASC LIMIT 3 OFFSET 4 FOR UPDATE", buf.String())
	// two functions cannot be compared
	require.Equal(t, 3, len(buf.Value()))
}

func BenchmarkSelectSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		_ = Select("a", "b").From("table").Where(Eq("c", 1)).OrderAsc("d").Build(dialect.MySQL, buf)
	}
}
