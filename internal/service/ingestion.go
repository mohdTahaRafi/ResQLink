package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/samaj-project/samaj/internal/ai"
	"github.com/samaj-project/samaj/internal/domain"
	"github.com/samaj-project/samaj/internal/repository"
)

// IngestionService orchestrates the AI extraction and urgency scoring pipeline.
type IngestionService struct {
	repo   *repository.FirestoreRepo
	gemini *ai.GeminiClient
}

// NewIngestionService creates an ingestion pipeline service.
func NewIngestionService(repo *repository.FirestoreRepo, gemini *ai.GeminiClient) *IngestionService {
	return &IngestionService{repo: repo, gemini: gemini}
}

// ProcessReport handles a Pub/Sub ingestion event end-to-end.
func (s *IngestionService) ProcessReport(ctx context.Context, event domain.IngestionEvent) error {
	log.Printf("[ingestion] processing report %s (type: %s)", event.ReportID, event.MediaType)

	extraction, err := s.extractWithGemini(ctx, event)
	if err != nil {
		return fmt.Errorf("ai extraction: %w", err)
	}

	log.Printf("[ingestion] extracted: category=%s ward=%s severity=%.0f",
		extraction.ProblemCategory, extraction.WardID, extraction.SeverityIndex)

	ward, err := s.repo.GetWard(ctx, extraction.WardID)
	if err != nil {
		log.Printf("[ingestion] ward lookup failed for %s, using defaults: %v", extraction.WardID, err)
		ward = &domain.Ward{ID: extraction.WardID, PopulationDensity: 5000}
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	recentReports, err := s.repo.GetReportsByWard(ctx, extraction.WardID, sevenDaysAgo)
	if err != nil {
		return fmt.Errorf("fetch recent reports: %w", err)
	}

	var oldestUnresolved time.Time
	var lastReportTime time.Time
	unresolved, err := s.repo.GetUnresolvedReportsByWard(ctx, extraction.WardID)
	if err == nil && len(unresolved) > 0 {
		oldestUnresolved = unresolved[len(unresolved)-1].CreatedAt
	}
	if len(recentReports) > 0 {
		lastReportTime = recentReports[0].CreatedAt
	}

	params := BuildUrgencyParams(recentReports, extraction.SeverityIndex, ward, oldestUnresolved, lastReportTime)
	urgencyScore := ComputeUrgency(params)

	log.Printf("[ingestion] urgency score: %.2f (F=%.0f, S=%.0f)", urgencyScore, params.Frequency, params.Severity)

	updates := []firestore.Update{
		{Path: "problem_category", Value: extraction.ProblemCategory},
		{Path: "ward_id", Value: extraction.WardID},
		{Path: "severity_index", Value: extraction.SeverityIndex},
		{Path: "affected_population_estimate", Value: extraction.AffectedPopulationEst},
		{Path: "summary", Value: extraction.Summary},
		{Path: "urgency_score", Value: urgencyScore},
	}

	if err := s.repo.UpdateReport(ctx, event.ReportID, updates); err != nil {
		return fmt.Errorf("update report: %w", err)
	}

	log.Printf("[ingestion] report %s enriched successfully", event.ReportID)
	return nil
}

func (s *IngestionService) extractWithGemini(ctx context.Context, event domain.IngestionEvent) (*domain.GeminiExtraction, error) {
	switch event.MediaType {
	case "image":
		return s.gemini.ParseImageReport(ctx, event.MediaURL, event.RawText)
	case "audio":
		return s.gemini.ParseAudioReport(ctx, event.MediaURL)
	case "text":
		return s.gemini.ParseTextReport(ctx, event.RawText)
	default:
		if event.RawText != "" {
			return s.gemini.ParseTextReport(ctx, event.RawText)
		}
		return nil, fmt.Errorf("unsupported media type: %s", event.MediaType)
	}
}

// SerializeIngestionEvent encodes an event for Pub/Sub.
func SerializeIngestionEvent(event domain.IngestionEvent) ([]byte, error) {
	return json.Marshal(event)
}

// DeserializeIngestionEvent decodes a Pub/Sub message into an event.
func DeserializeIngestionEvent(data []byte) (domain.IngestionEvent, error) {
	var event domain.IngestionEvent
	err := json.Unmarshal(data, &event)
	return event, err
}
