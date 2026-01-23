package types

import (
	"fmt"
	"strings"
)

type ColumnAlias fmt.Stringer

func Alias[C columnAlias](s string) C {
	return C(columnAlias(s))
}

type columnAlias string

func (a columnAlias) String() string {
	return string(a)
}

type TableAlias fmt.Stringer

type Builder interface {
	AddToBuilder(buf *strings.Builder, ta string, paramIndex *int, args *[]any)
}

type Buildable interface {
	Build() (string, []any)
}

type Scannable interface {
	Buildable
	ScanAbleFields() []string
}

type Query interface {
	Buildable
	Builder
	TableAlias() string
}
