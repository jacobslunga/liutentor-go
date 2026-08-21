package course

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"

	examservice "liutentor-go/internal/service/exam"
)

// pageSize is the number of rows fetched per PostgREST request. Supabase caps
// a single response at 1000 rows, so anything larger is silently truncated.
const pageSize = 1000

// cacheTTL keeps the course index warm between requests. Building it walks the
// whole exams and exam_stats tables, which is far too expensive to repeat on
// every call — and the list only changes when the scraper adds an exam.
const cacheTTL = time.Hour

// Course is one course code that has at least one exam in the archive.
type Course struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ExamCount int    `json:"examCount"`
}

type CoursesResult struct {
	Courses []Course `json:"courses"`
}

var ErrNoCourses = errors.New("no courses found")

type cacheEntry struct {
	result    *CoursesResult
	fetchedAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = map[examservice.University]cacheEntry{}
)

func cached(university examservice.University) *CoursesResult {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	entry, ok := cache[university]
	if !ok || time.Since(entry.fetchedAt) > cacheTTL {
		return nil
	}
	return entry.result
}

func store(university examservice.University, result *CoursesResult) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache[university] = cacheEntry{result: result, fetchedAt: time.Now()}
}

// GetCourses returns every course code with at least one exam for the given
// university, ordered by exam count descending. The code tie-break keeps the
// ordering stable across calls, which callers rely on.
func GetCourses(university examservice.University, db *supabase.Client) (*CoursesResult, error) {
	if result := cached(university); result != nil {
		return result, nil
	}

	counts, err := countExams(university, db)
	if err != nil {
		return nil, err
	}
	if len(counts) == 0 {
		return nil, ErrNoCourses
	}

	names, err := courseNames(db)
	if err != nil {
		// Names are a nicety, not a requirement — a failure here should not
		// take the whole course list down with it.
		names = map[string]string{}
	}

	courses := make([]Course, 0, len(counts))
	for code, count := range counts {
		courses = append(courses, Course{Code: code, Name: names[code], ExamCount: count})
	}

	sort.Slice(courses, func(i, j int) bool {
		if courses[i].ExamCount != courses[j].ExamCount {
			return courses[i].ExamCount > courses[j].ExamCount
		}
		return courses[i].Code < courses[j].Code
	})

	result := &CoursesResult{Courses: courses}
	store(university, result)

	return result, nil
}

func countExams(university examservice.University, db *supabase.Client) (map[string]int, error) {
	counts := map[string]int{}

	for offset := 0; ; offset += pageSize {
		var rows []struct {
			CourseCode string `json:"course_code"`
		}

		_, err := db.From("exams").
			Select("course_code", "exact", false).
			Eq("university", string(university)).
			Order("id", &postgrest.OrderOpts{Ascending: true}).
			Range(offset, offset+pageSize-1, "").
			ExecuteTo(&rows)
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			if code := normalize(row.CourseCode); code != "" {
				counts[code]++
			}
		}

		if len(rows) < pageSize {
			break
		}
	}

	return counts, nil
}

func courseNames(db *supabase.Client) (map[string]string, error) {
	names := map[string]string{}

	for offset := 0; ; offset += pageSize {
		var rows []struct {
			CourseCode    string `json:"course_code"`
			CourseNameSwe string `json:"course_name_swe"`
			CourseNameEng string `json:"course_name_eng"`
		}

		_, err := db.From("exam_stats").
			Select("course_code, course_name_swe, course_name_eng", "exact", false).
			Order("course_code", &postgrest.OrderOpts{Ascending: true}).
			Order("exam_date", &postgrest.OrderOpts{Ascending: true}).
			Range(offset, offset+pageSize-1, "").
			ExecuteTo(&rows)
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			code := normalize(row.CourseCode)
			if code == "" {
				continue
			}

			name := strings.TrimSpace(row.CourseNameSwe)
			if name == "" {
				name = strings.TrimSpace(row.CourseNameEng)
			}
			if name != "" {
				names[code] = name
			}
		}

		if len(rows) < pageSize {
			break
		}
	}

	return names, nil
}

func normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
