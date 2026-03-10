---
sidebar_position: 1
title: ratel schema
---

# ratel schema

Generate SQL schema from Go table definitions.

## Usage

```bash
ratel schema [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--package` | `-p` | Go package path containing models (required) |
| `--tables` | `-t` | Table variable names to include (comma-separated) |
| `--discover` | `-d` | Auto-discover tables from source files |
| `--output` | `-o` | Output SQL file (default: schema.sql) |

## Examples

```bash
# Auto-discover all tables
ratel schema -p github.com/myapp/models --discover -o schema.sql

# Specific tables
ratel schema -p github.com/myapp/models -t Users,Orders,Products -o schema.sql
```

## Output

Generates a `.sql` file with:

1. `CREATE SCHEMA IF NOT EXISTS` statements
2. `CREATE TABLE` statements in topological order (respecting FK dependencies)
3. `CREATE INDEX` statements
4. Post-statements (RLS, grants, etc.)
