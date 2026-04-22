# Handler pattern

Handlers are thin controllers. They: decode the request into a generated type,
call a service, write a response envelope using a generated type. No business
logic.

## Skeleton (authenticated endpoint)

```go
type OrdersHandler struct {
    orders domain.OrderService
}

func NewOrdersHandler(orders domain.OrderService) *OrdersHandler {
    return &OrdersHandler{orders: orders}
}

func (h *OrdersHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())

    var req api.CreateOrderRequest // generated from openapi.yaml
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        Error(w, r, domain.ErrInvalidInput)
        return
    }

    order, err := h.orders.Create(r.Context(), userID, req)
    if err != nil {
        Error(w, r, err)
        return
    }

    JSON(w, r, http.StatusCreated, toAPIOrder(order))
}
```

## Rules

- **Request types come from `internal/gen/api/types.gen.go`.** Never hand-write
  request DTOs. Every handler in this repo uses `api.*` request types — see
  `auth.go`, `user.go`, `mobile_auth.go`.
- **Response types come from `internal/gen/api/types.gen.go`.** When the domain
  type differs from the wire type (e.g. `domain.User.ID string` vs
  `api.User.Id openapi_types.UUID`), write a small converter like `writeUser`
  in `user.go`. Never return `map[string]any`.
- **Authenticated endpoints** start with `userID := middleware.GetUserID(r.Context())`.
  The `middleware.Auth` chain in `internal/app/router.go` seeds it; handlers
  trust it.
- **Strict JSON decoding**: always `dec.DisallowUnknownFields()`. Unknown keys
  should 400 (via `domain.ErrInvalidInput`), not silently succeed.
- Input validation beyond type coercion is the service's responsibility.
- Every error goes through `Error(w, r, err)` — never write error bodies manually.
- No SQL, no GORM, no third-party API calls in handlers. Those live in services
  or repos.

## Converting domain → api response type

For types where the domain representation differs from the wire (string IDs,
raw enums), encapsulate the conversion:

```go
func writeUser(w http.ResponseWriter, r *http.Request, u *domain.User) {
    id, err := uuid.Parse(u.ID)
    if err != nil {
        slog.Error("invalid user uuid", "id", u.ID, "error", err)
        Error(w, r, err)
        return
    }
    JSON(w, r, http.StatusOK, api.User{
        Id:        openapi_types.UUID(id),
        Email:     openapi_types.Email(u.Email),
        Name:      u.Name,
        AvatarUrl: u.AvatarURL,
        Role:      api.UserRole(u.Role),
        CreatedAt: u.CreatedAt,
    })
}
```

## Wiring

Handlers are constructed in `cmd/server/main.go` and passed into
`app.NewRouter(app.RouterDeps{...})`. The route registration itself happens
inside `internal/app/router.go` — do not wire routes in `main.go`. Adding a new
endpoint means:

1. Spec edit (including `required:` arrays on every schema) + `make gen:openapi`
2. Service method (accepts generated request type, returns domain entity)
3. Handler method (decodes request, calls service, converts domain → api response)
4. `app.RouterDeps` field (if a new handler struct)
5. Route registration in `router.go`

After step 5, the contract test passes. The handler test using `withUser()`
(see `patterns/test.md`) covers the auth'd path.
