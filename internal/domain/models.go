package domain

import "time"

// Report represents a field report submitted by an NGO worker.
type Report struct {
	ID                       string    `json:"id" firestore:"id"`
	SubmitterUID             string    `json:"submitter_uid" firestore:"submitter_uid"`
	WardID                   string    `json:"ward_id" firestore:"ward_id"`
	RawText                  string    `json:"raw_text,omitempty" firestore:"raw_text"`
	MediaURL                 string    `json:"media_url,omitempty" firestore:"media_url"`
	MediaType                string    `json:"media_type,omitempty" firestore:"media_type"` // "image", "audio", "text"
	ProblemCategory          string    `json:"problem_category" firestore:"problem_category"`
	SeverityIndex            float64   `json:"severity_index" firestore:"severity_index"`
	AffectedPopulationEst    int       `json:"affected_population_estimate" firestore:"affected_population_estimate"`
	Summary                  string    `json:"summary" firestore:"summary"`
	UrgencyScore             float64   `json:"urgency_score" firestore:"urgency_score"`
	Status                   string    `json:"status" firestore:"status"` // "pending", "processing", "resolved"
	Latitude                 float64   `json:"latitude" firestore:"latitude"`
	Longitude                float64   `json:"longitude" firestore:"longitude"`
	CreatedAt                time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" firestore:"updated_at"`
	ResolvedAt               time.Time `json:"resolved_at,omitempty" firestore:"resolved_at,omitempty"`
	AssignedVolunteerIDs     []string  `json:"assigned_volunteer_ids,omitempty" firestore:"assigned_volunteer_ids,omitempty"`
}

// Ward represents a municipal ward with population data.
type Ward struct {
	ID                string  `json:"id" firestore:"id"`
	Name              string  `json:"name" firestore:"name"`
	PopulationDensity float64 `json:"population_density" firestore:"population_density"`
	CenterLat         float64 `json:"center_lat" firestore:"center_lat"`
	CenterLng         float64 `json:"center_lng" firestore:"center_lng"`
}

// Volunteer represents a registered volunteer.
type Volunteer struct {
	ID              string   `json:"id" firestore:"id"`
	UID             string   `json:"uid" firestore:"uid"`
	Name            string   `json:"name" firestore:"name"`
	Skills          []string `json:"skills" firestore:"skills"`
	Latitude        float64  `json:"latitude" firestore:"latitude"`
	Longitude       float64  `json:"longitude" firestore:"longitude"`
	Available       bool     `json:"available" firestore:"available"`
	CompletedTasks  int      `json:"completed_tasks" firestore:"completed_tasks"`
	AssignedTasks   int      `json:"assigned_tasks" firestore:"assigned_tasks"`
	CompletionRate  float64  `json:"completion_rate" firestore:"completion_rate"`
	S2CellID        uint64   `json:"s2_cell_id" firestore:"s2_cell_id"`
}

// User represents a platform user with a role.
type User struct {
	UID       string    `json:"uid" firestore:"uid"`
	Email     string    `json:"email" firestore:"email"`
	Name      string    `json:"name" firestore:"name"`
	Role      string    `json:"role" firestore:"role"` // "ngo_worker", "lawyer", "clerk", "nagar_nigam", "donor"
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

// UrgencyParams holds the inputs for computing urgency score.
type UrgencyParams struct {
	Frequency      float64
	Severity       float64
	PopImpact      float64
	LagBonus       float64
	TimeSinceHours float64
}

// MatchResult stores the computed matching score for a volunteer.
type MatchResult struct {
	VolunteerID    string  `json:"volunteer_id"`
	VolunteerName  string  `json:"volunteer_name"`
	SkillScore     float64 `json:"skill_score"`
	DistanceScore  float64 `json:"distance_score"`
	Reliability    float64 `json:"reliability"`
	TotalScore     float64 `json:"total_score"`
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
