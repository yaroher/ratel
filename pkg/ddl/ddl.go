package ddl

import (
	"fmt"
)

type Datatype fmt.Stringer

// datatype is a simple implementation of Datatype
type datatype string

func (d datatype) String() string { return string(d) }

const (
	SMALLINT    datatype = "SMALLINT"
	INTEGER     datatype = "INTEGER"
	BIGINT      datatype = "BIGINT"
	SERIAL      datatype = "SERIAL"
	BIGSERIAL   datatype = "BIGSERIAL"
	REAL        datatype = "REAL"
	DOUBLE      datatype = "DOUBLE PRECISION"
	TEXT        datatype = "TEXT"
	BOOLEAN     datatype = "BOOLEAN"
	DATE        datatype = "DATE"
	TIME        datatype = "TIME"
	TIMESTAMP   datatype = "TIMESTAMP"
	TIMESTAMPTZ datatype = "TIMESTAMPTZ"
	INTERVAL    datatype = "INTERVAL"
	UUID        datatype = "UUID"
	JSON        datatype = "JSON"
	JSONB       datatype = "JSONB"
	BYTEA       datatype = "BYTEA"
)

func Char(length int) Datatype { return datatype(fmt.Sprintf("CHAR(%d)", length)) }

func Varchar(length int) Datatype { return datatype(fmt.Sprintf("VARCHAR(%d)", length)) }

func Numeric(precision, scale int) Datatype {
	return datatype(fmt.Sprintf("NUMERIC(%d, %d)", precision, scale))
}

func Array(d Datatype) Datatype { return datatype(fmt.Sprintf("ARRAY[%s]", d)) }
