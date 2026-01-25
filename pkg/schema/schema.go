package schema

import (
	ddl2 "github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml"
	exec2 "github.com/yaroher/ratel/pkg/exec"
	"github.com/yaroher/ratel/pkg/types"
)

type Table[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C]] struct {
	*ddl2.TableDDL[T, C]
	*dml.TableDML[T, C]
	*exec2.TableExecutor[T, C, S]
	constructor func() S
}

func NewTable[T types.TableAlias, C types.ColumnAlias, S exec2.Scanner[C]](
	alias T,
	constructor func() S,
	columns []*ddl2.ColumnDDL[C],
	ddlOptions ...ddl2.TableOptions[T, C],
) *Table[T, C, S] {
	allAliases := make([]C, 0, len(columns))
	for _, col := range columns {
		allAliases = append(allAliases, col.Alias())
	}
	return &Table[T, C, S]{
		TableDDL:      ddl2.NewTableDDL[T, C](alias, columns, ddlOptions...),
		TableDML:      dml.NewTableDML[T, C](alias, allAliases...),
		TableExecutor: exec2.NewTableExecutor[T, C, S](alias, allAliases, constructor),
		constructor:   constructor,
	}
}
