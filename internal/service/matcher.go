package service

import (
	"math"
	"sort"
	"sync"

	"github.com/resqlink-project/resqlink/internal/domain"
)

// Match weights.
const (
	wSkill       = 0.4
	wDistance     = 0.35
	wReliability = 0.25
)

// MatchVolunteers concurrently scores and ranks volunteers for a task.
func MatchVolunteers(
	taskSkills []string,
	taskLat, taskLng float64,
	volunteers []*domain.Volunteer,
) []domain.MatchResult {
	results := make([]domain.MatchResult, len(volunteers))
	var wg sync.WaitGroup

	for i, v := range volunteers {
		wg.Add(1)
		go func(idx int, vol *domain.Volunteer) {
			defer wg.Done()

			skill := cosineSimilarity(taskSkills, vol.Skills)
			dist := distanceScore(taskLat, taskLng, vol.Latitude, vol.Longitude)
			reliability := vol.CompletionRate

			total := (skill * wSkill) + (dist * wDistance) + (reliability * wReliability)

			results[idx] = domain.MatchResult{
				VolunteerID:   vol.ID,
				VolunteerName: vol.Name,
				SkillScore:    math.Round(skill*1000) / 1000,
				DistanceScore: math.Round(dist*1000) / 1000,
				Reliability:   math.Round(reliability*1000) / 1000,
				TotalScore:    math.Round(total*1000) / 1000,
			}
		}(i, v)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results
}

// cosineSimilarity computes similarity between two tag sets.
func cosineSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	universe := make(map[string]int)
	idx := 0
	for _, s := range a {
		if _, exists := universe[s]; !exists {
			universe[s] = idx
			idx++
		}
	}
	for _, s := range b {
		if _, exists := universe[s]; !exists {
			universe[s] = idx
			idx++
		}
	}

	vecA := make([]float64, len(universe))
	vecB := make([]float64, len(universe))

	for _, s := range a {
		vecA[universe[s]] = 1
	}
	for _, s := range b {
		vecB[universe[s]] = 1
	}

	var dotProduct, normA, normB float64
	for i := range vecA {
		dotProduct += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// distanceScore computes D = 1 / (1 + HaversineDistance).
func distanceScore(lat1, lng1, lat2, lng2 float64) float64 {
	dist := haversine(lat1, lng1, lat2, lng2)
	return 1.0 / (1.0 + dist)
}

// haversine computes great-circle distance in kilometers.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
