# Handler pattern

Handlers are thin controllers. They: parse input, call a service, write a
response envelope. They do not contain business logic.

## Skeleton

```go
type OrdersHandler struct {
    orders domain.OrderService
}

func NewOrdersHandler(orders domain.OrderService) *OrdersHandler {
    return &OrdersHandler{orders: orders}
}

func (h *OrdersHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req api.CreateOrderRequest // generated from openapi.yaml
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        Error(w, r, domain.ErrInvalidInput)
        return
    }

    order, err := h.orders.Create(r.Context(), req)
    if err != nil {
        Error(w, r, err) // respond.go maps domain error → code + status
        return
    }

    JSON(w, r, http.StatusCreated, order)
}
```

## Rules

- Request types come from `internal/gen/api/types.gen.go`. Never hand-write request DTOs.
- Input validation beyond type coercion is the service's responsibility.
- Always pass `r.Context()` to service calls.
- Every error goes through `Error(w, r, err)` — never write error bodies manually.
- No SQL, no GORM, no third-party API calls in handlers. Those live in services or repos.

## Wiring

Handlers are constructed in `cmd/server/main.go` and passed into
`app.NewRouter(app.RouterDeps{...})`. The route registration itself happens
inside `internal/app/router.go` — do not wire routes in `main.go`. Adding a new
endpoint means:

1. Spec edit + `make gen:openapi`
2. Service method
3. Handler method
4. `app.RouterDeps` field (if a new handler struct)
5. Route registration in `router.go`

After step 5, the contract test passes.
