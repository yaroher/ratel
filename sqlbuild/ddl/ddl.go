package ddl

import (
	"fmt"
	"strings"
	"sync"

	"github.com/yaroher/ratel/common/types"
)

type Datatype fmt.Stringer

type DDL interface {
	types.Buildable
	ddlQuery()
}

var sbPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}
