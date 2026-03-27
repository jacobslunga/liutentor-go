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
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Missing courseCode"})
	}
	if university == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Missing university"})
	}
	if !examservice.IsValidUniversity(university) {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "Invalid university"})
	}

	result, err := examservice.GetExams(courseCode, examservice.University(university), h.DB)
	if err != nil {
		if err == examservice.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"success": false, "data": nil, "message": "No exam documents found for this course"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"success": false, "data": nil, "message": "Failed to fetch exams"})
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": result, "message": "Exams fetched successfully"})
}

func (h *Handler) GetExam(c *echo.Context) error {
	examIDStr := c.Param("examId")

	examID, err := strconv.Atoi(examIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": "examId must be a positive integer"})
	}

	result, err := examservice.GetExam(examID, h.DB)
	if err != nil {
		if err == examservice.ErrInvalidID {
			return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "data": nil, "message": err.Error()})
		}
		return c.JSON(http.StatusNotFound, map[string]any{"success": false, "data": nil, "message": "Exam not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": result, "message": "Exam fetched successfully"})
}
