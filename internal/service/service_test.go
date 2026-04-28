package service

import (
	"math"
	"testing"

	"github.com/resqlink-project/resqlink/internal/domain"
)

func TestComputeUrgency_BasicCase(t *testing.T) {
	params := domain.UrgencyParams{
		Frequency:      5,
		Severity:       8,
		PopImpact:      3.0,
		LagBonus:       1.5,
		TimeSinceHours: 2.0,
	}

	score := ComputeUrgency(params)

	// U = (5*1.5) + (8*2.0) + (3.0*1.2) + (1.5*0.8) - (2.0*0.5)
	// U = 7.5 + 16.0 + 3.6 + 1.2 - 1.0 = 27.3
	expected := 27.3
	if math.Abs(score-expected) > 0.01 {
		t.Errorf("expected %.2f, got %.2f", expected, score)
	}
}

func TestComputeUrgency_ZeroFloor(t *testing.T) {
	params := domain.UrgencyParams{
		Frequency:      0,
		Severity:       1,
		PopImpact:      0.1,
		LagBonus:       0,
		TimeSinceHours: 100,
	}

	score := ComputeUrgency(params)
	if score < 0 {
		t.Errorf("urgency score should never be negative, got %.2f", score)
	}
}

func TestComputeUrgency_HighSeverity(t *testing.T) {
	low := ComputeUrgency(domain.UrgencyParams{Severity: 2, Frequency: 1})
	high := ComputeUrgency(domain.UrgencyParams{Severity: 9, Frequency: 1})

	if high <= low {
		t.Errorf("higher severity should yield higher urgency: low=%.2f high=%.2f", low, high)
	}
}

func TestMatchVolunteers_Concurrent(t *testing.T) {
	volunteers := make([]*domain.Volunteer, 50)
	for i := range volunteers {
		volunteers[i] = &domain.Volunteer{
			ID:             "v" + string(rune('0'+i%10)),
			Name:           "Volunteer",
			Skills:         []string{"water", "sanitation"},
			Latitude:       26.8 + float64(i)*0.01,
			Longitude:      80.9 + float64(i)*0.01,
			CompletionRate: 0.5 + float64(i%5)*0.1,
		}
	}

	results := MatchVolunteers(
		[]string{"water", "health"},
		26.85, 80.95,
		volunteers,
	)

	if len(results) != 50 {
		t.Errorf("expected 50 results, got %d", len(results))
	}

	for i := 1; i < len(results); i++ {
		if results[i].TotalScore > results[i-1].TotalScore {
			t.Errorf("results not sorted descending at index %d", i)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected float64
	}{
		{"identical", []string{"a", "b"}, []string{"a", "b"}, 1.0},
		{"disjoint", []string{"a", "b"}, []string{"c", "d"}, 0.0},
		{"partial", []string{"a", "b", "c"}, []string{"a", "b"}, 0.816},
		{"empty_a", []string{}, []string{"a"}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := cosineSimilarity(tt.a, tt.b)
			if math.Abs(score-tt.expected) > 0.01 {
				t.Errorf("expected ~%.3f, got %.3f", tt.expected, score)
			}
		})
	}
}

func TestDistanceScore(t *testing.T) {
	same := distanceScore(26.85, 80.95, 26.85, 80.95)
	if math.Abs(same-1.0) > 0.001 {
		t.Errorf("same location should give score ~1.0, got %.3f", same)
	}

	far := distanceScore(26.85, 80.95, 28.60, 77.20)
	if far >= same {
		t.Errorf("far distance should give lower score than same: same=%.3f far=%.3f", same, far)
	}
}
