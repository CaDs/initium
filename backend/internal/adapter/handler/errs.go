package handler

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
)

// APIError is the wire-shape error envelope returned to clients. Matches
// the existing ErrorResponse schema (code, message, request_id) verbatim
// so web Zod schemas, iOS Codable, and Android Moshi keep parsing without
// changes after the Huma migration.
//
// Implements huma.StatusError. When a Huma handler returns
// `(nil, &APIError{...})`, Huma uses GetStatus() for the HTTP status and
// serializes the struct to JSON via the field tags below.
type APIError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code" doc:"Machine-readable error code in SNAKE_UPPER format"`
	Message    string `json:"message" doc:"Human-readable error message"`
	RequestID  string `json:"request_id,omitempty" doc:"Correlates client error to server log entry"`
}

// Error satisfies the error interface.
func (e *APIError) Error() string { return e.Message }

// GetStatus satisfies huma.StatusError so Huma writes the right status code.
func (e *APIError) GetStatus() int { return e.HTTPStatus }

// MapDomainErr converts a domain error to the APIError envelope, ready to
// return from a Huma handler. Pulls request_id from the request context
// (set by middleware.RequestID). For unmapped errors the original error
// is logged server-side; the client receives "internal error" so internals
// never leak.
func MapDomainErr(ctx context.Context, err error) *APIError {
	code, status := mapError(err)
	msg := err.Error()
	if code == "INTERNAL_ERROR" {
		slog.Error("internal error", "error", err, "request_id", middleware.GetRequestID(ctx))
		msg = "internal error"
	}
	return &APIError{
		HTTPStatus: status,
		Code:       code,
		Message:    msg,
		RequestID:  middleware.GetRequestID(ctx),
	}
}

// InstallErrorEnvelope overrides huma.NewError so error responses generated
// by Huma itself (validation failures, missing-body errors) match the
// existing ErrorResponse wire shape. Called once at startup from app/api.go.
//
// huma.NewError is a package-level var; setting it once at process start is
// safe for our single-process model. Tests that build a fresh humatest API
// rely on this being called too — handlers/init_test.go takes care of it.
func InstallErrorEnvelope() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		// If the caller already produced an APIError (handlers via MapDomainErr),
		// pass it through unchanged — that path carries our code + request_id.
		for _, e := range errs {
			if ae, ok := e.(*APIError); ok {
				return ae
			}
		}
		// Otherwise this is Huma synthesizing an error (e.g. validation
		// failure). Best-effort code derivation from status.
		code := "INTERNAL_ERROR"
		switch status {
		case 400, 422:
			// 422 is what Huma returns when struct-tag validation fails
			// (required, format, minLength, etc.). Map to INVALID_INPUT
			// so clients see the same code as for hand-rolled 400s.
			code = "INVALID_INPUT"
		case 401:
			code = "INVALID_CREDENTIALS"
		case 403:
			code = "FORBIDDEN"
		case 404:
			code = "NOT_FOUND"
		case 409:
			code = "CONFLICT"
		case 429:
			code = "RATE_LIMITED"
		}
		return &APIError{HTTPStatus: status, Code: code, Message: msg}
	}
}
