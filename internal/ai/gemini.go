package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/resqlink-project/resqlink/internal/domain"
)

// GeminiClient is an offline-friendly AI adapter.
// In production, this can be swapped for a real Vertex AI client.
type GeminiClient struct{}

func NewGeminiClient(ctx context.Context, projectID, location string) (*GeminiClient, error) {
	return &GeminiClient{}, nil
}

func (g *GeminiClient) Close() error { return nil }

func inferIssue(text string) (string, string) {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "hospital"), strings.Contains(lower, "injury"), strings.Contains(lower, "medical"):
		return "health", "medical_emergency"
	case strings.Contains(lower, "law"), strings.Contains(lower, "court"), strings.Contains(lower, "legal"):
		return "legal", "legal_aid"
	case strings.Contains(lower, "flood"), strings.Contains(lower, "earthquake"), strings.Contains(lower, "disaster"):
		return "disaster", "disaster_relief"
	case strings.Contains(lower, "road"), strings.Contains(lower, "bridge"), strings.Contains(lower, "pothole"):
		return "road", "civic_issue"
	default:
		return "other", "civic_issue"
	}
}

func (g *GeminiClient) ParseTextReport(ctx context.Context, rawText string) (*domain.GeminiExtraction, error) {
	category, _ := inferIssue(rawText)
	return &domain.GeminiExtraction{
		ProblemCategory:       category,
		WardID:                "unknown",
		SeverityIndex:         5,
		AffectedPopulationEst: 10,
		Summary:               strings.TrimSpace(rawText),
	}, nil
}

func (g *GeminiClient) ParseImageReport(ctx context.Context, gcsURI string, caption string) (*domain.GeminiExtraction, error) {
	return g.ParseTextReport(ctx, caption)
}

func (g *GeminiClient) ParseAudioReport(ctx context.Context, gcsURI string) (*domain.GeminiExtraction, error) {
	return &domain.GeminiExtraction{ProblemCategory: "other", WardID: "unknown", SeverityIndex: 5, AffectedPopulationEst: 10, Summary: "Audio report received"}, nil
}

func (g *GeminiClient) AnalyzeImage(ctx context.Context, base64Image string) (*domain.ImageAnalysis, error) {
	return &domain.ImageAnalysis{Description: "Offline image analysis unavailable", DetectedIssueType: "civic_issue", SeverityHint: "medium", Tags: []string{"offline"}}, nil
}

func (g *GeminiClient) VerifyReport(ctx context.Context, base64Image, reportText string) (*domain.ReportVerification, error) {
	return &domain.ReportVerification{IsConsistent: true, Confidence: 0.5, VerificationSummary: "Offline verification fallback", RiskScore: 5}, nil
}

func (g *GeminiClient) DetectDuplicates(ctx context.Context, newReportText string, existingReports []string) (*domain.DuplicateDetection, error) {
	return &domain.DuplicateDetection{HasDuplicates: false, Recommendation: "submit_new", Summary: "Offline duplicate detection fallback"}, nil
}

func (g *GeminiClient) GenerateActionPlan(ctx context.Context, report *domain.Report) (*domain.ActionPlan, error) {
	return &domain.ActionPlan{Title: "Respond to issue", EstimatedDuration: "1-2 hours", PriorityLevel: "medium", Steps: []domain.ActionStep{{StepNumber: 1, Action: "Assess", Details: report.Summary, SafetyNotes: "Use caution"}}}, nil
}

func (g *GeminiClient) AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error) {
	return &domain.SentimentAnalysis{OverallSentiment: "neutral", Recommendation: "Proceed with triage"}, nil
}

func (g *GeminiClient) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*domain.Translation, error) {
	return &domain.Translation{TranslatedText: text, SourceLanguage: sourceLang, TargetLanguage: targetLang, Confidence: 0.5}, nil
}

func (g *GeminiClient) GenerateProgressReport(ctx context.Context, reports []*domain.Report) (*domain.ProgressReport, error) {
	var resolved, pending, escalated int
	for _, r := range reports {
		switch r.Status {
		case domain.StatusResolved:
			resolved++
		case domain.StatusEscalated:
			escalated++
		default:
			pending++
		}
	}
	return &domain.ProgressReport{Title: "Progress Report", TotalIssues: len(reports), ResolvedIssues: resolved, PendingIssues: pending, CriticalIssues: escalated, ExecutiveSummary: "Offline summary"}, nil
}

func (g *GeminiClient) RecommendSkills(ctx context.Context, currentSkills []string, completedTaskTypes []string) (*domain.SkillRecommendation, error) {
	return &domain.SkillRecommendation{CareerPath: "community responder", Strengths: currentSkills}, nil
}

func (g *GeminiClient) OCRDocument(ctx context.Context, base64Image string) (*domain.OCRResult, error) {
	return &domain.OCRResult{ExtractedText: "OCR unavailable in offline mode", DocumentType: "unknown", Language: "en"}, nil
}

func (g *GeminiClient) ChatWithAI(ctx context.Context, message string, taskContext string) (string, error) {
	return fmt.Sprintf("Offline AI fallback: %s", message), nil
}
