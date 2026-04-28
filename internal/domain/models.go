package domain

import "time"

// Report represents an issue submitted by a General User (Reporter).
type Report struct {
	ID                    string    `json:"id" firestore:"id"`
	SubmitterUID          string    `json:"submitter_uid" firestore:"submitter_uid"`
	RawText               string    `json:"raw_text,omitempty" firestore:"raw_text"`
	MediaURL              string    `json:"media_url,omitempty" firestore:"media_url"`
	MediaType             string    `json:"media_type,omitempty" firestore:"media_type"` // "image", "text"
	IssueType             string    `json:"issue_type" firestore:"issue_type"`           // "medical_emergency", "legal_aid", "civic_issue", "disaster_relief"
	UserUrgency           string    `json:"user_urgency" firestore:"user_urgency"`       // "normal", "urgent", "critical"
	RequiredVolunteers    int       `json:"required_volunteers" firestore:"required_volunteers"`
	Location              string    `json:"location" firestore:"location"` // text address
	Latitude              float64   `json:"latitude" firestore:"latitude"`
	Longitude             float64   `json:"longitude" firestore:"longitude"`
	ProblemCategory       string    `json:"problem_category" firestore:"problem_category"`
	SeverityIndex         float64   `json:"severity_index" firestore:"severity_index"`
	AffectedPopulationEst int       `json:"affected_population_estimate" firestore:"affected_population_estimate"`
	Summary               string    `json:"summary" firestore:"summary"`
	UrgencyScore          float64   `json:"urgency_score" firestore:"urgency_score"`
	Status                string    `json:"status" firestore:"status"` // "pending", "accepted", "in_progress", "escalated", "resolved"
	AssignedVolunteerIDs  []string  `json:"assigned_volunteer_ids,omitempty" firestore:"assigned_volunteer_ids,omitempty"`
	AssignedSpecialistUID string    `json:"assigned_specialist_uid,omitempty" firestore:"assigned_specialist_uid,omitempty"`
	CreatedAt             time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" firestore:"updated_at"`
	ResolvedAt            time.Time `json:"resolved_at,omitempty" firestore:"resolved_at,omitempty"`
}

// Volunteer represents a registered volunteer (general or specialist).
type Volunteer struct {
	ID             string   `json:"id" firestore:"id"`
	UID            string   `json:"uid" firestore:"uid"`
	Name           string   `json:"name" firestore:"name"`
	Role           string   `json:"role" firestore:"role"` // "volunteer" or "specialist"
	Specialization string   `json:"specialization,omitempty" firestore:"specialization,omitempty"` // "lawyer", "doctor"
	Skills         []string `json:"skills" firestore:"skills"`
	Latitude       float64  `json:"latitude" firestore:"latitude"`
	Longitude      float64  `json:"longitude" firestore:"longitude"`
	Available      bool     `json:"available" firestore:"available"`
	CompletedTasks int      `json:"completed_tasks" firestore:"completed_tasks"`
	AssignedTasks  int      `json:"assigned_tasks" firestore:"assigned_tasks"`
	CompletionRate float64  `json:"completion_rate" firestore:"completion_rate"`
	S2CellID       int64    `json:"s2_cell_id" firestore:"s2_cell_id"`
}

// User represents a platform user with a role.
type User struct {
	UID       string    `json:"uid" firestore:"uid"`
	Email     string    `json:"email" firestore:"email"`
	Name      string    `json:"name" firestore:"name"`
	Role      string    `json:"role" firestore:"role"` // "reporter", "volunteer", "specialist", "ngo_admin"
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

// CaseFile represents a bundle of documents assigned to a specialist for a set of reports.
type CaseFile struct {
	ID                   string         `json:"id" firestore:"id"`
	AssignedSpecialistUID string        `json:"assigned_specialist_uid" firestore:"assigned_specialist_uid"`
	ReportIDs            []string       `json:"report_ids" firestore:"report_ids"`
	Title                string         `json:"title" firestore:"title"`
	Status               string         `json:"status" firestore:"status"` // "open", "in_review", "closed"
	Documents            []CaseDocument `json:"documents" firestore:"documents"`
	CreatedAt            time.Time      `json:"created_at" firestore:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" firestore:"updated_at"`
}

// CaseDocument represents a document uploaded to a case file.
type CaseDocument struct {
	ID         string    `json:"id" firestore:"id"`
	FileName   string    `json:"file_name" firestore:"file_name"`
	Content    string    `json:"content" firestore:"content"` // base64 or extracted text
	FileType   string    `json:"file_type" firestore:"file_type"` // "pdf", "image", "text"
	UploadedAt time.Time `json:"uploaded_at" firestore:"uploaded_at"`
}

// Ward represents a municipal ward with population data.
type Ward struct {
	ID                string  `json:"id" firestore:"id"`
	Name              string  `json:"name" firestore:"name"`
	PopulationDensity float64 `json:"population_density" firestore:"population_density"`
	CenterLat         float64 `json:"center_lat" firestore:"center_lat"`
	CenterLng         float64 `json:"center_lng" firestore:"center_lng"`
}

// MatchResult stores the computed matching score for a volunteer.
type MatchResult struct {
	VolunteerID   string  `json:"volunteer_id"`
	VolunteerName string  `json:"volunteer_name"`
	SkillScore    float64 `json:"skill_score"`
	DistanceScore float64 `json:"distance_score"`
	Reliability   float64 `json:"reliability"`
	TotalScore    float64 `json:"total_score"`
}

// UrgencyParams holds the inputs for computing urgency score.
type UrgencyParams struct {
	Frequency      float64
	Severity       float64
	PopImpact      float64
	LagBonus       float64
	TimeSinceHours float64
}

// GeminiExtraction is the structured output from the Gemini parser.
type GeminiExtraction struct {
	ProblemCategory       string  `json:"problem_category"`
	WardID                string  `json:"ward_id"`
	SeverityIndex         float64 `json:"severity_index"`
	AffectedPopulationEst int     `json:"affected_population_estimate"`
	Summary               string  `json:"summary"`
}

// PubSubMessage represents a Cloud Pub/Sub push message body.
type PubSubMessage struct {
	Message struct {
		Data       string            `json:"data"`
		Attributes map[string]string `json:"attributes,omitempty"`
		ID         string            `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// IngestionEvent is the payload published to the report-ingestion topic.
type IngestionEvent struct {
	ReportID  string `json:"report_id"`
	MediaURL  string `json:"media_url"`
	MediaType string `json:"media_type"`
	RawText   string `json:"raw_text"`
}

// ══════════════════════════════════════════════════
// AI Feature Domain Models
// ══════════════════════════════════════════════════

// ImageAnalysis is the structured output from Gemini image analysis.
type ImageAnalysis struct {
	IssueType                 string   `json:"issue_type"`
	Description               string   `json:"description"`
	Severity                  int      `json:"severity"`
	DetectedObjects           []string `json:"detected_objects"`
	SuggestedCategory         string   `json:"suggested_category"`
	LocationHints             string   `json:"location_hints"`
	RequiresImmediateAttention bool    `json:"requires_immediate_attention"`
}

// ReportVerification is the result of cross-checking image vs text.
type ReportVerification struct {
	IsConsistent       bool     `json:"is_consistent"`
	Confidence         float64  `json:"confidence"`
	Discrepancies      []string `json:"discrepancies"`
	VerificationSummary string  `json:"verification_summary"`
	RiskScore          int      `json:"risk_score"`
}

// SimilarReport is a single match in duplicate detection.
type SimilarReport struct {
	Index           int     `json:"index"`
	SimilarityScore float64 `json:"similarity_score"`
	Reason          string  `json:"reason"`
}

// DuplicateDetection is the result of comparing reports for duplicates.
type DuplicateDetection struct {
	HasDuplicates  bool            `json:"has_duplicates"`
	SimilarReports []SimilarReport `json:"similar_reports"`
	Recommendation string          `json:"recommendation"`
	Summary        string          `json:"summary"`
}

// ActionStep is a single step in an action plan.
type ActionStep struct {
	StepNumber  int    `json:"step_number"`
	Action      string `json:"action"`
	Details     string `json:"details"`
	SafetyNotes string `json:"safety_notes"`
}

// ActionPlan is a volunteer's step-by-step task plan.
type ActionPlan struct {
	Title             string       `json:"title"`
	EstimatedDuration string       `json:"estimated_duration"`
	PriorityLevel     string       `json:"priority_level"`
	Steps             []ActionStep `json:"steps"`
	RequiredEquipment []string     `json:"required_equipment"`
	SafetyWarnings    []string     `json:"safety_warnings"`
	Tips              []string     `json:"tips"`
}

// SentimentAnalysis is the emotional tone analysis of a report.
type SentimentAnalysis struct {
	OverallSentiment   string   `json:"overall_sentiment"`
	UrgencyBoost       float64  `json:"urgency_boost"`
	EmotionsDetected   []string `json:"emotions_detected"`
	EmotionalIntensity float64  `json:"emotional_intensity"`
	KeyPhrases         []string `json:"key_phrases"`
	Recommendation     string   `json:"recommendation"`
}

// Translation is the result of a language translation.
type Translation struct {
	TranslatedText string  `json:"translated_text"`
	SourceLanguage string  `json:"source_language"`
	TargetLanguage string  `json:"target_language"`
	Confidence     float64 `json:"confidence"`
	CulturalNotes  string  `json:"cultural_notes,omitempty"`
}

// CategoryCount tracks issue category statistics.
type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Trend    string `json:"trend"`
}

// ProgressReport is an AI-generated weekly summary.
type ProgressReport struct {
	Title            string          `json:"title"`
	DateRange        string          `json:"date_range"`
	ExecutiveSummary string          `json:"executive_summary"`
	TotalIssues      int             `json:"total_issues"`
	ResolvedIssues   int             `json:"resolved_issues"`
	PendingIssues    int             `json:"pending_issues"`
	CriticalIssues   int             `json:"critical_issues"`
	ResolutionRate   float64         `json:"resolution_rate"`
	TopCategories    []CategoryCount `json:"top_categories"`
	HotspotAreas     []string        `json:"hotspot_areas"`
	KeyAchievements  []string        `json:"key_achievements"`
	AreasOfConcern   []string        `json:"areas_of_concern"`
	Recommendations  []string        `json:"recommendations"`
}

// RecommendedSkill is a single skill recommendation.
type RecommendedSkill struct {
	Skill    string `json:"skill"`
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
}

// SkillRecommendation is AI-suggested skill development for a volunteer.
type SkillRecommendation struct {
	RecommendedSkills   []RecommendedSkill `json:"recommended_skills"`
	TrainingSuggestions  []string           `json:"training_suggestions"`
	CareerPath          string             `json:"career_path"`
	Strengths           []string           `json:"strengths"`
}

// KeyField is a structured field extracted from OCR.
type KeyField struct {
	FieldName  string `json:"field_name"`
	FieldValue string `json:"field_value"`
}

// OCRResult is the text extracted from a document image.
type OCRResult struct {
	ExtractedText string     `json:"extracted_text"`
	DocumentType  string     `json:"document_type"`
	Language      string     `json:"language"`
	KeyFields     []KeyField `json:"key_fields"`
	Confidence    float64    `json:"confidence"`
	Summary       string     `json:"summary"`
}
