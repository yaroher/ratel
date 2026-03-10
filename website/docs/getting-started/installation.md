---
sidebar_position: 1
title: Installation
---

# Installation

## Prerequisites

- **Go 1.22+**
- **PostgreSQL 14+**
- **protoc** (Protocol Buffers compiler) — if using proto-first approach

## Install Ratel CLI

```bash
go install github.com/yaroher/ratel/cmd/ratel@latest
```

## Install Protoc Plugin

If you define your schema using Protocol Buffers:

```bash
go install github.com/yaroher/ratel/cmd/protoc-gen-ratel@latest
```

You also need the plain Go message generator (generates Go structs without protobuf runtime dependency):

```bash
go install github.com/yaroher/protoc-gen-go-plain/cmd/protoc-gen-plain@latest
```

And the standard Go protobuf generator:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

## Add Ratel as a Dependency

```bash
go get github.com/yaroher/ratel
```

## Verify Installation

```bash
ratel --help
protoc-gen-ratel --version
```

## What's Next

Choose your workflow:

- **[Quick Start (Proto)](/docs/getting-started/quick-start-proto)** — define schema in `.proto` files, generate everything
- **[Quick Start (SQL)](/docs/getting-started/quick-start-sql)** — define models manually in Go
