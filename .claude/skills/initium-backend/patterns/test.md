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
