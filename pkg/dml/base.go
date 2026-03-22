package dml

import (
	"strings"
	"sync"

	"github.com/yaroher/ratel/pkg/types"
)

var sbPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

type BaseQuery[T types.TableAlias, C types.ColumnAlias] struct {
	Ta          T
	FromName    string // schema-qualified table name for FROM (empty = use Ta.String())
	UsingFields []C
	AllFields   []C
}

// fromName returns the table name for FROM clauses (qualified if schema is set).
func (q *BaseQuery[T, C]) fromName() string {
	if q.FromName != "" {
		return q.FromName
	}
	return q.Ta.String()
}

func (q *BaseQuery[T, C]) TableAlias() string {
	return q.Ta.String()
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
