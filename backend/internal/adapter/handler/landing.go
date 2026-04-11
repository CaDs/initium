package handler

import "net/http"

// Landing returns public landing page data.
func Landing(w http.ResponseWriter, r *http.Request) {
	JSON(w, r, http.StatusOK, map[string]any{
		"name":        "Initium",
		"description": "Your next great idea starts here.",
		"version":     "0.1.0",
	})
}
