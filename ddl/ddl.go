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

// datatype is a simple implementation of Datatype
type datatype string

func (d datatype) String() string { return string(d) }

// Common PostgreSQL datatypes
func SmallInt() Datatype     { return datatype("SMALLINT") }
func Integer() Datatype      { return datatype("INTEGER") }
func BigInt() Datatype       { return datatype("BIGINT") }
func Real() Datatype         { return datatype("REAL") }
func Double() Datatype       { return datatype("DOUBLE PRECISION") }
func Numeric() Datatype      { return datatype("NUMERIC") }
func Serial() Datatype       { return datatype("SERIAL") }
func BigSerial() Datatype    { return datatype("BIGSERIAL") }
func Text() Datatype         { return datatype("TEXT") }
func Varchar(n int) Datatype { return datatype(fmt.Sprintf("VARCHAR(%d)", n)) }
func Char(n int) Datatype    { return datatype(fmt.Sprintf("CHAR(%d)", n)) }
func Boolean() Datatype      { return datatype("BOOLEAN") }
func Date() Datatype         { return datatype("DATE") }
func Time() Datatype         { return datatype("TIME") }
func Timestamp() Datatype    { return datatype("TIMESTAMP") }
func Timestamptz() Datatype  { return datatype("TIMESTAMPTZ") }
func Interval() Datatype     { return datatype("INTERVAL") }
func UUID() Datatype         { return datatype("UUID") }
func JSON() Datatype         { return datatype("JSON") }
func JSONB() Datatype        { return datatype("JSONB") }
func Bytea() Datatype        { return datatype("BYTEA") }
