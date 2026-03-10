package migrate

// SchemaState represents the full state of a database realm.
type SchemaState struct {
	Schemas []Schema
}

type Schema struct {
	Name       string
	Tables     []Table
	Extensions []Extension
	Functions  []Function
	Triggers   []Trigger // schema-level triggers
}

type Table struct {
	Name        string
	Schema      string
	Columns     []Column
	PrimaryKey  *Index
	Indexes     []Index
	ForeignKeys []ForeignKey
	Checks      []Check
	Policies    []Policy
	Triggers    []Trigger
	RLSEnabled  bool
	RLSForced   bool
	Comment     string
}

type Column struct {
	Name     string
	Type     string
	Nullable bool
	Default  string // raw SQL default expression
	Identity *Identity
	Comment  string
}

type Identity struct {
	Generation string // "ALWAYS" or "BY DEFAULT"
	Start      int64
	Increment  int64
}

type Index struct {
	Name          string
	Columns       []IndexColumn
	Unique        bool
	Method        string // btree, hash, gin, gist, etc.
	Where         string // partial index predicate
	Include       []string
	NullsDistinct *bool
	Comment       string
}

type IndexColumn struct {
	Name  string
	Desc  bool
	Order string // ASC, DESC, or empty
}

type ForeignKey struct {
	Name       string
	Columns    []string
	RefTable   string
	RefSchema  string
	RefColumns []string
	OnUpdate   string
	OnDelete   string
}

type Check struct {
	Name      string
	Expr      string
	NoInherit bool
}

type Policy struct {
	Name       string
	Permissive bool
	Command    string
	Roles      []string
	Using      string
	WithCheck  string
}

type Trigger struct {
	Name       string
	Table      string
	Events     []string
	Timing     string
	ForEachRow bool
	When       string
	Function   string
	Args       []string
	Comment    string
}

type Extension struct {
	Name    string
	Schema  string
	Version string
}

type Function struct {
	Name       string
	Schema     string
	Args       []FunctionArg
	ReturnType string
	Language   string
	Body       string
	Volatility string
	Security   string
	Comment    string
}

type FunctionArg struct {
	Name    string
	Type    string
	Mode    string
	Default string
}
