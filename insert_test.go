package tyr

import (
	"testing"

	"github.com/kubuskotak/tyr/dialect"
	"github.com/stretchr/testify/require"
)

type insertTest struct {
	A int
	C string `sql:"b"`
}

func TestInsertStmt(t *testing.T) {
	buf := NewBuffer()
	builder := InsertInto("table").Ignore().Columns("a", "b").Values(1, "one").Record(&insertTest{
		A: 2,
		C: "two",
	}).Comment("INSERT TEST")
	err := builder.ToSQL(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "/* INSERT TEST */\nINSERT IGNORE INTO `table` (`a`,`b`) VALUES (?,?), (?,?)", buf.String())
	require.Equal(t, []interface{}{1, "one", 2, "two"}, buf.Value())
}

func BenchmarkInsertValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b").Values(1, "one").ToSQL(dialect.MySQL, buf)
	}
}

func BenchmarkInsertRecordSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b").Record(&insertTest{
			A: 2,
			C: "two",
		}).ToSQL(dialect.MySQL, buf)
	}
}
