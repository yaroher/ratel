---
sidebar_position: 2
title: ratel diff
---

# ratel diff

Generate migration diffs using Atlas.

## Usage

```bash
ratel diff [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--sql` | `-s` | SQL schema file to compare against |
| `--package` | `-p` | Go package path containing models |
| `--tables` | `-t` | Table variable names |
| `--discover` | `-d` | Auto-discover tables |
| `--dir` | | Migration directory (default: ./migrations) |
| `--name` | `-n` | Migration name (default: migration) |
| `--pg_version` | | PostgreSQL version (default: 18) |

## Examples

```bash
# From Go models
ratel diff -p github.com/myapp/models --discover \
  -d ./migrations -n add_profiles

# From SQL file
ratel diff -s schema.sql -d ./migrations -n initial
```

## How It Works

1. Generates the desired schema from your models (or reads from SQL file)
2. Uses Atlas to diff against the migration directory state
3. Produces a migration file with the required changes

The generated migration file is placed in the `--dir` directory with a timestamp prefix.
