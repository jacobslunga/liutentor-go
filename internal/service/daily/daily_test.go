package daily

import (
	"testing"

	courseservice "liutentor-go/internal/service/course"
)

func TestEligibleCodes(t *testing.T) {
	courses := []courseservice.Course{
		{Code: "TATA43", Name: "Flervariabelanalys", ExamCount: 65},
		{Code: "TDDC88", Name: "Programutveckling", ExamCount: 49},
		{Code: "TDDX01", Name: "Sällan given kurs", ExamCount: 3}, // too few exams
		{Code: "TDDX02", Name: "", ExamCount: 40},                 // unnamed
		{Code: "TDDX03", Name: "   ", ExamCount: 40},              // blank name
		{Code: "ABCDEFG", Name: "För lång", ExamCount: 40},        // wrong length
		{Code: "9ÄEGEX", Name: "Konstig kod", ExamCount: 40},      // six runes, kept
	}

	got := EligibleCodes(courses)
	want := []string{"TATA43", "TDDC88", "9ÄEGEX"}

	if len(got) != len(want) {
		t.Fatalf("EligibleCodes() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("EligibleCodes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEligibleCodesEmpty(t *testing.T) {
	if got := EligibleCodes(nil); len(got) != 0 {
		t.Errorf("EligibleCodes(nil) = %v, want empty", got)
	}
}

func TestSelectCodeIsDeterministic(t *testing.T) {
	eligible := []string{"TATA43", "TDDC88", "TDTS10", "TAOP88", "732A75"}

	first := SelectCode("2026-08-21:LIU", eligible, nil)
	for i := 0; i < 20; i++ {
		if got := SelectCode("2026-08-21:LIU", eligible, nil); got != first {
			t.Fatalf("SelectCode() returned %q then %q for the same seed", first, got)
		}
	}
}

func TestSelectCodeIgnoresInputOrder(t *testing.T) {
	forward := []string{"TATA43", "TDDC88", "TDTS10", "TAOP88", "732A75"}
	reversed := []string{"732A75", "TAOP88", "TDTS10", "TDDC88", "TATA43"}

	if a, b := SelectCode("seed", forward, nil), SelectCode("seed", reversed, nil); a != b {
		t.Errorf("SelectCode() = %q for one order and %q for another", a, b)
	}
}

func TestSelectCodeSkipsRecent(t *testing.T) {
	eligible := []string{"TATA43", "TDDC88", "TDTS10", "TAOP88", "732A75"}

	// Rule out every code but one; that one must be the pick.
	recent := map[string]struct{}{
		"TATA43": {}, "TDDC88": {}, "TDTS10": {}, "TAOP88": {},
	}

	if got := SelectCode("2026-08-21:LIU", eligible, recent); got != "732A75" {
		t.Errorf("SelectCode() = %q, want the only code not aired recently", got)
	}
}

func TestSelectCodeFallsBackWhenAllRecent(t *testing.T) {
	// A pool smaller than the repeat window means everything has aired. Returning
	// nothing would leave the day with no puzzle at all.
	eligible := []string{"TATA43", "TDDC88"}
	recent := map[string]struct{}{"TATA43": {}, "TDDC88": {}}

	got := SelectCode("2026-08-21:LIU", eligible, recent)
	if got != "TATA43" && got != "TDDC88" {
		t.Errorf("SelectCode() = %q, want one of the eligible codes", got)
	}
}

func TestSelectCodeEmpty(t *testing.T) {
	if got := SelectCode("seed", nil, nil); got != "" {
		t.Errorf("SelectCode(nil) = %q, want empty", got)
	}
}

func TestSelectCodeVariesByDay(t *testing.T) {
	eligible := []string{
		"TATA43", "TDDC88", "TDTS10", "TAOP88", "732A75",
		"TATA41", "TDDD37", "TMMV11", "TDDC90", "TSEA22",
	}

	seen := map[string]struct{}{}
	for _, day := range []string{
		"2026-08-21", "2026-08-22", "2026-08-23", "2026-08-24", "2026-08-25",
	} {
		seen[SelectCode(day+":LIU", eligible, nil)] = struct{}{}
	}

	// Not a guarantee of the algorithm, but a stuck seed would show up here.
	if len(seen) < 2 {
		t.Errorf("SelectCode() produced %d distinct codes across 5 days, want >= 2", len(seen))
	}
}

func TestTodayIsISODate(t *testing.T) {
	got := Today()
	if len(got) != 10 || got[4] != '-' || got[7] != '-' {
		t.Errorf("Today() = %q, want YYYY-MM-DD", got)
	}
}
