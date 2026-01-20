package query

import (
	"strings"
	"sync"

	"github.com/yaroher/ratel/pkg/types"
)

var sbPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

type BaseQuery[C types.ColumnAlias] struct {
	Ta          string
	UsingFields []C
	AllFields   []C
}

func (q *BaseQuery[C]) TableAlias() string { return q.Ta }
func (q *BaseQuery[C]) ScanAbleFields() []string {
	mp := make([]string, len(q.UsingFields))
	for i, f := range q.UsingFields {
		mp[i] = f.String()
	}
	return mp
}
func (q *BaseQuery[С]) buildReturning(sb *strings.Builder) {
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
