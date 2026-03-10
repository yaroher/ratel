---
sidebar_position: 2
title: ratel diff
---

# ratel diff

Generate migration diffs by comparing your models against the current database state.

## Usage

```bash
ratel diff [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--sql` | `-s` | SQL schema file to compare against |
| `--package` | `-p` | Go package path(s) containing models (comma-separated or repeated) |
| `--tables` | `-t` | Table variable names |
| `--discover` | `-d` | Auto-discover tables |
| `--dir` | | Migration directory (default: ./migrations) |
| `--name` | `-n` | Migration name (default: migration) |
| `--pg_version` | | PostgreSQL version (default: 18) |
| `--engine` | `-e` | Migration engine: `atlas` or `ratel` (default: `atlas`) |

## Migration Engines

### `atlas` (default)

Uses Atlas OSS for diffing. Supports tables, columns, indexes, foreign keys, and check constraints.

### `ratel`

Native engine with full PostgreSQL support. Use this when you need **Row Level Security**, **triggers**, **functions**, or **extensions** — features that require Atlas Pro (paid license).

```bash
ratel diff --engine ratel -p github.com/myapp/models --discover \
  -d ./migrations -n add_rls_policies
```

## Examples

```bash
# From Go models (single package)
ratel diff -p github.com/myapp/models --discover \
  -d ./migrations -n add_profiles

# From multiple packages
ratel diff -p github.com/myapp/auth,github.com/myapp/store \
  --discover -d ./migrations -n init

# From SQL file
ratel diff -s schema.sql -d ./migrations -n initial

# With RLS/triggers support
ratel diff --engine ratel -p github.com/myapp/models --discover \
  -d ./migrations -n add_security
```

## How It Works

1. Generates the desired schema from your models (or reads from SQL file)
2. Compares against the current database state (from applied migrations)
3. Produces a migration file with the required changes
4. Updates `atlas.sum` checksum for compatibility

The generated migration file is placed in the `--dir` directory with a timestamp prefix.
