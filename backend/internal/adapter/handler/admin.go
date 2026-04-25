package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// adminPingOutput is the response body for GET /api/admin/ping.
type adminPingOutput struct {
	Body AdminPingResponse
}

// RegisterAdmin wires admin-only endpoints onto the Huma API. Both auth
// and role middlewares are passed in from the caller (app/api.go) and
// attached to each operation.
func RegisterAdmin(
	api huma.API,
	authMW func(huma.Context, func(huma.Context)),
	requireAdmin func(huma.Context, func(huma.Context)),
) {
	huma.Register(api, huma.Operation{
		OperationID: "admin-ping",
		Method:      http.MethodGet,
		Path:        "/api/admin/ping",
		Summary:     "Verify admin role enforcement",
		Tags:        []string{"admin"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{authMW, requireAdmin},
	}, func(_ context.Context, _ *struct{}) (*adminPingOutput, error) {
		return &adminPingOutput{Body: AdminPingResponse{Role: "admin"}}, nil
	})
}
