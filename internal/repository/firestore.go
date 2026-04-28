package repository

import (
	"context"
	"fmt"
	"time"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/resqlink-project/resqlink/internal/domain"
	"google.golang.org/api/iterator"
)

// FirestoreRepo implements persistence against Cloud Firestore.
type FirestoreRepo struct {
	client *firestore.Client
	mu sync.Mutex
	reports map[string]*domain.Report
	volunteers map[string]*domain.Volunteer
	cases map[string]*domain.CaseFile
	wards map[string]*domain.Ward
}

// NewFirestoreRepo creates a new Firestore repository.
func NewFirestoreRepo(client *firestore.Client) *FirestoreRepo {
	return &FirestoreRepo{
		client: client,
		reports: make(map[string]*domain.Report),
		volunteers: make(map[string]*domain.Volunteer),
		cases: make(map[string]*domain.CaseFile),
		wards: make(map[string]*domain.Ward),
	}
}

func (r *FirestoreRepo) isMemory() bool { return r.client == nil }

// SaveReport persists a new report.
func (r *FirestoreRepo) SaveReport(ctx context.Context, report *domain.Report) error {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if report.ID == "" {
			report.ID = fmt.Sprintf("report-%d", time.Now().UnixNano())
		}
		report.CreatedAt = time.Now()
		report.UpdatedAt = report.CreatedAt
		if report.Status == "" {
			report.Status = domain.StatusPending
		}
		cp := *report
		r.reports[report.ID] = &cp
		return nil
	}
	report.CreatedAt = time.Now()
	report.UpdatedAt = report.CreatedAt
	if report.Status == "" {
		report.Status = domain.StatusPending
	}
	if len(report.AssignedVolunteerIDs) == 0 && len(report.AssignedVolunteers) > 0 {
		report.AssignedVolunteerIDs = append([]string{}, report.AssignedVolunteers...)
	}

	ref, _, err := r.client.Collection("reports").Add(ctx, report)
	if err != nil {
		return fmt.Errorf("firestore create report: %w", err)
	}
	report.ID = ref.ID
	return nil
}

// CreateReport persists a new report and returns the generated ID.
func (r *FirestoreRepo) CreateReport(ctx context.Context, report *domain.Report) (string, error) {
	if err := r.SaveReport(ctx, report); err != nil {
		return "", err
	}
	return report.ID, nil
}

// GetReport retrieves a single report by ID.
func (r *FirestoreRepo) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if report, ok := r.reports[id]; ok {
			cp := *report
			return &cp, nil
		}
		return nil, fmt.Errorf("firestore get report: not found")
	}
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
func (r *FirestoreRepo) UpdateReport(ctx context.Context, id string, updates interface{}) error {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		report, ok := r.reports[id]
		if !ok {
			return fmt.Errorf("firestore update report: not found")
		}
		switch v := updates.(type) {
		case []firestore.Update:
			for _, u := range v {
				applyReportUpdate(report, u.Path, u.Value)
			}
		case map[string]interface{}:
			for k, val := range v {
				applyReportUpdate(report, k, val)
			}
		}
		report.UpdatedAt = time.Now()
		return nil
	}
	var firestoreUpdates []firestore.Update
	switch v := updates.(type) {
	case []firestore.Update:
		firestoreUpdates = append(firestoreUpdates, v...)
	case map[string]interface{}:
		for k, val := range v {
			firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: k, Value: val})
		}
	default:
		return fmt.Errorf("unsupported update payload type %T", updates)
	}
	firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: "updated_at", Value: time.Now()})
	_, err := r.client.Collection("reports").Doc(id).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("firestore update report: %w", err)
	}
	return nil
}

// GetReportsByWard fetches reports for a ward within a time window.
func (r *FirestoreRepo) GetReportsByWard(ctx context.Context, wardID string, since time.Time) ([]*domain.Report, error) {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var reports []*domain.Report
		for _, rpt := range r.reports {
			if rpt.WardID == wardID && rpt.CreatedAt.After(since) {
				cp := *rpt
				reports = append(reports, &cp)
			}
		}
		return reports, nil
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var reports []*domain.Report
		for _, rpt := range r.reports {
			if rpt.WardID == wardID && rpt.Status != domain.StatusResolved {
				cp := *rpt
				reports = append(reports, &cp)
			}
		}
		return reports, nil
	}
	iter := r.client.Collection("reports").
		Where("ward_id", "==", wardID).
		Where("status", "!=", domain.StatusResolved).
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if ward, ok := r.wards[wardID]; ok {
			cp := *ward
			return &cp, nil
		}
		return &domain.Ward{ID: wardID, WardID: wardID, PopulationDensity: 1000}, nil
	}
	doc, err := r.client.Collection("wards").Doc(wardID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore get ward: %w", err)
	}
	var ward domain.Ward
	if err := doc.DataTo(&ward); err != nil {
		return nil, fmt.Errorf("firestore decode ward: %w", err)
	}
	ward.ID = doc.Ref.ID
	ward.WardID = doc.Ref.ID
	return &ward, nil
}

// GetVolunteersByWard retrieves available volunteers near a ward using S2 cell prefix.
func (r *FirestoreRepo) GetVolunteersByWard(ctx context.Context, wardID string) ([]*domain.Volunteer, error) {
	if r.isMemory() {
		return r.GetAllVolunteers(ctx)
	}
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

// SaveVolunteer persists a new volunteer.
func (r *FirestoreRepo) SaveVolunteer(ctx context.Context, vol *domain.Volunteer) error {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if vol.ID == "" {
			vol.ID = fmt.Sprintf("vol-%d", time.Now().UnixNano())
		}
		if vol.IsAvailable {
			vol.Available = true
		}
		vol.CreatedAt = time.Now()
		cp := *vol
		r.volunteers[vol.UID] = &cp
		return nil
	}
	if vol.IsAvailable {
		vol.Available = true
	}
	ref, _, err := r.client.Collection("volunteers").Add(ctx, vol)
	if err != nil {
		return fmt.Errorf("firestore create volunteer: %w", err)
	}
	vol.ID = ref.ID
	return nil
}

// CreateVolunteer persists a new volunteer and returns the generated ID.
func (r *FirestoreRepo) CreateVolunteer(ctx context.Context, vol *domain.Volunteer) (string, error) {
	if err := r.SaveVolunteer(ctx, vol); err != nil {
		return "", err
	}
	return vol.ID, nil
}

// GetVolunteerByUID finds a volunteer by their Firebase UID (for duplicate prevention).
func (r *FirestoreRepo) GetVolunteerByUID(ctx context.Context, uid string) (*domain.Volunteer, error) {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if v, ok := r.volunteers[uid]; ok {
			cp := *v
			return &cp, nil
		}
		return nil, nil
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var reports []*domain.Report
		for _, rpt := range r.reports {
			cp := *rpt
			reports = append(reports, &cp)
		}
		return reports, nil
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if cf.ID == "" {
			cf.ID = fmt.Sprintf("case-%d", time.Now().UnixNano())
		}
		cf.CreatedAt = time.Now()
		cf.UpdatedAt = cf.CreatedAt
		r.cases[cf.ID] = cf
		return cf.ID, nil
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var cases []*domain.CaseFile
		for _, cf := range r.cases {
			if cf.AssignedSpecialistUID == uid {
				cp := *cf
				cases = append(cases, &cp)
			}
		}
		return cases, nil
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if cf, ok := r.cases[id]; ok {
			cp := *cf
			return &cp, nil
		}
		return nil, fmt.Errorf("firestore get case file: not found")
	}
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
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if cf, ok := r.cases[caseID]; ok {
			cf.Documents = append(cf.Documents, doc)
			cf.UpdatedAt = time.Now()
			return nil
		}
		return fmt.Errorf("firestore update case file: not found")
	}
	_, err := r.client.Collection("case_files").Doc(caseID).Update(ctx, []firestore.Update{
		{Path: "documents", Value: firestore.ArrayUnion(doc)},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

func (r *FirestoreRepo) AddDocumentToCase(ctx context.Context, caseID string, doc domain.CaseDocument) error {
	return r.AddDocumentToCaseFile(ctx, caseID, doc)
}

// GetReportsByAssignedVolunteer returns reports assigned to a specific volunteer UID.
func (r *FirestoreRepo) GetReportsByAssignedVolunteer(ctx context.Context, uid string) ([]*domain.Report, error) {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var reports []*domain.Report
		for _, rpt := range r.reports {
			for _, a := range rpt.AssignedVolunteerIDs {
				if a == uid {
					cp := *rpt
					reports = append(reports, &cp)
					break
				}
			}
		}
		return reports, nil
	}
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

func (r *FirestoreRepo) GetCasesBySpecialist(ctx context.Context, uid string) ([]*domain.CaseFile, error) {
	return r.GetCaseFilesBySpecialist(ctx, uid)
}

// GetAllVolunteers retrieves all available volunteers.
func (r *FirestoreRepo) GetAllVolunteers(ctx context.Context) ([]*domain.Volunteer, error) {
	if r.isMemory() {
		r.mu.Lock()
		defer r.mu.Unlock()
		var volunteers []*domain.Volunteer
		for _, v := range r.volunteers {
			cp := *v
			volunteers = append(volunteers, &cp)
		}
		return volunteers, nil
	}
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

func (r *FirestoreRepo) GetVolunteer(ctx context.Context, uid string) (*domain.Volunteer, error) {
	return r.GetVolunteerByUID(ctx, uid)
}

func (r *FirestoreRepo) GetAllAvailableVolunteers(ctx context.Context) ([]*domain.Volunteer, error) {
	return r.GetAllVolunteers(ctx)
}

func (r *FirestoreRepo) GetReportsByVolunteer(ctx context.Context, uid string) ([]*domain.Report, error) {
	return r.GetReportsByAssignedVolunteer(ctx, uid)
}

func (r *FirestoreRepo) CountReportsByStatus(ctx context.Context, status domain.ReportStatus) (int, error) {
	iter := r.client.Collection("reports").Where("status", "==", status).Documents(ctx)
	defer iter.Stop()
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("firestore count by status: %w", err)
		}
		count++
	}
	return count, nil
}

func (r *FirestoreRepo) GetPrioritizedReports(ctx context.Context) ([]*domain.Report, error) {
	return r.GetAllReports(ctx, 0)
}

func (r *FirestoreRepo) GetRecentReportCountForWard(ctx context.Context, wardID string, hours int) (int, error) {
	reports, err := r.GetReportsByWard(ctx, wardID, time.Now().Add(-time.Duration(hours)*time.Hour))
	if err != nil {
		return 0, err
	}
	return len(reports), nil
}

func (r *FirestoreRepo) SaveCaseFile(ctx context.Context, cf *domain.CaseFile) error {
	_, err := r.CreateCaseFile(ctx, cf)
	return err
}

func applyReportUpdate(report *domain.Report, path string, value interface{}) {
	switch path {
	case "status":
		if s, ok := value.(string); ok {
			report.Status = domain.ReportStatus(s)
		}
	case "ward_id":
		if s, ok := value.(string); ok {
			report.WardID = s
		}
	case "summary":
		if s, ok := value.(string); ok {
			report.Summary = s
		}
	case "problem_category":
		if s, ok := value.(string); ok {
			report.ProblemCategory = s
		}
	case "severity_index":
		switch v := value.(type) {
		case float64:
			report.SeverityIndex = v
		case int:
			report.SeverityIndex = float64(v)
		}
	}
}
