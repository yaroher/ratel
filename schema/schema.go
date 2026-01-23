package schema

import (
	"github.com/yaroher/ratel/common/types"
	"github.com/yaroher/ratel/ddl"
	"github.com/yaroher/ratel/dml"
	"github.com/yaroher/ratel/exec"
)

type Table[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C]] struct {
	*ddl.TableDDL[T, C]
	*dml.TableDML[T, C]
	*exec.TableExecutor[T, C, S]
}

func NewTable[T types.TableAlias, C types.ColumnAlias, S exec.Scanner[C]](
	alias T,
	constructor func() S,
	columns ...*ddl.ColumnDDL[C],
) *Table[T, C, S] {
	allAliases := make([]C, 0, len(columns))
	for _, col := range columns {
		allAliases = append(allAliases, col.Alias())
	}
	return &Table[T, C, S]{
		TableDDL:      ddl.NewTableDDL[T, C](alias, columns...),
		TableDML:      dml.NewTableDML[T, C](alias, allAliases...),
		TableExecutor: exec.NewTableExecutor[T, C, S](alias, allAliases, constructor),
	}
}
