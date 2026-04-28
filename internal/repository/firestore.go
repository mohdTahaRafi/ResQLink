package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/resqlink-project/resqlink/internal/domain"
	"google.golang.org/api/iterator"
)

// FirestoreRepo implements persistence against Cloud Firestore.
type FirestoreRepo struct {
	client *firestore.Client
}

// NewFirestoreRepo creates a new Firestore repository.
func NewFirestoreRepo(client *firestore.Client) *FirestoreRepo {
	return &FirestoreRepo{client: client}
}

// CreateReport persists a new report.
func (r *FirestoreRepo) CreateReport(ctx context.Context, report *domain.Report) (string, error) {
	report.CreatedAt = time.Now()
	report.UpdatedAt = report.CreatedAt
	report.Status = "pending"

	ref, _, err := r.client.Collection("reports").Add(ctx, report)
	if err != nil {
		return "", fmt.Errorf("firestore create report: %w", err)
	}
	report.ID = ref.ID
	return ref.ID, nil
}

// GetReport retrieves a single report by ID.
func (r *FirestoreRepo) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	doc, err := r.client.Collection("reports").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore get report: %w", err)
	}
	var report domain.Report
	if err := doc.DataTo(&report); err != nil {
		return nil, fmt.Errorf("firestore decode report: %w", err)
	}
	report.ID = doc.Ref.ID
	return &report, nil
}

// UpdateReport updates specific fields of a report.
func (r *FirestoreRepo) UpdateReport(ctx context.Context, id string, updates []firestore.Update) error {
	updates = append(updates, firestore.Update{Path: "updated_at", Value: time.Now()})
	_, err := r.client.Collection("reports").Doc(id).Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("firestore update report: %w", err)
	}
	return nil
}

// GetReportsByWard fetches reports for a ward within a time window.
func (r *FirestoreRepo) GetReportsByWard(ctx context.Context, wardID string, since time.Time) ([]*domain.Report, error) {
	iter := r.client.Collection("reports").
		Where("ward_id", "==", wardID).
		Where("created_at", ">=", since).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var reports []*domain.Report
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate reports: %w", err)
		}
		var rpt domain.Report
		if err := doc.DataTo(&rpt); err != nil {
			return nil, fmt.Errorf("firestore decode report in list: %w", err)
		}
		rpt.ID = doc.Ref.ID
		reports = append(reports, &rpt)
	}
	return reports, nil
}

// GetUnresolvedReportsByWard returns unresolved reports for a ward.
func (r *FirestoreRepo) GetUnresolvedReportsByWard(ctx context.Context, wardID string) ([]*domain.Report, error) {
	iter := r.client.Collection("reports").
		Where("ward_id", "==", wardID).
		Where("status", "!=", "resolved").
		Documents(ctx)
	defer iter.Stop()

	var reports []*domain.Report
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate unresolved: %w", err)
		}
		var rpt domain.Report
		if err := doc.DataTo(&rpt); err != nil {
			return nil, fmt.Errorf("firestore decode unresolved: %w", err)
		}
		rpt.ID = doc.Ref.ID
		reports = append(reports, &rpt)
	}
	return reports, nil
}

// GetWard retrieves ward details.
func (r *FirestoreRepo) GetWard(ctx context.Context, wardID string) (*domain.Ward, error) {
	doc, err := r.client.Collection("wards").Doc(wardID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore get ward: %w", err)
	}
	var ward domain.Ward
	if err := doc.DataTo(&ward); err != nil {
		return nil, fmt.Errorf("firestore decode ward: %w", err)
	}
	ward.ID = doc.Ref.ID
	return &ward, nil
}

// GetVolunteersByWard retrieves available volunteers near a ward using S2 cell prefix.
func (r *FirestoreRepo) GetVolunteersByWard(ctx context.Context, wardID string) ([]*domain.Volunteer, error) {
	iter := r.client.Collection("volunteers").
		Where("available", "==", true).
		Documents(ctx)
	defer iter.Stop()

	var volunteers []*domain.Volunteer
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate volunteers: %w", err)
		}
		var v domain.Volunteer
		if err := doc.DataTo(&v); err != nil {
			return nil, fmt.Errorf("firestore decode volunteer: %w", err)
		}
		v.ID = doc.Ref.ID
		volunteers = append(volunteers, &v)
	}
	return volunteers, nil
}

// CreateVolunteer persists a new volunteer.
func (r *FirestoreRepo) CreateVolunteer(ctx context.Context, vol *domain.Volunteer) (string, error) {
	ref, _, err := r.client.Collection("volunteers").Add(ctx, vol)
	if err != nil {
		return "", fmt.Errorf("firestore create volunteer: %w", err)
	}
	vol.ID = ref.ID
	return ref.ID, nil
}

// GetVolunteerByUID finds a volunteer by their Firebase UID (for duplicate prevention).
func (r *FirestoreRepo) GetVolunteerByUID(ctx context.Context, uid string) (*domain.Volunteer, error) {
	iter := r.client.Collection("volunteers").
		Where("uid", "==", uid).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("firestore get volunteer by uid: %w", err)
	}

	var v domain.Volunteer
	if err := doc.DataTo(&v); err != nil {
		return nil, fmt.Errorf("firestore decode volunteer: %w", err)
	}
	v.ID = doc.Ref.ID
	return &v, nil
}

// GetAllReports retrieves all reports, optionally limited.
func (r *FirestoreRepo) GetAllReports(ctx context.Context, limit int) ([]*domain.Report, error) {
	q := r.client.Collection("reports").OrderBy("created_at", firestore.Desc)
	if limit > 0 {
		q = q.Limit(limit)
	}
	iter := q.Documents(ctx)
	defer iter.Stop()

	var reports []*domain.Report
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate all reports: %w", err)
		}
		var rpt domain.Report
		if err := doc.DataTo(&rpt); err != nil {
			return nil, fmt.Errorf("firestore decode report: %w", err)
		}
		rpt.ID = doc.Ref.ID
		reports = append(reports, &rpt)
	}
	return reports, nil
}

// CreateCaseFile persists a new case file.
func (r *FirestoreRepo) CreateCaseFile(ctx context.Context, cf *domain.CaseFile) (string, error) {
	cf.CreatedAt = time.Now()
	cf.UpdatedAt = cf.CreatedAt
	cf.Status = "open"

	ref, _, err := r.client.Collection("case_files").Add(ctx, cf)
	if err != nil {
		return "", fmt.Errorf("firestore create case file: %w", err)
	}
	cf.ID = ref.ID
	return ref.ID, nil
}

// GetCaseFilesBySpecialist retrieves case files assigned to a specialist.
func (r *FirestoreRepo) GetCaseFilesBySpecialist(ctx context.Context, uid string) ([]*domain.CaseFile, error) {
	iter := r.client.Collection("case_files").
		Where("assigned_specialist_uid", "==", uid).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var cases []*domain.CaseFile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate case files: %w", err)
		}
		var cf domain.CaseFile
		if err := doc.DataTo(&cf); err != nil {
			return nil, fmt.Errorf("firestore decode case file: %w", err)
		}
		cf.ID = doc.Ref.ID
		cases = append(cases, &cf)
	}
	return cases, nil
}

// GetCaseFile retrieves a single case file by ID.
func (r *FirestoreRepo) GetCaseFile(ctx context.Context, id string) (*domain.CaseFile, error) {
	doc, err := r.client.Collection("case_files").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore get case file: %w", err)
	}
	var cf domain.CaseFile
	if err := doc.DataTo(&cf); err != nil {
		return nil, fmt.Errorf("firestore decode case file: %w", err)
	}
	cf.ID = doc.Ref.ID
	return &cf, nil
}

// AddDocumentToCaseFile adds a document to an existing case file.
func (r *FirestoreRepo) AddDocumentToCaseFile(ctx context.Context, caseID string, doc domain.CaseDocument) error {
	_, err := r.client.Collection("case_files").Doc(caseID).Update(ctx, []firestore.Update{
		{Path: "documents", Value: firestore.ArrayUnion(doc)},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

// GetReportsByAssignedVolunteer returns reports assigned to a specific volunteer UID.
func (r *FirestoreRepo) GetReportsByAssignedVolunteer(ctx context.Context, uid string) ([]*domain.Report, error) {
	iter := r.client.Collection("reports").
		Where("assigned_volunteer_ids", "array-contains", uid).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var reports []*domain.Report
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate volunteer tasks: %w", err)
		}
		var rpt domain.Report
		if err := doc.DataTo(&rpt); err != nil {
			return nil, fmt.Errorf("firestore decode volunteer task: %w", err)
		}
		rpt.ID = doc.Ref.ID
		reports = append(reports, &rpt)
	}
	return reports, nil
}

// GetAllVolunteers retrieves all available volunteers.
func (r *FirestoreRepo) GetAllVolunteers(ctx context.Context) ([]*domain.Volunteer, error) {
	iter := r.client.Collection("volunteers").
		Where("available", "==", true).
		Documents(ctx)
	defer iter.Stop()

	var volunteers []*domain.Volunteer
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestore iterate all volunteers: %w", err)
		}
		var v domain.Volunteer
		if err := doc.DataTo(&v); err != nil {
			return nil, fmt.Errorf("firestore decode volunteer: %w", err)
		}
		v.ID = doc.Ref.ID
		volunteers = append(volunteers, &v)
	}
	return volunteers, nil
}

