---
sidebar_position: 6
title: pkg/repository
---

# pkg/repository

The repository package provides higher-level CRUD abstractions over the exec and DML packages.

## ScannerRepository

```go
import "github.com/yaroher/ratel/pkg/repository"

repo := repository.NewScannerRepository(Users)
```

### Find

```go
// By ID
user, err := repo.FindByID(ctx, db, 1)

// With custom query
users, err := repo.FindAll(ctx, db,
    Users.SelectAll().
        Where(Users.IsActive.Eq(true)).
        OrderByDESC(Users.CreatedAt),
)
```

### Create

```go
created, err := repo.Create(ctx, db, &UsersScanner{
    Email: "john@example.com",
    Name:  "John Doe",
})
// created.ID is populated via RETURNING
```

### Update

```go
updated, err := repo.Update(ctx, db, &UsersScanner{
    ID:   1,
    Name: "John Smith",
})
```

### Delete

```go
err := repo.Delete(ctx, db, 1)
```

## Transaction Support

All repository methods accept any `DB` interface, including `pgx.Tx`:

```go
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)

user, _ := userRepo.Create(ctx, tx, &UsersScanner{...})
order, _ := orderRepo.Create(ctx, tx, &OrdersScanner{UserID: user.ID})

tx.Commit(ctx)
```
