package course

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/supabase-community/supabase-go"

	courseservice "liutentor-go/internal/service/course"
	examservice "liutentor-go/internal/service/exam"
)

type Handler struct {
	DB *supabase.Client
}

func NewHandler(db *supabase.Client) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) GetCourses(c *echo.Context) error {
	university := c.Param("university")

	if university == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Missing university"})
	}
	if !examservice.IsValidUniversity(university) {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Invalid university"})
	}

	result, err := courseservice.GetCourses(examservice.University(university), h.DB)
	if err != nil {
		if err == courseservice.ErrNoCourses {
			return c.JSON(http.StatusNotFound, map[string]any{"success": false, "data": nil, "message": "No courses found for this university"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"success": false, "data": nil, "message": "Failed to fetch courses"})
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": result, "message": "Courses fetched successfully"})
}
