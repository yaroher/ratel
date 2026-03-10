---
sidebar_position: 9
title: Migrations
---

# Migrations

Ratel integrates with [Atlas](https://atlasgo.io/) for schema diffing and migration generation.

## Generate Migration Diff

Compare your current Go models against an existing schema and generate a migration:

```bash
ratel diff \
  -p github.com/myapp/models \
  --discover \
  -d ./migrations \
  -n add_user_profiles
```

This:
1. Generates the desired schema from your Go models
2. Compares it against the current database state
3. Produces a migration file in `./migrations/`

## From SQL File

You can also diff against a SQL schema file:

```bash
ratel diff \
  -s schema.sql \
  -d ./migrations \
  -n initial
```

## CLI Options

```
ratel diff [flags]

Flags:
  -s, --sql string         SQL schema file to compare against
  -p, --package string     Go package path containing models
  -t, --tables strings     Table variable names to include
  -d, --discover           Auto-discover tables from source files
  --dir string             Migration directory (default: ./migrations)
  -n, --name string        Migration name (default: migration)
  --pg_version int         PostgreSQL version (default: 18)
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

# Generate migration
ratel diff -p myapp/models --discover -d ./migrations -n add_bio_field

# Review
cat migrations/*_add_bio_field.sql

# Apply (using your preferred tool)
psql -f migrations/*_add_bio_field.sql
```
