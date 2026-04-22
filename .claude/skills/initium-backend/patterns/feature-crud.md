# Feature CRUD walkthrough

Step-by-step for adding a new resource to the backend. Example: `Order` with
create + list.

## 1. OpenAPI spec (always first)

`backend/api/openapi.yaml`:

```yaml
paths:
  /api/orders:
    post:
      summary: Create an order
      security: [{bearerAuth: []}]
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: '#/components/schemas/CreateOrderRequest'}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: '#/components/schemas/Order'}
        "400": {description: Invalid input, content: {application/json: {schema: {$ref: '#/components/schemas/ErrorResponse'}}}}
        "401": {description: Not authenticated, content: {application/json: {schema: {$ref: '#/components/schemas/ErrorResponse'}}}}
    get:
      summary: List current user's orders
      security: [{bearerAuth: []}]
      responses:
        "200":
          description: Order list
          content:
            application/json:
              schema: {$ref: '#/components/schemas/OrderList'}

components:
  schemas:
    Order:
      type: object
      required: [id, user_id, total_cents, created_at]
      properties:
        id: {type: string, format: uuid}
        user_id: {type: string, format: uuid}
        total_cents: {type: integer}
        created_at: {type: string, format: date-time}
    OrderList:                        # envelope — NEVER a bare array
      type: object
      required: [orders]
      properties:
        orders:
          type: array
          items: {$ref: '#/components/schemas/Order'}
    CreateOrderRequest:
      type: object
      required: [total_cents]
      properties:
        total_cents: {type: integer, minimum: 1}
```

Run `make gen:openapi`.

## 2. Domain

`backend/internal/domain/order.go`:

```go
package domain

import "time"

type Order struct {
    ID         string
    UserID     string
    TotalCents int
    CreatedAt  time.Time
}
```

`backend/internal/domain/errors.go` — add a sentinel **only** if the
client needs to distinguish this error programmatically. For simple "bad
input", reuse `domain.ErrInvalidInput`. For "amount must be positive" as
its own client-facing code, add:

```go
var ErrOrderAmountInvalid = errors.New("order amount invalid")
```

`backend/internal/domain/port.go` — interfaces use **plain Go types only**,
never `api.*` (keeps `domain/` import-clean):

```go
type OrderRepository interface {
    Save(ctx context.Context, o *Order) error
    ListForUser(ctx context.Context, userID string) ([]*Order, error)
}

type OrderService interface {
    Create(ctx context.Context, userID string, totalCents int) (*Order, error)
    List(ctx context.Context, userID string) ([]*Order, error)
}
```

The handler adapts `api.CreateOrderRequest` into plain args (see §7).

## 3. Error mapping

Only needed if you added a new sentinel above.
`backend/internal/adapter/handler/respond.go` (extend `mapError`):

```go
case errors.Is(err, domain.ErrOrderAmountInvalid):
    return "ORDER_AMOUNT_INVALID", http.StatusBadRequest
```

`backend/internal/adapter/handler/error_envelope_test.go` (add a row):

```go
{"order_amount_invalid", domain.ErrOrderAmountInvalid, http.StatusBadRequest, "ORDER_AMOUNT_INVALID"},
```

## 4. Persistence

`backend/internal/adapter/persistence/models.go` — add `OrderModel` with GORM
tags + `toDomain()` / `fromDomain()`.

`backend/internal/adapter/persistence/order_repo.go` — `GormOrderRepo`
implementing `domain.OrderRepository`. Copy shape from `user_repo.go`.

## 5. Migration

```bash
make db:create NAME=add_orders_table
```

Fill in the `.up.sql` / `.down.sql` pair. Never use `gorm.AutoMigrate`.

`make db:create` calls the `migrate` CLI (golang-migrate). If not
installed: `brew install golang-migrate` or
`go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`. If
the binary still isn't on PATH, hand-create two numbered files under
`backend/migrations/` (`NNN_name.up.sql` / `.down.sql`).

The schema uses `DEFAULT gen_random_uuid()` for `id` columns:

```sql
CREATE TABLE orders (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    total_cents INTEGER NOT NULL CHECK (total_cents > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 6. Service

`backend/internal/service/order.go`. Takes **plain args**, not
`api.CreateOrderRequest`. Keeps `domain/` import-clean.

```go
type OrderService struct {
    orders domain.OrderRepository
}

func NewOrderService(orders domain.OrderRepository) *OrderService { ... }

func (s *OrderService) Create(ctx context.Context, userID string, totalCents int) (*domain.Order, error) {
    if totalCents <= 0 {
        return nil, domain.ErrInvalidInput
    }
    o := &domain.Order{
        UserID:     userID,
        TotalCents: totalCents,
    }
    // ID + CreatedAt assigned by Postgres (DEFAULT gen_random_uuid() / now()).
    // The repo copies them back into `o` after INSERT.
    if err := s.orders.Save(ctx, o); err != nil {
        return nil, fmt.Errorf("saving order: %w", err)
    }
    return o, nil
}
```

**IDs are assigned by Postgres via `DEFAULT gen_random_uuid()`.** Services
never call `uuid.NewString()`. The repo's `Save` copies the DB-assigned
ID back into the domain entity after INSERT — see `user_repo.go`.

Service test at `backend/internal/service/order_test.go` — table-driven,
mock repo, `t.Parallel()`.

## 7. Handler

`backend/internal/adapter/handler/order.go`:

```go
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())

    var req api.CreateOrderRequest
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        Error(w, r, domain.ErrInvalidInput)
        return
    }

    // Handler adapts the generated request type into plain service args.
    o, err := h.orders.Create(r.Context(), userID, req.TotalCents)
    if err != nil {
        Error(w, r, err)
        return
    }

    writeOrder(w, r, http.StatusCreated, o)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())

    orders, err := h.orders.List(r.Context(), userID)
    if err != nil {
        Error(w, r, err)
        return
    }

    // Pre-allocate with len(orders) so JSON emits "orders":[] for zero
    // items, not "orders":null. A nil slice marshals to null, which
    // breaks every client's list-type assumption.
    apiOrders := make([]api.Order, 0, len(orders))
    for _, o := range orders {
        apiOrders = append(apiOrders, toAPIOrder(o))
    }
    JSON(w, r, http.StatusOK, api.OrderList{Orders: apiOrders})
}
```

Add a converter `toAPIOrder(*domain.Order) api.Order` for UUID / time
conversions, like `writeUser` in `user.go`. `writeUser` returns via
`JSON(...)` directly and can 500 on a bad UUID; follow that shape for any
response that needs error handling on conversion.

Handler test at `order_test.go` — use `withUser()` helper from
`patterns/test.md` to seed `middleware.UserIDKey`.

## 8. Wiring

`backend/internal/app/router.go` — add to `RouterDeps` and register the route
inside the protected `r.Group` that already applies `middleware.Auth`.

`backend/internal/app/contract_test.go` — both `TestRouter_MatchesOpenAPISpec`
and `TestRouter_NoDebugRoutesInProduction` construct their own `RouterDeps{}`;
add the new handler field to both (nil is fine — the tests don't invoke
handlers).

`backend/cmd/server/main.go` — construct `OrderRepo`, `OrderService`,
`OrderHandler` and pass them into `RouterDeps`.

## 9. Verify

```bash
make preflight    # lint + test + check:openapi + check:parity + check:skills + check:staged
```

Or run the individual gates:
```bash
make test:backend          # contract test passes = route↔spec parity
make check:openapi         # DTO drift still clean
make check:parity          # every /api/ path has a client consumer
```

## 10. Parity

Backend endpoints almost always imply web + mobile UI. Either land the
mirror work or explicitly call out "N/A because X" in the PR description.
See `.claude/skills/_shared/parity.md`.
