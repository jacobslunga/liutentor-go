package daily

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/supabase-community/supabase-go"

	dailyservice "liutentor-go/internal/service/daily"
	examservice "liutentor-go/internal/service/exam"
)

type Handler struct {
	DB *supabase.Client
}

func NewHandler(db *supabase.Client) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) GetPuzzle(c *echo.Context) error {
	university := c.Param("university")

	if university == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Missing university"})
	}
	if !examservice.IsValidUniversity(university) {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Invalid university"})
	}

	result, err := dailyservice.GetPuzzle(examservice.University(university), h.DB)
	if err != nil {
		if err == dailyservice.ErrNoCandidates {
			return c.JSON(http.StatusServiceUnavailable, map[string]any{"success": false, "data": nil, "message": "No daily puzzle available"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"success": false, "data": nil, "message": "Failed to fetch the daily puzzle"})
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": result, "message": "Daily puzzle fetched successfully"})
}
