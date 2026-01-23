package dml

import (
	"strings"
	"sync"

	"github.com/yaroher/ratel/common/types"
)

var sbPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

type BaseQuery[T types.TableAlias, C types.ColumnAlias] struct {
	Ta          T
	UsingFields []C
	AllFields   []C
}

func (q *BaseQuery[T, C]) ScanAbleFields() []string {
	mp := make([]string, len(q.UsingFields))
	for i, f := range q.UsingFields {
		mp[i] = f.String()
	}
	return mp
}
func (q *BaseQuery[T, C]) buildReturning(sb *strings.Builder) {
	// Не добавляем RETURNING, если нет полей для возврата
	if len(q.UsingFields) == 0 {
		return
	}

	sb.WriteString(" RETURNING ")
	for i, field := range q.UsingFields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field.String())
	}
}
