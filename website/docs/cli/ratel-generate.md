---
sidebar_position: 3
title: ratel generate
---

# ratel generate

Generate Go model code from a SQL schema file.

## Usage

```bash
ratel generate [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input SQL schema file (required) |
| `--output` | `-o` | Output directory (default: ./models) |
| `--package` | `-p` | Package name for generated files (default: models) |

## Example

```bash
ratel generate -i schema.sql -o ./models -p mymodels
```

## Output

Generates Go files with:

- Table type definitions
- Column alias constants
- Scanner structs with `GetTarget`
- Table constructor with column definitions
- Global table variable
