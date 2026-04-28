package repository

import (
	"context"
	"testing"
	"time"

	"github.com/resqlink-project/resqlink/internal/domain"
)

// TestSaveAndGetReport tests basic save and retrieve operations
func TestSaveAndGetReport(t *testing.T) {
	// Note: This test requires a Firestore emulator or real Firestore instance
	// For CI/CD, use: gcloud beta emulators firestore start

	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil) // Would need actual client in real test

	report := &domain.Report{
		ID:          "test-report-1",
		SubmitterUID: "user-123",
		RawText:     "Test report",
		IssueType:   "medical",
		Status:      domain.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save report
	err := repo.SaveReport(ctx, report)
	if err != nil {
		t.Fatalf("SaveReport failed: %v", err)
	}

	// Get report
	retrieved, err := repo.GetReport(ctx, "test-report-1")
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved report is nil")
	}

	if retrieved.ID != report.ID {
		t.Errorf("Expected ID %s, got %s", report.ID, retrieved.ID)
	}
}

// TestUpdateReport tests update operations
func TestUpdateReport(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	// Update non-existent report should not error
	err := repo.UpdateReport(ctx, "non-existent", map[string]interface{}{
		"status": domain.StatusAccepted,
	})

	if err != nil {
		t.Logf("UpdateReport on non-existent returned error (expected): %v", err)
	}
}

// TestGetPrioritizedReports tests urgency-based sorting
func TestGetPrioritizedReports(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	reports, err := repo.GetPrioritizedReports(ctx)
	if err != nil {
		t.Fatalf("GetPrioritizedReports failed: %v", err)
	}

	// Verify reports are sorted by urgency score descending
	for i := 1; i < len(reports); i++ {
		if reports[i-1].UrgencyScore < reports[i].UrgencyScore {
			t.Errorf("Reports not properly sorted by urgency score")
		}
	}
}

// TestGetReportsByVolunteer tests volunteer task retrieval
func TestGetReportsByVolunteer(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	volunteerUID := "volunteer-123"
	reports, err := repo.GetReportsByVolunteer(ctx, volunteerUID)
	if err != nil {
		t.Fatalf("GetReportsByVolunteer failed: %v", err)
	}

	// Verify all returned reports have this volunteer assigned
	for _, report := range reports {
		found := false
		for _, uid := range report.AssignedVolunteers {
			if uid == volunteerUID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Report %s does not have volunteer %s assigned", report.ID, volunteerUID)
		}
	}
}

// TestCountReportsByStatus tests status-based counting
func TestCountReportsByStatus(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	count, err := repo.CountReportsByStatus(ctx, domain.StatusResolved)
	if err != nil {
		t.Fatalf("CountReportsByStatus failed: %v", err)
	}

	if count < 0 {
		t.Errorf("Count cannot be negative, got %d", count)
	}
}

// TestSaveAndGetVolunteer tests volunteer operations
func TestSaveAndGetVolunteer(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	volunteer := &domain.Volunteer{
		UID:            "volunteer-123",
		Name:           "John Doe",
		Skills:         []string{"first_aid", "cpr"},
		Latitude:       12.97,
		Longitude:      77.59,
		CompletionRate: 0.95,
		IsAvailable:    true,
		CreatedAt:      time.Now(),
	}

	err := repo.SaveVolunteer(ctx, volunteer)
	if err != nil {
		t.Fatalf("SaveVolunteer failed: %v", err)
	}

	retrieved, err := repo.GetVolunteer(ctx, "volunteer-123")
	if err != nil {
		t.Fatalf("GetVolunteer failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved volunteer is nil")
	}

	if retrieved.UID != volunteer.UID {
		t.Errorf("Expected UID %s, got %s", volunteer.UID, retrieved.UID)
	}
}

// TestGetAllAvailableVolunteers tests filtering available volunteers
func TestGetAllAvailableVolunteers(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	volunteers, err := repo.GetAllAvailableVolunteers(ctx)
	if err != nil {
		t.Fatalf("GetAllAvailableVolunteers failed: %v", err)
	}

	// Verify all returned volunteers are available
	for _, vol := range volunteers {
		if !vol.IsAvailable {
			t.Errorf("Volunteer %s is not available but was returned", vol.UID)
		}
	}
}

// TestGetWard tests ward retrieval
func TestGetWard(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	ward, err := repo.GetWard(ctx, "ward-123")
	if err != nil {
		t.Logf("GetWard returned error (may be expected if ward doesn't exist): %v", err)
	}

	if ward == nil && err == nil {
		// This is acceptable - ward might not exist
		return
	}

	if ward != nil && ward.WardID != "ward-123" {
		t.Errorf("Expected ward ID ward-123, got %s", ward.WardID)
	}
}

// TestGetRecentReportCountForWard tests time-based report counting
func TestGetRecentReportCountForWard(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	count, err := repo.GetRecentReportCountForWard(ctx, "ward-123", 24)
	if err != nil {
		t.Fatalf("GetRecentReportCountForWard failed: %v", err)
	}

	if count < 0 {
		t.Errorf("Count cannot be negative, got %d", count)
	}
}

// TestSaveAndGetCaseFile tests case operations
func TestSaveAndGetCaseFile(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	caseFile := &domain.CaseFile{
		ID:                    "case-123",
		Title:                 "Medical Emergency Response",
		Status:                "open",
		AssignedSpecialistUID: "specialist-123",
		LinkedReportIDs:       []string{"report-1", "report-2"},
		Documents:             []domain.CaseDocument{},
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.SaveCaseFile(ctx, caseFile)
	if err != nil {
		t.Fatalf("SaveCaseFile failed: %v", err)
	}

	retrieved, err := repo.GetCaseFile(ctx, "case-123")
	if err != nil {
		t.Fatalf("GetCaseFile failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved case file is nil")
	}

	if retrieved.ID != caseFile.ID {
		t.Errorf("Expected case ID %s, got %s", caseFile.ID, retrieved.ID)
	}
}

// TestGetCasesBySpecialist tests specialist case filtering
func TestGetCasesBySpecialist(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	specialistUID := "specialist-123"
	cases, err := repo.GetCasesBySpecialist(ctx, specialistUID)
	if err != nil {
		t.Fatalf("GetCasesBySpecialist failed: %v", err)
	}

	// Verify all returned cases are assigned to the specialist
	for _, caseFile := range cases {
		if caseFile.AssignedSpecialistUID != specialistUID {
			t.Errorf("Case %s not assigned to specialist %s", caseFile.ID, specialistUID)
		}
	}
}

// TestAddDocumentToCase tests document addition
func TestAddDocumentToCase(t *testing.T) {
	t.Skip("Requires Firestore emulator setup")

	ctx := context.Background()
	repo := NewFirestoreRepo(nil)

	doc := domain.CaseDocument{
		ID:         "doc-123",
		Filename:   "evidence.pdf",
		MediaURL:   "https://storage.example.com/evidence.pdf",
		MimeType:   "application/pdf",
		UploadedAt: time.Now(),
	}

	err := repo.AddDocumentToCase(ctx, "case-123", doc)
	if err != nil {
		t.Logf("AddDocumentToCase returned error (may be expected if case doesn't exist): %v", err)
	}
}
