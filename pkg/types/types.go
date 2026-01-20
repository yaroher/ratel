package types

import (
	"fmt"
	"strings"
)

/*
	██████╗  █████╗ ████████╗███████╗██╗
	██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██║
	██████╔╝███████║   ██║   █████╗  ██║
	██╔══██╗██╔══██║   ██║   ██╔══╝  ██║
	██║  ██║██║  ██║   ██║   ███████╗███████╗
	╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝
*/

type ColumnAlias = fmt.Stringer

type BufferBuilder interface {
	AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any)
}

type SqlBuilder interface {
	Build() (string, []any)
}

type Clause[C ColumnAlias] interface {
	BufferBuilder
}

type ValueSetter[C ColumnAlias] interface {
	BufferBuilder
	Value() any
	Column() C
}

type OrmQuery interface {
	BufferBuilder
	ScanAbleFields() []string
	TableAlias() string
	Build() (string, []any)
}
