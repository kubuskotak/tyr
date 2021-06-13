package tyr

import (
	"testing"

	"github.com/kubuskotak/tyr/dialect"
	"github.com/stretchr/testify/require"
)

func TestUpdateStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Update("table").Set("a", 1).Where(Eq("b", 2)).Comment("UPDATE TEST")
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)

	require.Equal(t, "/* UPDATE TEST */\nUPDATE `table` SET `a` = ? WHERE (`b` = ?)", buf.String())
	require.Equal(t, []interface{}{1, 2}, buf.Value())
}

func BenchmarkUpdateValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").Set("a", 1).Build(dialect.MySQL, buf)
	}
}

func BenchmarkUpdateMapSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").SetMap(map[string]interface{}{"a": 1, "b": 2}).Build(dialect.MySQL, buf)
	}
}

func TestUpdateIncrBy(t *testing.T) {
	buf := NewBuffer()
	builder := Update("table").IncrBy("a", 1).Where(Eq("b", 2))
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)

	sqlstr, err := InterpolateForDialect(buf.String(), buf.Value(), dialect.MySQL)
	require.NoError(t, err)

	require.Equal(t, "UPDATE `table` SET `a` = 'a' + 1 WHERE (`b` = 2)", sqlstr)
}
