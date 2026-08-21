package daily

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"

	courseservice "liutentor-go/internal/service/course"
	examservice "liutentor-go/internal/service/exam"
)

// minExams is how many exams a course needs before it can be the answer. The
// archive is full of one-off codes nobody would recognise, let alone guess.
const minExams = 10

// repeatWindow is how many recent days are excluded from the draw, so the same
// course cannot come back within roughly six weeks.
const repeatWindow = 45

// stockholm is the timezone that decides when the puzzle rolls over. Everyone
// gets the new code at the same Swedish midnight, wherever they are.
var stockholm *time.Location

func init() {
	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		loc = time.FixedZone("CET", 3600)
	}
	stockholm = loc
}

type Puzzle struct {
	Date       string `json:"date"`
	CourseCode string `json:"courseCode"`
	CourseName string `json:"courseName"`
}

type puzzleRow struct {
	PuzzleDate string `json:"puzzle_date"`
	University string `json:"university"`
	CourseCode string `json:"course_code"`
}

var ErrNoCandidates = errors.New("no eligible courses for the daily puzzle")

// Today returns the date in Stockholm as YYYY-MM-DD.
func Today() string {
	return time.Now().In(stockholm).Format("2006-01-02")
}

// GetPuzzle returns the answer for today, choosing and recording one the first
// time it is asked for.
//
// The answer is stored rather than derived so it cannot move once players have
// started. Deriving it from the date meant it depended on the candidate pool
// being byte-identical across every server and every moment of the day, and the
// pool tracks live exam counts.
func GetPuzzle(university examservice.University, db *supabase.Client) (*Puzzle, error) {
	date := Today()

	if code, err := storedCode(date, university, db); err == nil && code != "" {
		return withName(date, code, university, db)
	}

	code, err := choose(date, university, db)
	if err != nil {
		return nil, err
	}

	// Two first-requests of the day can race here. A plain insert is deliberate:
	// the primary key makes the loser's insert fail, which is the point — an
	// upsert would let it overwrite an answer someone may already be playing.
	// Re-reading afterwards means both callers return whichever row landed.
	_, _, insertErr := db.From("daily_puzzle").
		Insert(puzzleRow{PuzzleDate: date, University: string(university), CourseCode: code},
			false, "", "minimal", "").
		Execute()

	stored, err := storedCode(date, university, db)
	if err != nil || stored == "" {
		if insertErr != nil {
			return nil, insertErr
		}
		return nil, ErrNoCandidates
	}
	code = stored

	return withName(date, code, university, db)
}

func storedCode(date string, university examservice.University, db *supabase.Client) (string, error) {
	var rows []puzzleRow

	_, err := db.From("daily_puzzle").
		Select("puzzle_date, university, course_code", "exact", false).
		Eq("puzzle_date", date).
		Eq("university", string(university)).
		ExecuteTo(&rows)
	if err != nil || len(rows) == 0 {
		return "", err
	}

	return rows[0].CourseCode, nil
}

// choose picks the day's answer from the eligible courses, skipping anything
// that has aired recently.
func choose(date string, university examservice.University, db *supabase.Client) (string, error) {
	index, err := courseservice.GetCourses(university, db)
	if err != nil {
		// A university with nothing in the archive has no puzzle, which is a
		// fact about the data rather than a server fault.
		if errors.Is(err, courseservice.ErrNoCourses) {
			return "", ErrNoCandidates
		}
		return "", err
	}

	eligible := EligibleCodes(index.Courses)
	if len(eligible) == 0 {
		return "", ErrNoCandidates
	}

	return SelectCode(fmt.Sprintf("%s:%s", date, university), eligible, recentCodes(university, db)), nil
}

// EligibleCodes narrows the course index to codes that can be an answer: a
// real six-character code, well established enough that a student would
// recognise it, and named so the reveal can say what it actually was.
func EligibleCodes(courses []courseservice.Course) []string {
	eligible := make([]string, 0, len(courses))

	for _, course := range courses {
		if len([]rune(course.Code)) != 6 {
			continue
		}
		if course.ExamCount < minExams {
			continue
		}
		if strings.TrimSpace(course.Name) == "" {
			continue
		}
		eligible = append(eligible, course.Code)
	}

	return eligible
}

// SelectCode picks one code, skipping anything that has aired recently.
//
// The pick is seeded by the day rather than random so two concurrent
// first-requests agree and never contend on the insert. Once the row is
// written the seed stops mattering — the stored answer is the answer.
func SelectCode(seed string, eligible []string, recent map[string]struct{}) string {
	if len(eligible) == 0 {
		return ""
	}

	candidates := make([]string, 0, len(eligible))
	for _, code := range eligible {
		if _, aired := recent[code]; !aired {
			candidates = append(candidates, code)
		}
	}
	// If everything has aired recently the archive is smaller than the window;
	// fall back to the full set rather than returning nothing.
	if len(candidates) == 0 {
		candidates = append(candidates, eligible...)
	}

	sort.Strings(candidates)

	h := fnv.New32a()
	fmt.Fprint(h, seed)

	return candidates[int(h.Sum32())%len(candidates)]
}

func recentCodes(university examservice.University, db *supabase.Client) map[string]struct{} {
	var rows []puzzleRow
	recent := map[string]struct{}{}

	_, err := db.From("daily_puzzle").
		Select("puzzle_date, university, course_code", "exact", false).
		Eq("university", string(university)).
		Order("puzzle_date", &postgrest.OrderOpts{Ascending: false}).
		Limit(repeatWindow, "").
		ExecuteTo(&rows)
	if err != nil {
		return recent
	}

	for _, row := range rows {
		recent[row.CourseCode] = struct{}{}
	}

	return recent
}

func withName(date, code string, university examservice.University, db *supabase.Client) (*Puzzle, error) {
	puzzle := &Puzzle{Date: date, CourseCode: code}

	index, err := courseservice.GetCourses(university, db)
	if err != nil {
		return puzzle, nil
	}

	for _, course := range index.Courses {
		if course.Code == code {
			puzzle.CourseName = course.Name
			break
		}
	}

	return puzzle, nil
}
