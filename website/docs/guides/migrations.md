---
sidebar_position: 9
title: Migrations
---

# Migrations

Ratel provides two migration engines for schema diffing and migration generation.

## Why Two Engines?

[Atlas](https://atlasgo.io/) OSS handles standard DDL — tables, columns, indexes, foreign keys, and check constraints. However, advanced PostgreSQL features like **Row Level Security**, **triggers**, **functions**, and **extensions** require Atlas Pro (paid license).

The built-in **Ratel engine** (`--engine ratel`) unlocks these features for free by using its own PostgreSQL inspector, differ, and planner. Choose the engine that fits your needs:

| Feature | Atlas (default) | Ratel |
|---------|:-:|:-:|
| Tables, columns, indexes | ✅ | ✅ |
| Foreign keys, checks | ✅ | ✅ |
| Row Level Security | ❌ (Pro) | ✅ |
| Policies | ❌ (Pro) | ✅ |
| Triggers | ❌ (Pro) | ✅ |
| Functions | ❌ (Pro) | ✅ |
| Extensions | ❌ (Pro) | ✅ |

## Generate Migration Diff

Compare your current Go models against an existing schema and generate a migration:

```bash
# Atlas engine (default) — basic DDL
ratel diff \
  -p github.com/myapp/models \
  --discover \
  -d ./migrations \
  -n add_user_profiles

# Ratel engine — full PostgreSQL support
ratel diff \
  -p github.com/myapp/models \
  --discover \
  --engine ratel \
  -d ./migrations \
  -n add_rls_policies
```

This:
1. Generates the desired schema from your Go models
2. Compares it against the current database state
3. Produces a migration file in `./migrations/`

## Multiple Packages

When your models are split across packages, pass them all at once to avoid migrations that drop tables from other packages:

```bash
ratel diff \
  -p github.com/myapp/auth,github.com/myapp/store \
  --discover \
  -d ./migrations \
  -n init
```

All tables from all packages are collected into a single schema before diffing.

## From SQL File

You can also diff against a SQL schema file:

```bash
ratel diff \
  -s schema.sql \
  -d ./migrations \
  -n initial
```

## Row Level Security

Declare RLS policies via `post_statements` in proto annotations:

```protobuf
message User {
  option (ratel.table) = {
    generate: true
    table_name: "users"
    post_statements: [
      "ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
      "CREATE POLICY users_own_data ON {table} FOR ALL USING (id = current_setting('app.current_user_id')::bigint)"
    ]
  };
  // ...
}
```

Or in Go models with `WithPostStatements`:

```go
schema.NewTable[UsersAlias, UsersColumnAlias, *UsersScanner](
    UsersAliasName,
    factory,
    columns,
    ddl.WithPostStatements[UsersAlias, UsersColumnAlias](
        "ALTER TABLE {table} ENABLE ROW LEVEL SECURITY",
        "CREATE POLICY users_own_data ON {table} FOR ALL USING (user_id = current_setting('app.current_user_id')::bigint)",
    ),
)
```

The `{table}` placeholder is replaced with the schema-qualified table name.

Then generate migrations with the ratel engine:

```bash
ratel diff --engine ratel -p github.com/myapp/models --discover -d ./migrations -n add_rls
```

## File-Level SQL (Extensions, Functions)

For SQL that isn't tied to a specific table (extensions, standalone functions), use the `additional_code` file-level proto option:

```protobuf
import "ratelproto/ratelproto.proto";

option (ratel.additional_code) = "CREATE EXTENSION IF NOT EXISTS pg_trgm";
option (ratel.additional_code) = "CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN NEW.updated_at = now(); RETURN NEW; END; $$";
```

This generates a `var AdditionalSQL` in Go code that can be passed to `ddl.SchemaSQL()`.

In Go code (without proto), use `ddl.RawSQL`:

```go
ddl.SchemaSQL(
    ddl.RawSQL("CREATE EXTENSION IF NOT EXISTS pg_trgm"),
    models.Users,
    models.Orders,
)
```

## Recalculating Checksums

After manually editing or adding `.sql` files in the migrations directory, recalculate `atlas.sum`:

```bash
# Atlas engine
ratel migrate hash -d ./migrations

# Ratel engine
ratel migrate hash -d ./migrations -e ratel
```

Both engines use the `atlas.sum` format for compatibility with Atlas tooling.

## CLI Options

```
ratel diff [flags]

Flags:
  -s, --sql string         SQL schema file to compare against
  -p, --package strings    Go package path(s) containing models (comma-separated)
  -t, --tables strings     Table variable names to include
  -d, --discover           Auto-discover tables from source files
  --dir string             Migration directory (default: ./migrations)
  -n, --name string        Migration name (default: migration)
  --pg_version int         PostgreSQL version (default: 18)
  -e, --engine string      Migration engine: atlas or ratel (default: "atlas")
```

```
ratel migrate hash [flags]

Flags:
  -d, --dir string      Migration directory (default: ./migrations)
  -e, --engine string   Migration engine: atlas or ratel (default: "atlas")
```

## Workflow

1. Modify your proto files or Go models
2. Regenerate code (if using proto)
3. Run `ratel diff` to generate migration
4. Review the generated migration
5. Apply it to your database

```bash
# Regenerate from proto
protoc --go_out=. --ratel_out=. schema.proto

# Generate migration with RLS support
ratel diff -p myapp/models --discover --engine ratel -d ./migrations -n add_rls

# Review
cat migrations/*_add_rls.sql

# Apply (using your preferred tool)
psql -f migrations/*_add_rls.sql
```
