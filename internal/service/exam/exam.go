package exam

import (
	"errors"
	"fmt"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
)

type University string

const (
	LIU University = "LIU"
	KTH University = "KTH"
	CTH University = "CTH"
	LTH University = "LTH"
)

var ValidUniversities = []University{LIU, KTH, CTH, LTH}

func IsValidUniversity(u string) bool {
	for _, v := range ValidUniversities {
		if University(u) == v {
			return true
		}
	}
	return false
}

type Exam struct {
	ID          int     `json:"id"`
	CourseCode  string  `json:"course_code"`
	ExamDate    string  `json:"exam_date"`
	PdfURL      string  `json:"pdf_url"`
	ExamName    string  `json:"exam_name"`
	HasSolution bool    `json:"has_solution"`
	Statistics  any     `json:"statistics"`
	PassRate    float64 `json:"pass_rate"`
}

type ExamsResult struct {
	CourseCode string `json:"courseCode"`
	CourseName string `json:"courseName"`
	Exams      []Exam `json:"exams"`
}

type Solution struct {
	ID     int `json:"id"`
	ExamID int `json:"exam_id"`
}

type ExamResult struct {
	Exam     any `json:"exam"`
	Solution any `json:"solution"`
}

var ErrNotFound = errors.New("not found")
var ErrInvalidID = errors.New("examId must be a positive integer")

func GetExams(courseCode string, university University, db *supabase.Client) (*ExamsResult, error) {
	var examsData []map[string]any
	_, err := db.From("exams").
		Select("id, course_code, exam_date, pdf_url, exam_name, solutions(exam_id)", "exact", false).
		Eq("course_code", courseCode).
		Eq("university", string(university)).
		Order("exam_date", &postgrest.OrderOpts{Ascending: false}).
		ExecuteTo(&examsData)
	if err != nil || len(examsData) == 0 {
		return nil, ErrNotFound
	}

	var statsData []map[string]any
	_, _ = db.From("exam_stats").
		Select("exam_date, statistics, pass_rate, course_name_swe", "exact", false).
		Eq("course_code", courseCode).
		ExecuteTo(&statsData)

	statsMap := make(map[string]map[string]any)
	courseName := ""
	for _, stat := range statsData {
		date, _ := stat["exam_date"].(string)
		statsMap[date] = stat
		if name, ok := stat["course_name_swe"].(string); ok && name != "" {
			courseName = name
		}
	}

	exams := make([]Exam, 0, len(examsData))
	for _, e := range examsData {
		date, _ := e["exam_date"].(string)
		stats := statsMap[date]

		solutions, _ := e["solutions"].([]any)
		passRate, _ := stats["pass_rate"].(float64)

		exams = append(exams, Exam{
			ID:          int(e["id"].(float64)),
			CourseCode:  e["course_code"].(string),
			ExamDate:    date,
			PdfURL:      e["pdf_url"].(string),
			ExamName:    e["exam_name"].(string),
			HasSolution: len(solutions) > 0,
			Statistics:  stats["statistics"],
			PassRate:    passRate,
		})
	}

	return &ExamsResult{
		CourseCode: courseCode,
		CourseName: courseName,
		Exams:      exams,
	}, nil
}

func GetExam(examID int, db *supabase.Client) (*ExamResult, error) {
	if examID <= 0 {
		return nil, ErrInvalidID
	}

	var data []map[string]any
	_, err := db.From("exams").
		Select("id, course_code, exam_date, pdf_url, solutions(*)", "exact", false).
		Eq("id", fmt.Sprintf("%d", examID)).
		ExecuteTo(&data)
	if err != nil || len(data) == 0 {
		return nil, ErrNotFound
	}

	row := data[0]
	solutions, _ := row["solutions"].([]any)

	exam := map[string]any{
		"id":          row["id"],
		"course_code": row["course_code"],
		"exam_date":   row["exam_date"],
		"pdf_url":     row["pdf_url"],
	}

	var solution any = nil
	if len(solutions) > 0 {
		solution = solutions[0]
	}

	return &ExamResult{Exam: exam, Solution: solution}, nil
}
