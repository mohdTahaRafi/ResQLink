package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/samaj-project/samaj/internal/domain"
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
	model      *genai.GenerativeModel
	aiClient   *genai.Client
	apiKey     string
	restModel  string
	httpClient *http.Client
}

// NewGeminiClient initializes a Gemini client via Vertex AI.
func NewGeminiClient(ctx context.Context, projectID, location string) (*GeminiClient, error) {
	apiKey := firstNonEmpty(
		os.Getenv("GEMINI_API_KEY"),
		os.Getenv("GOOGLE_API_KEY"),
		os.Getenv("GOOGLE_GENERATIVE_AI_API_KEY"),
		os.Getenv("GOOGLE_MAPS_API_KEY"),
	)

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		if apiKey == "" {
			return nil, fmt.Errorf("vertex ai client init: %w", err)
		}
		return &GeminiClient{
			apiKey:     apiKey,
			restModel:  "gemini-2.0-flash",
			httpClient: &http.Client{Timeout: 60 * time.Second},
		}, nil
	}

	model := client.GenerativeModel("gemini-2.0-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}
	model.ResponseMIMEType = "application/json"
	model.SetTemperature(0.2)
	model.SetTopK(40)
	model.SetTopP(0.95)

	return &GeminiClient{
		model:      model,
		aiClient:   client,
		apiKey:     apiKey,
		restModel:  "gemini-2.0-flash",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Close releases the underlying client resources.
func (g *GeminiClient) Close() error {
	if g.aiClient == nil {
		return nil
	}
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

func (g *GeminiClient) generateText(ctx context.Context, systemPrompt, prompt string, jsonMode bool) (string, error) {
	var restErr error
	if g.apiKey != "" {
		text, err := g.generateREST(ctx, systemPrompt, prompt, nil, jsonMode)
		if err == nil {
			return text, nil
		}
		restErr = err
	}

	if g.aiClient == nil {
		return "", fmt.Errorf("gemini REST fallback failed: %w", restErr)
	}
	var model *genai.GenerativeModel
	if jsonMode {
		model = g.newJSONModel(systemPrompt)
	} else {
		model = g.newFreeformModel(systemPrompt)
	}
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		if restErr != nil {
			return "", fmt.Errorf("gemini REST failed: %v; vertex failed: %w", restErr, err)
		}
		return "", err
	}
	return extractText(resp)
}

func (g *GeminiClient) generateImageText(ctx context.Context, systemPrompt, base64Image, prompt string, jsonMode bool) (string, error) {
	imageBytes, mimeType, err := decodeBase64Image(base64Image)
	if err != nil {
		return "", err
	}

	var restErr error
	if g.apiKey != "" {
		text, err := g.generateREST(ctx, systemPrompt, prompt, &restInlineData{
			MIMEType: mimeType,
			Data:     base64.StdEncoding.EncodeToString(imageBytes),
		}, jsonMode)
		if err == nil {
			return text, nil
		}
		restErr = err
	}

	if g.aiClient == nil {
		return "", fmt.Errorf("gemini REST fallback failed: %w", restErr)
	}
	var model *genai.GenerativeModel
	if jsonMode {
		model = g.newJSONModel(systemPrompt)
	} else {
		model = g.newFreeformModel(systemPrompt)
	}
	resp, err := model.GenerateContent(ctx, genai.Blob{MIMEType: mimeType, Data: imageBytes}, genai.Text(prompt))
	if err != nil {
		if restErr != nil {
			return "", fmt.Errorf("gemini REST failed: %v; vertex failed: %w", restErr, err)
		}
		return "", err
	}
	return extractText(resp)
}

type restInlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type restPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *restInlineData `json:"inlineData,omitempty"`
}

type restContent struct {
	Role  string     `json:"role,omitempty"`
	Parts []restPart `json:"parts"`
}

func (g *GeminiClient) generateREST(ctx context.Context, systemPrompt, prompt string, image *restInlineData, jsonMode bool) (string, error) {
	parts := []restPart{{Text: prompt}}
	if image != nil {
		parts = append([]restPart{{InlineData: image}}, parts...)
	}

	reqBody := map[string]any{
		"systemInstruction": restContent{Parts: []restPart{{Text: systemPrompt}}},
		"contents":          []restContent{{Role: "user", Parts: parts}},
		"generationConfig": map[string]any{
			"temperature": 0.2,
			"topK":        40,
			"topP":        0.95,
		},
	}
	if jsonMode {
		reqBody["generationConfig"].(map[string]any)["responseMimeType"] = "application/json"
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.restModel, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := g.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if result.Error != nil {
			return "", fmt.Errorf("gemini developer API %s: %s", result.Error.Status, result.Error.Message)
		}
		return "", fmt.Errorf("gemini developer API HTTP %d", resp.StatusCode)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini developer API returned empty response")
	}
	text := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	if text == "" {
		return "", fmt.Errorf("gemini developer API returned empty text")
	}
	return text, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// ══════════════════════════════════════════════════
// FEATURE 1: Report Parsing (existing — text, image, audio)
// ══════════════════════════════════════════════════

func (g *GeminiClient) ParseTextReport(ctx context.Context, rawText string) (*domain.GeminiExtraction, error) {
	prompt := fmt.Sprintf("Parse the following field report and extract structured data:\n\n%s", rawText)
	text, err := g.generateText(ctx, systemInstruction, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("gemini text generation: %w", err)
	}
	return g.decodeExtraction(text)
}

func (g *GeminiClient) ParseImageReport(ctx context.Context, gcsURI string, caption string) (*domain.GeminiExtraction, error) {
	prompt := "Analyze this field image and extract structured problem data."
	if caption != "" {
		prompt = fmt.Sprintf("%s Additional context from the worker: %s", prompt, caption)
	}
	if strings.HasPrefix(gcsURI, "data:image") {
		text, err := g.generateImageText(ctx, systemInstruction, gcsURI, prompt, true)
		if err != nil {
			return nil, fmt.Errorf("gemini image generation: %w", err)
		}
		return g.decodeExtraction(text)
	}
	if g.aiClient == nil {
		return nil, fmt.Errorf("image report requires Vertex AI for file URI input")
	}
	img := genai.FileData{MIMEType: "image/jpeg", FileURI: gcsURI}
	resp, err := g.model.GenerateContent(ctx, img, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini image generation: %w", err)
	}
	return g.extractResult(resp)
}

func (g *GeminiClient) ParseAudioReport(ctx context.Context, gcsURI string) (*domain.GeminiExtraction, error) {
	if g.aiClient == nil {
		return nil, fmt.Errorf("audio report requires Vertex AI file URI input")
	}
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
	return g.decodeExtraction(text)
}

func (g *GeminiClient) decodeExtraction(text string) (*domain.GeminiExtraction, error) {
	var extraction domain.GeminiExtraction
	if err := decodeJSON(text, &extraction); err != nil {
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
	systemPrompt := `You are an expert image analyst for a civic issue reporting platform.
Analyze the image and return JSON with:
- issue_type: one of ["medical_emergency", "legal_aid", "civic_issue", "disaster_relief"]
- description: detailed 2-3 sentence description of what you see
- severity: integer 1-10
- detected_objects: array of strings of key objects detected
- suggested_category: one of ["water", "sanitation", "road", "electricity", "housing", "health", "education", "legal", "other"]
- location_hints: any visible location clues (signs, landmarks)
- requires_immediate_attention: boolean`
	text, err := g.generateImageText(ctx, systemPrompt, base64Image, "Analyze this image for civic issue reporting.", true)
	if err != nil {
		return nil, fmt.Errorf("image analysis: %w", err)
	}
	var result domain.ImageAnalysis
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("image analysis decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 3: AI Report Verification
// ══════════════════════════════════════════════════

// VerifyReport cross-checks image against text to detect inconsistencies.
func (g *GeminiClient) VerifyReport(ctx context.Context, base64Image, reportText string) (*domain.ReportVerification, error) {
	systemPrompt := `You are a report verification system. Compare the image with the text description.
Return JSON with:
- is_consistent: boolean (true if image matches text)
- confidence: float 0-1
- discrepancies: array of strings describing any mismatches
- verification_summary: 1-2 sentence summary
- risk_score: integer 1-10 (1=genuine, 10=likely fake)`
	prompt := fmt.Sprintf("Compare this image with the following report text and verify consistency:\n\nReport: %s", reportText)
	text, err := g.generateImageText(ctx, systemPrompt, base64Image, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("verify report: %w", err)
	}
	var result domain.ReportVerification
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("verify decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 4: Duplicate Report Detection
// ══════════════════════════════════════════════════

// DetectDuplicates compares a new report's text against existing reports.
func (g *GeminiClient) DetectDuplicates(ctx context.Context, newReportText string, existingReports []string) (*domain.DuplicateDetection, error) {
	systemPrompt := `You are a duplicate detection system for civic issue reports.
Compare the NEW report with existing reports and identify duplicates or similar issues.
Return JSON with:
- has_duplicates: boolean
- similar_reports: array of objects with {index: int, similarity_score: float 0-1, reason: string}
- recommendation: one of ["submit_new", "merge_with_existing", "likely_duplicate"]
- summary: 1-2 sentence explanation`

	existingList := ""
	for i, r := range existingReports {
		existingList += fmt.Sprintf("\n[Report %d]: %s", i+1, r)
	}

	prompt := fmt.Sprintf("NEW REPORT: %s\n\nEXISTING REPORTS:%s\n\nIdentify if the new report is a duplicate or similar to any existing reports.", newReportText, existingList)
	text, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("duplicate detection: %w", err)
	}
	var result domain.DuplicateDetection
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("duplicate decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 5: AI-Generated Action Plans
// ══════════════════════════════════════════════════

// GenerateActionPlan creates a step-by-step volunteer action plan.
func (g *GeminiClient) GenerateActionPlan(ctx context.Context, report *domain.Report) (*domain.ActionPlan, error) {
	systemPrompt := `You are a field operations planner. Generate an action plan for a volunteer.
Return JSON with:
- title: short title for the plan
- estimated_duration: string (e.g., "2-3 hours")
- priority_level: one of ["low", "medium", "high", "critical"]
- steps: array of objects with {step_number: int, action: string, details: string, safety_notes: string}
- required_equipment: array of strings
- safety_warnings: array of strings
- tips: array of strings for effective resolution`

	prompt := fmt.Sprintf(`Create an action plan for a volunteer assigned to this issue:
- Type: %s
- Urgency: %s
- Description: %s
- Location: %s
- Severity: %.0f/10
- Affected Population: %d`, report.IssueType, report.UserUrgency, report.RawText, report.Location, report.SeverityIndex, report.AffectedPopulationEst)

	text, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("action plan: %w", err)
	}
	var result domain.ActionPlan
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("action plan decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 6: Sentiment & Emotion Analysis
// ══════════════════════════════════════════════════

// AnalyzeSentiment evaluates the emotional tone and urgency of a report.
func (g *GeminiClient) AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error) {
	systemPrompt := `You are a sentiment and emotion analyzer for emergency reports.
Analyze the emotional content and urgency signals in the text.
Return JSON with:
- overall_sentiment: one of ["positive", "neutral", "negative", "distressed", "panicked"]
- urgency_boost: float 0-3 (additional urgency score based on emotional intensity)
- emotions_detected: array of strings (e.g., ["fear", "frustration", "desperation"])
- emotional_intensity: float 0-1
- key_phrases: array of strings that indicate emotional urgency
- recommendation: string suggesting how to handle based on emotional state`

	prompt := fmt.Sprintf("Analyze the emotional tone and urgency signals in this report:\n\n%s", text)
	respText, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("sentiment analysis: %w", err)
	}
	var result domain.SentimentAnalysis
	if err := decodeJSON(respText, &result); err != nil {
		return nil, fmt.Errorf("sentiment decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 7: Real-Time Translation
// ══════════════════════════════════════════════════

// TranslateMessage translates text between languages.
func (g *GeminiClient) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*domain.Translation, error) {
	systemPrompt := `You are a professional translator for a civic platform.
Return JSON with:
- translated_text: the translated text
- source_language: detected source language
- target_language: target language
- confidence: float 0-1
- cultural_notes: any cultural context that may be relevant (optional string)`

	prompt := fmt.Sprintf("Translate the following from %s to %s:\n\n%s", sourceLang, targetLang, text)
	respText, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("translation: %w", err)
	}
	var result domain.Translation
	if err := decodeJSON(respText, &result); err != nil {
		return nil, fmt.Errorf("translation decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 8: AI Progress Report
// ══════════════════════════════════════════════════

// GenerateProgressReport creates an AI-powered weekly summary for NGOs.
func (g *GeminiClient) GenerateProgressReport(ctx context.Context, reports []*domain.Report) (*domain.ProgressReport, error) {
	systemPrompt := `You are an NGO operations analyst. Generate a weekly progress report.
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
- recommendations: array of strings for next week`

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
	text, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("progress report: %w", err)
	}
	var result domain.ProgressReport
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("progress report decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 9: Volunteer Skill Recommendations
// ══════════════════════════════════════════════════

// RecommendSkills suggests new skills for volunteers based on task history.
func (g *GeminiClient) RecommendSkills(ctx context.Context, currentSkills []string, completedTaskTypes []string) (*domain.SkillRecommendation, error) {
	systemPrompt := `You are a volunteer development advisor.
Based on a volunteer's current skills and completed task history, suggest new skills.
Return JSON with:
- recommended_skills: array of objects {skill: string, reason: string, priority: "high"/"medium"/"low"}
- training_suggestions: array of strings
- career_path: string describing potential growth direction
- strengths: array of strings based on completed tasks`

	prompt := fmt.Sprintf("Current skills: %v\nCompleted task types: %v\n\nRecommend new skills for this volunteer.", currentSkills, completedTaskTypes)
	text, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("skill recommendation: %w", err)
	}
	var result domain.SkillRecommendation
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("skill recommendation decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 10: Document OCR
// ══════════════════════════════════════════════════

// OCRDocument extracts text from a document image.
func (g *GeminiClient) OCRDocument(ctx context.Context, base64Image string) (*domain.OCRResult, error) {
	systemPrompt := `You are an expert document reader (OCR). Extract all text from the document image.
Return JSON with:
- extracted_text: full text content of the document
- document_type: one of ["legal", "medical", "government", "general", "id_card", "receipt"]
- language: detected language
- key_fields: array of objects {field_name: string, field_value: string} for structured data
- confidence: float 0-1
- summary: 1-2 sentence summary of the document`

	text, err := g.generateImageText(ctx, systemPrompt, base64Image, "Extract all text and structured data from this document image.", true)
	if err != nil {
		return nil, fmt.Errorf("ocr: %w", err)
	}
	var result domain.OCRResult
	if err := decodeJSON(text, &result); err != nil {
		return nil, fmt.Errorf("ocr decode: %w", err)
	}
	return &result, nil
}

// ══════════════════════════════════════════════════
// FEATURE 11: AI Chatbot
// ══════════════════════════════════════════════════

// ChatWithAI provides context-aware conversation for volunteers.
func (g *GeminiClient) ChatWithAI(ctx context.Context, message string, taskContext string) (string, error) {
	systemPrompt := `You are SAMAJ AI Assistant — a helpful, friendly assistant for volunteers
working on civic issues, medical emergencies, legal aid, and disaster relief.
You provide practical, actionable guidance. Keep responses concise (2-4 paragraphs max).
If asked about safety, always err on the side of caution.
You can speak in Hindi or English based on the user's language.`

	prompt := message
	if taskContext != "" {
		prompt = fmt.Sprintf("Context about my current assigned task:\n%s\n\nMy question: %s", taskContext, message)
	}

	text, err := g.generateText(ctx, systemPrompt, prompt, false)
	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}
	return text, nil
}

// ══════════════════════════════════════════════════
// FEATURE 12: Case Q&A (Specialist AI-Powered Answers)
// ══════════════════════════════════════════════════

// AskCaseQuestion sends case documents and a question to Gemini for AI-powered Q&A.
func (g *GeminiClient) AskCaseQuestion(ctx context.Context, documentContext string, question string) (string, error) {
	systemPrompt := `You are a legal and medical case analyst for the SAMAJ platform.
You have access to case documents and reports. Answer the user's question based ONLY on the provided documents.
If the documents don't contain relevant information, say so clearly.
Cite specific document sections when possible. Be thorough but concise (3-5 paragraphs max).
You can respond in Hindi or English based on the question language.`

	prompt := fmt.Sprintf("CASE DOCUMENTS:\n%s\n\nQUESTION: %s\n\nProvide a detailed answer based on the documents above.", documentContext, question)
	text, err := g.generateText(ctx, systemPrompt, prompt, false)
	if err != nil {
		return "", fmt.Errorf("case Q&A: %w", err)
	}
	return text, nil
}

// ══════════════════════════════════════════════════
// FEATURE 13: Semantic Document Search
// ══════════════════════════════════════════════════

// SemanticSearchResult represents a single semantic search match.
type SemanticSearchResult struct {
	FileName string  `json:"file_name"`
	Excerpt  string  `json:"excerpt"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
}

// SemanticSearch uses Gemini to find relevant sections in documents.
func (g *GeminiClient) SemanticSearch(ctx context.Context, documentContext string, query string) ([]SemanticSearchResult, error) {
	systemPrompt := `You are a semantic document search engine.
Search through the provided documents and find sections relevant to the query.
Return JSON array with objects containing:
- file_name: name of the document
- excerpt: relevant text excerpt (50-200 chars)
- score: relevance score 0-1
- reason: brief explanation of why this is relevant
Return an empty array if no relevant sections found.
Return the results as a JSON object with key "results" containing the array.`

	prompt := fmt.Sprintf("DOCUMENTS:\n%s\n\nSEARCH QUERY: %s\n\nFind all relevant sections.", documentContext, query)
	text, err := g.generateText(ctx, systemPrompt, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}
	var wrapper struct {
		Results []SemanticSearchResult `json:"results"`
	}
	if err := decodeJSON(text, &wrapper); err != nil {
		// Try direct array parse
		var results []SemanticSearchResult
		if err2 := decodeJSON(text, &results); err2 != nil {
			return nil, fmt.Errorf("semantic search decode: %w (raw: %s)", err, text)
		}
		return results, nil
	}
	return wrapper.Results, nil
}

// ══════════════════════════════════════════════════
// FEATURE 14: Reverse Geocoding
// ══════════════════════════════════════════════════

// ReverseGeocode converts GPS coordinates into a human-readable location name using Gemini.
func (g *GeminiClient) ReverseGeocode(ctx context.Context, lat, lng float64) (string, error) {
	systemPrompt := `You are a geography expert. Given GPS coordinates (latitude, longitude),
return ONLY the location name — city, district, state, and country.
Format: "Area/Locality, City, State, Country"
If you cannot determine the exact location, provide your best estimate based on the coordinates.
Return ONLY the location string, nothing else.`

	prompt := fmt.Sprintf("What is the location at coordinates: Latitude %.6f, Longitude %.6f?", lat, lng)
	text, err := g.generateText(ctx, systemPrompt, prompt, false)
	if err != nil {
		return "", fmt.Errorf("reverse geocode: %w", err)
	}
	return strings.TrimSpace(text), nil
}

// ══════════════════════════════════════════════════
// Helper
// ══════════════════════════════════════════════════

func decodeJSON(raw string, target any) error {
	cleaned := cleanJSONText(raw)
	if err := json.Unmarshal([]byte(cleaned), target); err == nil {
		return nil
	}

	if object := extractJSONContainer(cleaned, "{", "}"); object != "" {
		if err := json.Unmarshal([]byte(object), target); err == nil {
			return nil
		}
	}
	if array := extractJSONContainer(cleaned, "[", "]"); array != "" {
		if err := json.Unmarshal([]byte(array), target); err == nil {
			return nil
		}
	}
	return json.Unmarshal([]byte(cleaned), target)
}

func cleanJSONText(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "\ufeff")
	if strings.HasPrefix(text, "```") {
		text = regexp.MustCompile(`(?s)^```(?:json)?\s*|\s*```$`).ReplaceAllString(text, "")
	}
	return strings.TrimSpace(text)
}

func extractJSONContainer(text, open, close string) string {
	start := strings.Index(text, open)
	end := strings.LastIndex(text, close)
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return strings.TrimSpace(text[start : end+1])
}

func decodeBase64Image(b64 string) ([]byte, string, error) {
	mimeType := "image/jpeg"
	if idx := strings.Index(b64, ","); idx != -1 {
		header := strings.ToLower(strings.TrimSpace(b64[:idx]))
		if strings.HasPrefix(header, "data:") && strings.Contains(header, ";base64") {
			mimeType = strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
		}
		b64 = b64[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return nil, "", fmt.Errorf("invalid base64 image: %w", err)
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("empty image data")
	}
	return data, mimeType, nil
}
