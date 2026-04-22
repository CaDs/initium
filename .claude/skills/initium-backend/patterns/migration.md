# Migration pattern

Schema changes go through numbered SQL files under `backend/migrations/`.
**Never** use `gorm.AutoMigrate`.

## Creating a migration

```bash
make db:create NAME=add_orders_table
```

This produces two files:

```
backend/migrations/000NN_add_orders_table.up.sql
backend/migrations/000NN_add_orders_table.down.sql
```

## Rules

- Every up migration has a matching down migration (or `-- irreversible`
  with a clear comment explaining why).
- Prefer additive changes. Adding a NOT NULL column to a populated table
  requires a three-step dance: add nullable → backfill → alter to NOT NULL.
- Parameterized names only — migration names are free text but should be
  short and describe intent (`add_orders_role_column`, not `fix1`).
- Never include seed data in migrations. Use `make db:seed` (and
  `backend/cmd/seed/main.go`) for dev data.

## Rolling out

```bash
make db:migrate   # apply pending
make db:rollback  # roll back one
make db:reset     # drop + migrate (destroys data, DEV ONLY)
```

## GORM models

When adding a table, add the matching GORM struct in
`backend/internal/adapter/persistence/models.go` with `toDomain()` and
`fromDomain()` mappers. Domain entities in `internal/domain/` must not have
GORM tags — those live only in the persistence model.
