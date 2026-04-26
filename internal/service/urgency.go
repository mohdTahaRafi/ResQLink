package service

import (
	"math"
	"time"

	"github.com/samaj-project/samaj/internal/domain"
)

// Weights for the urgency formula.
const (
	wFrequency = 1.5
	wSeverity  = 2.0
	wPopImpact = 1.2
	wLagBonus  = 0.8
	wTimeDecay = 0.5

	lagThresholdHours = 48
)

// ComputeUrgency implements: U = (F·1.5) + (S·2.0) + (P_impact·1.2) + (L_bonus·0.8) - (T_decay·0.5)
func ComputeUrgency(params domain.UrgencyParams) float64 {
	score := (params.Frequency * wFrequency) +
		(params.Severity * wSeverity) +
		(params.PopImpact * wPopImpact) +
		(params.LagBonus * wLagBonus) -
		(params.TimeSinceHours * wTimeDecay)

	if score < 0 {
		score = 0
	}
	return math.Round(score*100) / 100
}

// BuildUrgencyParams constructs parameters from report context.
func BuildUrgencyParams(
	reportsInWindow []*domain.Report,
	severity float64,
	ward *domain.Ward,
	oldestUnresolved time.Time,
	lastReportTime time.Time,
) domain.UrgencyParams {
	frequency := float64(len(reportsInWindow))
	popImpact := ward.PopulationDensity / 1000.0

	var lagBonus float64
	if !oldestUnresolved.IsZero() {
		hoursUnresolved := time.Since(oldestUnresolved).Hours()
		if hoursUnresolved > lagThresholdHours {
			lagBonus = math.Log2(hoursUnresolved / lagThresholdHours)
			if lagBonus > 5.0 {
				lagBonus = 5.0
			}
		}
	}

	timeSince := time.Since(lastReportTime).Hours()
	if timeSince < 0 {
		timeSince = 0
	}

	return domain.UrgencyParams{
		Frequency:      frequency,
		Severity:       severity,
		PopImpact:      popImpact,
		LagBonus:       lagBonus,
		TimeSinceHours: timeSince,
	}
}
