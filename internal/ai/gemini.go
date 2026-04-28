package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/resqlink-project/resqlink/internal/domain"
)

const systemInstruction = `Act as a bilingual field-data parser for a community reporting platform.
You will receive raw inputs from field workers — these may be images of infrastructure problems,
voice transcriptions in Hindi/local dialects, or text descriptions.

Your job:
1. Extract structured information from the input.
2. Translate any Hindi or regional dialect content to formal English for the database.
3. Return ONLY valid JSON with no markdown formatting.

Required output fields:
- problem_category: one of ["water", "sanitation", "road", "electricity", "housing", "health", "education", "legal", "other"]
- ward_id: extract or infer the ward/area identifier (string)
- severity_index: integer 1-10 based on urgency and impact
- affected_population_estimate: integer estimate of people affected
- summary: 2-3 sentence English summary of the problem`

// GeminiClient wraps the Vertex AI Generative Model.
type GeminiClient struct {
	model    *genai.GenerativeModel
	aiClient *genai.Client
}

// NewGeminiClient initializes a Gemini client via Vertex AI.
func NewGeminiClient(ctx context.Context, projectID, location string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("vertex ai client init: %w", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}
	model.ResponseMIMEType = "application/json"
	model.SetTemperature(0.2)
	model.SetTopK(40)
	model.SetTopP(0.95)

	return &GeminiClient{model: model, aiClient: client}, nil
}

// Close releases the underlying client resources.
func (g *GeminiClient) Close() error {
	return g.aiClient.Close()
}

// newChat creates a fresh model with a custom system instruction (no JSON mode).
func (g *GeminiClient) newFreeformModel(systemPrompt string) *genai.GenerativeModel {
	model := g.aiClient.GenerativeModel("gemini-2.0-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
	model.SetTemperature(0.3)
	model.SetTopK(40)
	model.SetTopP(0.95)
	return model
}

// newJSONModel creates a fresh model with JSON response mode.
func (g *GeminiClient) newJSONModel(systemPrompt string) *genai.GenerativeModel {
	model := g.newFreeformModel(systemPrompt)
	model.ResponseMIMEType = "application/json"
	return model
}

// extractText pulls the first text part from a Gemini response.
func extractText(resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("gemini response part is not text")
	}
	return string(text), nil
}

// ══════════════════════════════════════════════════
// FEATURE 1: Report Parsing (existing — text, image, audio)
// ══════════════════════════════════════════════════

func (g *GeminiClient) ParseTextReport(ctx context.Context, rawText string) (*domain.GeminiExtraction, error) {
	prompt := fmt.Sprintf("Parse the following field report and extract structured data:\n\n%s", rawText)
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini text generation: %w", err)
	}
	return g.extractResult(resp)
}

func (g *GeminiClient) ParseImageReport(ctx context.Context, gcsURI string, caption string) (*domain.GeminiExtraction, error) {
	img := genai.FileData{MIMEType: "image/jpeg", FileURI: gcsURI}
	prompt := "Analyze this field image and extract structured problem data."
	if caption != "" {
		prompt = fmt.Sprintf("%s Additional context from the worker: %s", prompt, caption)
	}
	resp, err := g.model.GenerateContent(ctx, img, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini image generation: %w", err)
	}
	return g.extractResult(resp)
}

func (g *GeminiClient) ParseAudioReport(ctx context.Context, gcsURI string) (*domain.GeminiExtraction, error) {
	audio := genai.FileData{MIMEType: "audio/mpeg", FileURI: gcsURI}
	prompt := "Transcribe and parse this voice report from a field worker. The audio may be in Hindi or a local dialect."
	resp, err := g.model.GenerateContent(ctx, audio, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini audio generation: %w", err)
	}
	return g.extractResult(resp)
}

func (g *GeminiClient) extractResult(resp *genai.GenerateContentResponse) (*domain.GeminiExtraction, error) {
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var extraction domain.GeminiExtraction
	if err := json.Unmarshal([]byte(text), &extraction); err != nil {
		return nil, fmt.Errorf("gemini json decode: %w (raw: %s)", err, text)
	}
	if extraction.SeverityIndex < 1 {
		extraction.SeverityIndex = 1
	}
	if extraction.SeverityIndex > 10 {
		extraction.SeverityIndex = 10
	}
	return &extraction, nil
}

// ══════════════════════════════════════════════════
// FEATURE 2: Multimodal Image Analysis
// ══════════════════════════════════════════════════

// AnalyzeImage takes a base64 image and returns structured analysis.
func (g *GeminiClient) AnalyzeImage(ctx context.Context, base64Image string) (*domain.ImageAnalysis, error) {
	model := g.newJSONModel(`You are an expert image analyst for a civic issue reporting platform.
Analyze the image and return JSON with:
- issue_type: one of ["medical_emergency", "legal_aid", "civic_issue", "disaster_relief"]
- description: detailed 2-3 sentence description of what you see
- severity: integer 1-10
- detected_objects: array of strings of key objects detected
- suggested_category: one of ["water", "sanitation", "road", "electricity", "housing", "health", "education", "legal", "other"]
- location_hints: any visible location clues (signs, landmarks)
- requires_immediate_attention: boolean`)

	imgPart := genai.Blob{MIMEType: "image/jpeg", Data: decodeBase64(base64Image)}
	resp, err := model.GenerateContent(ctx, imgPart, genai.Text("Analyze this image for civic issue reporting."))
	if err != nil {
		return nil, fmt.Errorf("image analysis: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.ImageAnalysis
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("image analysis decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 3: AI Report Verification
// ══════════════════════════════════════════════════

// VerifyReport cross-checks image against text to detect inconsistencies.
func (g *GeminiClient) VerifyReport(ctx context.Context, base64Image, reportText string) (*domain.ReportVerification, error) {
	model := g.newJSONModel(`You are a report verification system. Compare the image with the text description.
Return JSON with:
- is_consistent: boolean (true if image matches text)
- confidence: float 0-1
- discrepancies: array of strings describing any mismatches
- verification_summary: 1-2 sentence summary
- risk_score: integer 1-10 (1=genuine, 10=likely fake)`)

	imgPart := genai.Blob{MIMEType: "image/jpeg", Data: decodeBase64(base64Image)}
	prompt := fmt.Sprintf("Compare this image with the following report text and verify consistency:\n\nReport: %s", reportText)
	resp, err := model.GenerateContent(ctx, imgPart, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("verify report: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.ReportVerification
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("verify decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 4: Duplicate Report Detection
// ══════════════════════════════════════════════════

// DetectDuplicates compares a new report's text against existing reports.
func (g *GeminiClient) DetectDuplicates(ctx context.Context, newReportText string, existingReports []string) (*domain.DuplicateDetection, error) {
	model := g.newJSONModel(`You are a duplicate detection system for civic issue reports.
Compare the NEW report with existing reports and identify duplicates or similar issues.
Return JSON with:
- has_duplicates: boolean
- similar_reports: array of objects with {index: int, similarity_score: float 0-1, reason: string}
- recommendation: one of ["submit_new", "merge_with_existing", "likely_duplicate"]
- summary: 1-2 sentence explanation`)

	existingList := ""
	for i, r := range existingReports {
		existingList += fmt.Sprintf("\n[Report %d]: %s", i+1, r)
	}

	prompt := fmt.Sprintf("NEW REPORT: %s\n\nEXISTING REPORTS:%s\n\nIdentify if the new report is a duplicate or similar to any existing reports.", newReportText, existingList)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("duplicate detection: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.DuplicateDetection
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("duplicate decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 5: AI-Generated Action Plans
// ══════════════════════════════════════════════════

// GenerateActionPlan creates a step-by-step volunteer action plan.
func (g *GeminiClient) GenerateActionPlan(ctx context.Context, report *domain.Report) (*domain.ActionPlan, error) {
	model := g.newJSONModel(`You are a field operations planner. Generate an action plan for a volunteer.
Return JSON with:
- title: short title for the plan
- estimated_duration: string (e.g., "2-3 hours")
- priority_level: one of ["low", "medium", "high", "critical"]
- steps: array of objects with {step_number: int, action: string, details: string, safety_notes: string}
- required_equipment: array of strings
- safety_warnings: array of strings
- tips: array of strings for effective resolution`)

	prompt := fmt.Sprintf(`Create an action plan for a volunteer assigned to this issue:
- Type: %s
- Urgency: %s
- Description: %s
- Location: %s
- Severity: %.0f/10
- Affected Population: %d`, report.IssueType, report.UserUrgency, report.RawText, report.Location, report.SeverityIndex, report.AffectedPopulationEst)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("action plan: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.ActionPlan
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("action plan decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 6: Sentiment & Emotion Analysis
// ══════════════════════════════════════════════════

// AnalyzeSentiment evaluates the emotional tone and urgency of a report.
func (g *GeminiClient) AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error) {
	model := g.newJSONModel(`You are a sentiment and emotion analyzer for emergency reports.
Analyze the emotional content and urgency signals in the text.
Return JSON with:
- overall_sentiment: one of ["positive", "neutral", "negative", "distressed", "panicked"]
- urgency_boost: float 0-3 (additional urgency score based on emotional intensity)
- emotions_detected: array of strings (e.g., ["fear", "frustration", "desperation"])
- emotional_intensity: float 0-1
- key_phrases: array of strings that indicate emotional urgency
- recommendation: string suggesting how to handle based on emotional state`)

	prompt := fmt.Sprintf("Analyze the emotional tone and urgency signals in this report:\n\n%s", text)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("sentiment analysis: %w", err)
	}
	respText, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.SentimentAnalysis
	if err := json.Unmarshal([]byte(respText), &result); err != nil {
		return nil, fmt.Errorf("sentiment decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 7: Real-Time Translation
// ══════════════════════════════════════════════════

// TranslateMessage translates text between languages.
func (g *GeminiClient) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*domain.Translation, error) {
	model := g.newJSONModel(`You are a professional translator for a civic platform.
Return JSON with:
- translated_text: the translated text
- source_language: detected source language
- target_language: target language
- confidence: float 0-1
- cultural_notes: any cultural context that may be relevant (optional string)`)

	prompt := fmt.Sprintf("Translate the following from %s to %s:\n\n%s", sourceLang, targetLang, text)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("translation: %w", err)
	}
	respText, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.Translation
	if err := json.Unmarshal([]byte(respText), &result); err != nil {
		return nil, fmt.Errorf("translation decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 8: AI Progress Report
// ══════════════════════════════════════════════════

// GenerateProgressReport creates an AI-powered weekly summary for NGOs.
func (g *GeminiClient) GenerateProgressReport(ctx context.Context, reports []*domain.Report) (*domain.ProgressReport, error) {
	model := g.newJSONModel(`You are an NGO operations analyst. Generate a weekly progress report.
Return JSON with:
- title: "Weekly Progress Report"
- date_range: string
- executive_summary: 3-4 sentence overview
- total_issues: int
- resolved_issues: int
- pending_issues: int
- critical_issues: int
- resolution_rate: float (percentage)
- top_categories: array of objects {category: string, count: int, trend: "up"/"down"/"stable"}
- hotspot_areas: array of strings (top affected locations)
- key_achievements: array of strings
- areas_of_concern: array of strings
- recommendations: array of strings for next week`)

	// Compile report data
	var summaries []string
	for _, r := range reports {
		status := r.Status
		if status == "" {
			status = "pending"
		}
		summaries = append(summaries, fmt.Sprintf("[%s|%s|%s|%s] %s", r.IssueType, r.UserUrgency, status, r.Location, r.RawText))
	}
	prompt := fmt.Sprintf("Generate a weekly progress report based on these %d issues:\n\n%s", len(reports), strings.Join(summaries, "\n"))
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("progress report: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.ProgressReport
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("progress report decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 9: Volunteer Skill Recommendations
// ══════════════════════════════════════════════════

// RecommendSkills suggests new skills for volunteers based on task history.
func (g *GeminiClient) RecommendSkills(ctx context.Context, currentSkills []string, completedTaskTypes []string) (*domain.SkillRecommendation, error) {
	model := g.newJSONModel(`You are a volunteer development advisor.
Based on a volunteer's current skills and completed task history, suggest new skills.
Return JSON with:
- recommended_skills: array of objects {skill: string, reason: string, priority: "high"/"medium"/"low"}
- training_suggestions: array of strings
- career_path: string describing potential growth direction
- strengths: array of strings based on completed tasks`)

	prompt := fmt.Sprintf("Current skills: %v\nCompleted task types: %v\n\nRecommend new skills for this volunteer.", currentSkills, completedTaskTypes)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("skill recommendation: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.SkillRecommendation
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("skill recommendation decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 10: Document OCR
// ══════════════════════════════════════════════════

// OCRDocument extracts text from a document image.
func (g *GeminiClient) OCRDocument(ctx context.Context, base64Image string) (*domain.OCRResult, error) {
	model := g.newJSONModel(`You are an expert document reader (OCR). Extract all text from the document image.
Return JSON with:
- extracted_text: full text content of the document
- document_type: one of ["legal", "medical", "government", "general", "id_card", "receipt"]
- language: detected language
- key_fields: array of objects {field_name: string, field_value: string} for structured data
- confidence: float 0-1
- summary: 1-2 sentence summary of the document`)

	imgPart := genai.Blob{MIMEType: "image/jpeg", Data: decodeBase64(base64Image)}
	resp, err := model.GenerateContent(ctx, imgPart, genai.Text("Extract all text and structured data from this document image."))
	if err != nil {
		return nil, fmt.Errorf("ocr: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return nil, err
	}
	var result domain.OCRResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("ocr decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 11: AI Chatbot
// ══════════════════════════════════════════════════

// ChatWithAI provides context-aware conversation for volunteers.
func (g *GeminiClient) ChatWithAI(ctx context.Context, message string, taskContext string) (string, error) {
	model := g.newFreeformModel(`You are RESQLINK AI Assistant — a helpful, friendly assistant for volunteers
working on civic issues, medical emergencies, legal aid, and disaster relief.
You provide practical, actionable guidance. Keep responses concise (2-4 paragraphs max).
If asked about safety, always err on the side of caution.
You can speak in Hindi or English based on the user's language.`)

	prompt := message
	if taskContext != "" {
		prompt = fmt.Sprintf("Context about my current assigned task:\n%s\n\nMy question: %s", taskContext, message)
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}
	text, err := extractText(resp)
	if err != nil {
		return "", err
	}
	return text, nil
}

// ══════════════════════════════════════════════════
// Helper
// ══════════════════════════════════════════════════

func decodeBase64(b64 string) []byte {
	// Strip data URI prefix if present
	if idx := strings.Index(b64, ","); idx != -1 {
		b64 = b64[idx+1:]
	}
	data, _ := base64.StdEncoding.DecodeString(b64)
	return data
}
