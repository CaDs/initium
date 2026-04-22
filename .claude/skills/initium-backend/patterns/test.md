# Test pattern

## Layout

- Unit tests live alongside the code: `service/auth.go` → `service/auth_test.go`.
- Handler tests in `adapter/handler/*_test.go`. Use `httptest.NewRecorder`.
- Contract/integration tests in `internal/app/contract_test.go` (route↔spec).
- Error envelope tests in `internal/adapter/handler/error_envelope_test.go`.

## Pattern

```go
func TestAuthService_VerifyMagicLink_TokenUsed_ReturnsConflict(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name    string
        token   string
        setup   func(*mockSessionRepo)
        wantErr error
    }{
        {
            name:  "token already used returns ErrTokenUsed",
            token: "abc",
            setup: func(m *mockSessionRepo) {
                m.findMagicLinkFn = func(ctx context.Context, hash string) (*domain.MagicLinkToken, error) {
                    now := time.Now()
                    return &domain.MagicLinkToken{UsedAt: &now}, nil
                }
            },
            wantErr: domain.ErrTokenUsed,
        },
        // ...
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            repo := &mockSessionRepo{}
            tc.setup(repo)
            svc := NewAuthService(nil, repo, nil, nil, nil)
            _, _, err := svc.VerifyMagicLink(context.Background(), tc.token)
            require.ErrorIs(t, err, tc.wantErr)
        })
    }
}
```

## Rules

- `t.Parallel()` unless tests share state.
- `require` for fail-fast preconditions, `assert` for subsequent checks.
- Mock at the domain interface boundary. See `service/auth_test.go` for
  hand-written mock stubs — no mocking library needed.
- 80% coverage floor. Every bug fix lands with a regression test.
- Never test generated code.
- Integration tests go under a `//go:build integration` tag if they need a
  live DB; run them via a dedicated CI job, not the default `make test:backend`.

## Contract tests

`internal/app/contract_test.go` enforces route↔spec parity and the
"no /_debug routes in production" guarantee. If it fails, either your new
handler is missing from the OpenAPI spec or vice versa. Do not add to the
`excludedPaths` list without a strong justification.

The walker strips trailing `/*` from chi routes (see `contract_test.go:70`).
Paths with URL params like `/api/notes/{id}` must appear in `openapi.yaml`
using `{id}` syntax (matches chi's `{id}` pattern directly).

## Testing an authenticated handler

The `middleware.Auth` chain seeds `middleware.UserIDKey` into the request
context. Handler tests must inject it manually via a small helper:

```go
func withUser(r *http.Request, userID string) *http.Request {
    ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
    return r.WithContext(ctx)
}

func TestUserHandler_GetProfile_ReturnsUser(t *testing.T) {
    t.Parallel()

    svc := &mockUserService{
        getProfileFn: func(_ context.Context, id string) (*domain.User, error) {
            return &domain.User{
                ID:    id,
                Email: "dev@example.com",
                Name:  "Dev",
                Role:  "user",
            }, nil
        },
    }
    h := NewUserHandler(svc)

    req := withUser(httptest.NewRequest(http.MethodGet, "/api/me", nil), "00000000-0000-0000-0000-000000000001")
    rec := httptest.NewRecorder()
    h.GetProfile(rec, req)

    require.Equal(t, http.StatusOK, rec.Code)

    var resp api.User
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
    assert.Equal(t, "dev@example.com", string(resp.Email))
    assert.Equal(t, api.UserRole("user"), resp.Role)
}
```

- Seed a valid UUID string — `writeUser` parses it via `uuid.Parse` and will
  500 if invalid.
- Decode the response into the generated `api.*` type, not `map[string]any`.
  That's the whole point of the typed response: tests assert on the wire
  contract.
