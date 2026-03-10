---
sidebar_position: 8
title: Repository Pattern
---

# Repository Pattern

Ratel provides repository abstractions for common CRUD operations.

## ScannerRepository

Works with Scanner structs directly:

```go
import "github.com/yaroher/ratel/pkg/repository"

repo := repository.NewScannerRepository(Users)

// Find by ID
user, err := repo.FindByID(ctx, db, 1)

// Find all with conditions
users, err := repo.FindAll(ctx, db,
    Users.SelectAll().
        Where(Users.IsActive.Eq(true)),
)

// Create
created, err := repo.Create(ctx, db, &UsersScanner{
    Email: "john@example.com",
    Name:  "John",
})

// Update
updated, err := repo.Update(ctx, db, &UsersScanner{
    ID:   1,
    Name: "John Smith",
})

// Delete
err = repo.Delete(ctx, db, 1)
```

## Transactions

Repositories work with any `DB` interface, including transactions:

```go
tx, err := pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Use the transaction
user, err := repo.Create(ctx, tx, &UsersScanner{
    Email: "john@example.com",
})
if err != nil {
    return err
}

order, err := orderRepo.Create(ctx, tx, &OrdersScanner{
    UserID: user.ID,
    Status: "NEW",
})
if err != nil {
    return err
}

return tx.Commit(ctx)
```

Since `pgx.Tx` satisfies the `DB` interface, you pass it directly to any query or repository method.
