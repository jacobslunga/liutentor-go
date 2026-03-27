package exam

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/supabase-community/supabase-go"

	examservice "liutentor-go-api/internal/service/exam"
)

type Handler struct {
	DB *supabase.Client
}

func NewHandler(db *supabase.Client) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) GetExams(c *echo.Context) error {
	courseCode := c.Param("courseCode")
	university := c.Param("university")

	if courseCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "Missing courseCode"})
	}
	if university == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "Missing university"})
	}
	if !examservice.IsValidUniversity(university) {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid university"})
	}

	result, err := examservice.GetExams(courseCode, examservice.University(university), h.DB)
	if err != nil {
		if err == examservice.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"error": "No exam documents found for this course"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to fetch exams"})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result, "message": "Exams fetched successfully"})
}

func (h *Handler) GetExam(c *echo.Context) error {
	examIDStr := c.Param("examId")

	examID, err := strconv.Atoi(examIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "examId must be a positive integer"})
	}

	result, err := examservice.GetExam(examID, h.DB)
	if err != nil {
		if err == examservice.ErrInvalidID {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return c.JSON(http.StatusNotFound, map[string]any{"error": "Exam not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result, "message": "Exam fetched successfully"})
}
