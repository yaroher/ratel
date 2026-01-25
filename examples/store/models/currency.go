package models

import (
	"github.com/yaroher/ratel/pkg/ddl"
	"github.com/yaroher/ratel/pkg/dml/set"
	"github.com/yaroher/ratel/pkg/schema"
)

// CurrencyAlias is the table alias type for the currency table
type CurrencyAlias string

func (c CurrencyAlias) String() string { return string(c) }

const CurrencyAliasName CurrencyAlias = "currency"

// CurrencyColumnAlias represents column names for the currency table
type CurrencyColumnAlias string

func (c CurrencyColumnAlias) String() string { return string(c) }

const (
	CurrencyColumnCode CurrencyColumnAlias = "code"
	CurrencyColumnName CurrencyColumnAlias = "name"
)

// CurrencyScanner is the scanner struct for currency rows
type CurrencyScanner struct {
	Code string
	Name string
}

func (c *CurrencyScanner) GetTarget(s string) func() any {
	switch CurrencyColumnAlias(s) {
	case CurrencyColumnCode:
		return func() any { return &c.Code }
	case CurrencyColumnName:
		return func() any { return &c.Name }
	default:
		panic("unknown field: " + s)
	}
}

func (c *CurrencyScanner) GetSetter(f CurrencyColumnAlias) func() set.ValueSetter[CurrencyColumnAlias] {
	switch f {
	case CurrencyColumnCode:
		return func() set.ValueSetter[CurrencyColumnAlias] { return set.NewSetter(f, &c.Code) }
	case CurrencyColumnName:
		return func() set.ValueSetter[CurrencyColumnAlias] { return set.NewSetter(f, &c.Name) }
	default:
		panic("unknown field: " + string(f))
	}
}

func (c *CurrencyScanner) GetValue(f CurrencyColumnAlias) func() any {
	switch f {
	case CurrencyColumnCode:
		return func() any { return c.Code }
	case CurrencyColumnName:
		return func() any { return c.Name }
	default:
		panic("unknown field: " + string(f))
	}
}

// CurrencyTable represents the currency table with its columns
type CurrencyTable struct {
	*schema.Table[CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner]
	Code schema.CharColumnI[CurrencyColumnAlias]
	Name schema.TextColumnI[CurrencyColumnAlias]
}

// Currency is the global currency table instance
var Currency = func() CurrencyTable {
	codeCol := schema.CharColumn(CurrencyColumnCode, 3, ddl.WithPrimaryKey[CurrencyColumnAlias]())
	nameCol := schema.TextColumn(CurrencyColumnName, ddl.WithNotNull[CurrencyColumnAlias]())

	return CurrencyTable{
		Table: schema.NewTable[CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner](
			CurrencyAliasName,
			func() *CurrencyScanner { return &CurrencyScanner{} },
			[]*ddl.ColumnDDL[CurrencyColumnAlias]{
				codeCol.DDL(),
				nameCol.DDL(),
			},
		),
		Code: codeCol,
		Name: nameCol,
	}
}()

// CurrencyRef is a reference to the currency table for relations
var CurrencyRef schema.RelationTableAlias[CurrencyAlias] = Currency.Table
