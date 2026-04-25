package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// LandingInfo is the public landing payload — name, tagline, version.
// First Huma-typed schema in this codebase; serves as the worked
// example for new endpoints.
type LandingInfo struct {
	Name        string `json:"name" doc:"Product name"`
	Description string `json:"description" doc:"Tagline shown on the landing page"`
	Version     string `json:"version" doc:"API version"`
}

type landingOutput struct {
	Body LandingInfo
}

// RegisterLanding wires GET /api/landing onto the Huma API.
func RegisterLanding(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-landing",
		Method:      http.MethodGet,
		Path:        "/api/landing",
		Summary:     "Public landing page data",
		Tags:        []string{"public"},
	}, func(_ context.Context, _ *struct{}) (*landingOutput, error) {
		return &landingOutput{Body: LandingInfo{
			Name:        "Initium",
			Description: "Your next great idea starts here.",
			Version:     "0.1.0",
		}}, nil
	})
}
