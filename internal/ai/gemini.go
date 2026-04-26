package ai

import (
	"context"
	"encoding/json"
	"fmt"

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
	model     *genai.GenerativeModel
	aiClient  *genai.Client
}

// NewGeminiClient initializes a Gemini 1.5 Pro client via Vertex AI.
func NewGeminiClient(ctx context.Context, projectID, location string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("vertex ai client init: %w", err)
	}

	model := client.GenerativeModel("gemini-3.1-pro")
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

// ParseTextReport sends a text-only report for structured extraction.
func (g *GeminiClient) ParseTextReport(ctx context.Context, rawText string) (*domain.GeminiExtraction, error) {
	prompt := fmt.Sprintf("Parse the following field report and extract structured data:\n\n%s", rawText)

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini text generation: %w", err)
	}

	return g.extractResult(resp)
}

// ParseImageReport sends an image (from GCS URI) for multimodal extraction.
func (g *GeminiClient) ParseImageReport(ctx context.Context, gcsURI string, caption string) (*domain.GeminiExtraction, error) {
	img := genai.FileData{
		MIMEType: "image/jpeg",
		FileURI:  gcsURI,
	}

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

// ParseAudioReport sends an audio file (from GCS URI) for transcription and extraction.
func (g *GeminiClient) ParseAudioReport(ctx context.Context, gcsURI string) (*domain.GeminiExtraction, error) {
	audio := genai.FileData{
		MIMEType: "audio/mpeg",
		FileURI:  gcsURI,
	}

	prompt := "Transcribe and parse this voice report from a field worker. The audio may be in Hindi or a local dialect."

	resp, err := g.model.GenerateContent(ctx, audio, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini audio generation: %w", err)
	}

	return g.extractResult(resp)
}

func (g *GeminiClient) extractResult(resp *genai.GenerateContentResponse) (*domain.GeminiExtraction, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("gemini response part is not text")
	}

	var extraction domain.GeminiExtraction
	if err := json.Unmarshal([]byte(text), &extraction); err != nil {
		return nil, fmt.Errorf("gemini json decode: %w (raw: %s)", err, string(text))
	}

	if extraction.SeverityIndex < 1 {
		extraction.SeverityIndex = 1
	}
	if extraction.SeverityIndex > 10 {
		extraction.SeverityIndex = 10
	}

	return &extraction, nil
}
