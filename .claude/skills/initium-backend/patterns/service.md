# Service pattern

Services hold business logic and implement interfaces declared in `domain/port.go`.

## Skeleton

```go
// In domain/port.go:
type OrderService interface {
    Create(ctx context.Context, req api.CreateOrderRequest) (*domain.Order, error)
    Get(ctx context.Context, id string) (*domain.Order, error)
}

// In service/orders.go:
type OrdersService struct {
    orders domain.OrderRepository
    users  domain.UserRepository
    logger *slog.Logger
}

func NewOrdersService(orders domain.OrderRepository, users domain.UserRepository) *OrdersService {
    return &OrdersService{orders: orders, users: users, logger: slog.Default()}
}

func (s *OrdersService) Create(ctx context.Context, req api.CreateOrderRequest) (*domain.Order, error) {
    user, err := s.users.FindByID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("loading user: %w", err)
    }
    if !user.CanPlaceOrder() {
        return nil, domain.ErrOrderForbidden
    }

    order := domain.NewOrder(user.ID, req.Items)
    if err := s.orders.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("saving order: %w", err)
    }
    return order, nil
}
```

## Rules

- Accept interfaces (`domain.OrderRepository`), return structs (`*domain.Order`).
- `context.Context` is always the first parameter.
- Wrap every error with meaningful context — `fmt.Errorf("operation: %w", err)`.
- Return domain error sentinels for expected cases (`domain.ErrOrderForbidden`),
  let infrastructure errors bubble up wrapped.
- No HTTP types, no SQL, no framework imports.
- Use `slog` for logging. Redact PII.

## Testing

- Mock repositories via hand-written stubs in the test file (follow
  `service/auth_test.go`).
- Table-driven tests. Name: `TestOrdersService_Create_UserCannotOrder_ReturnsForbidden`.
- Assert on behavior (return value, repository calls), not implementation detail.
