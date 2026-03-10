---
sidebar_position: 5
title: ratel migrate
---

# ratel migrate

Migration management commands.

## ratel migrate hash

Recalculate `atlas.sum` checksum for a migrations directory. Use this after manually adding or editing `.sql` migration files.

### Usage

```bash
ratel migrate hash [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dir` | `-d` | Migration directory (default: `./migrations`) |
| `--engine` | `-e` | Migration engine: `atlas` or `ratel` (default: `atlas`) |

### Examples

```bash
# Recalculate after manual edit
ratel migrate hash -d ./migrations

# With ratel engine
ratel migrate hash -d ./migrations -e ratel
```

### Why?

Both Ratel and Atlas track migration file integrity via `atlas.sum`. If you manually add, edit, or delete a `.sql` file in the migrations directory, `atlas.sum` becomes stale and tools will reject the directory.

Running `ratel migrate hash` recomputes the checksum from all `.sql` files in the directory.
