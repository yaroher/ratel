package storepb

import "testing"

// TestIntoPlainRepeatedFieldsNonNil verifies that IntoPlain() returns non-nil
// slices for empty repeated fields, so database drivers send empty arrays
// instead of NULL for NOT NULL array columns.
func TestIntoPlainRepeatedFieldsNonNil(t *testing.T) {
	// AuditLog has Tags ([]string, NOT NULL) and RelatedIds ([]int64, NOT NULL)
	pb := &AuditLog{
		Id:         1,
		Action:     "test",
		EntityType: "user",
		EntityId:   42,
		// Tags and RelatedIds intentionally left empty (nil in proto)
	}

	plain := pb.IntoPlain()

	if plain.Tags == nil {
		t.Error("IntoPlain().Tags is nil, expected empty non-nil slice for NOT NULL array column")
	}
	if plain.RelatedIds == nil {
		t.Error("IntoPlain().RelatedIds is nil, expected empty non-nil slice for NOT NULL array column")
	}
	if len(plain.Tags) != 0 {
		t.Errorf("expected empty Tags, got %v", plain.Tags)
	}
	if len(plain.RelatedIds) != 0 {
		t.Errorf("expected empty RelatedIds, got %v", plain.RelatedIds)
	}
}

// TestIntoPlainRepeatedFieldsPreserved verifies that non-empty repeated fields
// are correctly passed through IntoPlain().
func TestIntoPlainRepeatedFieldsPreserved(t *testing.T) {
	pb := &AuditLog{
		Id:         1,
		Action:     "test",
		EntityType: "user",
		EntityId:   42,
		Tags:       []string{"admin", "security"},
		RelatedIds: []int64{100, 200},
	}

	plain := pb.IntoPlain()

	if len(plain.Tags) != 2 || plain.Tags[0] != "admin" || plain.Tags[1] != "security" {
		t.Errorf("expected Tags=[admin security], got %v", plain.Tags)
	}
	if len(plain.RelatedIds) != 2 || plain.RelatedIds[0] != 100 || plain.RelatedIds[1] != 200 {
		t.Errorf("expected RelatedIds=[100 200], got %v", plain.RelatedIds)
	}
}

// TestIntoPlainRepeatedMessageFieldsNonNil verifies that repeated message fields
// (embed path via generateEmbedFieldAssignment) also produce non-nil slices.
func TestIntoPlainRepeatedMessageFieldsNonNil(t *testing.T) {
	// Product has Categories ([]CategoryScanner) and Tags ([]TagScanner)
	// These go through the embed repeated message code path.
	pb := &Product{
		Sku:  "TEST-001",
		Name: "Test Product",
		// Categories and Tags intentionally left empty
	}

	plain := pb.IntoPlain()

	if plain.Categories == nil {
		t.Error("IntoPlain().Categories is nil, expected empty non-nil slice")
	}
	if plain.Tags == nil {
		t.Error("IntoPlain().Tags is nil, expected empty non-nil slice")
	}
	if len(plain.Categories) != 0 {
		t.Errorf("expected empty Categories, got %v", plain.Categories)
	}
	if len(plain.Tags) != 0 {
		t.Errorf("expected empty Tags, got %v", plain.Tags)
	}
}

// TestIntoPlainRepeatedMessageFieldsPreserved verifies non-empty repeated message
// fields are correctly converted through IntoPlain().
func TestIntoPlainRepeatedMessageFieldsPreserved(t *testing.T) {
	pb := &Product{
		Sku:  "TEST-001",
		Name: "Test Product",
		Categories: []*Category{
			{Name: "Electronics"},
			{Name: "Gadgets"},
		},
		Tags: []*Tag{
			{Name: "sale"},
		},
	}

	plain := pb.IntoPlain()

	if len(plain.Categories) != 2 {
		t.Fatalf("expected 2 Categories, got %d", len(plain.Categories))
	}
	if plain.Categories[0].Name != "Electronics" || plain.Categories[1].Name != "Gadgets" {
		t.Errorf("unexpected Categories: %+v", plain.Categories)
	}
	if len(plain.Tags) != 1 || plain.Tags[0].Name != "sale" {
		t.Errorf("unexpected Tags: %+v", plain.Tags)
	}
}

// TestIntoPlainEnumAsString verifies that enum fields are converted to string
// in IntoPlain() so pgx can scan TEXT columns directly.
func TestIntoPlainEnumAsString(t *testing.T) {
	tests := []struct {
		name     string
		severity AuditSeverity
		want     string
	}{
		{"unspecified", AuditSeverity_AUDIT_SEVERITY_UNSPECIFIED, "AUDIT_SEVERITY_UNSPECIFIED"},
		{"low", AuditSeverity_AUDIT_SEVERITY_LOW, "AUDIT_SEVERITY_LOW"},
		{"medium", AuditSeverity_AUDIT_SEVERITY_MEDIUM, "AUDIT_SEVERITY_MEDIUM"},
		{"high", AuditSeverity_AUDIT_SEVERITY_HIGH, "AUDIT_SEVERITY_HIGH"},
		{"critical", AuditSeverity_AUDIT_SEVERITY_CRITICAL, "AUDIT_SEVERITY_CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := &AuditLog{
				Id:       1,
				Action:   "test",
				Severity: tt.severity,
			}
			plain := pb.IntoPlain()

			if plain.Severity != tt.want {
				t.Errorf("IntoPlain().Severity = %q, want %q", plain.Severity, tt.want)
			}
		})
	}
}

// TestIntoPbEnumFromString verifies that IntoPb() converts string back to enum.
func TestIntoPbEnumFromString(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want AuditSeverity
	}{
		{"low", "AUDIT_SEVERITY_LOW", AuditSeverity_AUDIT_SEVERITY_LOW},
		{"critical", "AUDIT_SEVERITY_CRITICAL", AuditSeverity_AUDIT_SEVERITY_CRITICAL},
		{"unknown", "UNKNOWN_VALUE", AuditSeverity_AUDIT_SEVERITY_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plain := &AuditLogScanner{
				Id:       1,
				Action:   "test",
				Severity: tt.str,
			}
			pb := plain.IntoPb()

			if pb.Severity != tt.want {
				t.Errorf("IntoPb().Severity = %v, want %v", pb.Severity, tt.want)
			}
		})
	}
}

// TestIntoPlainRepeatedEnumAsString verifies repeated enum → []string conversion.
func TestIntoPlainRepeatedEnumAsString(t *testing.T) {
	pb := &AuditLog{
		Id:     1,
		Action: "test",
		AffectedLevels: []AuditSeverity{
			AuditSeverity_AUDIT_SEVERITY_LOW,
			AuditSeverity_AUDIT_SEVERITY_CRITICAL,
		},
	}
	plain := pb.IntoPlain()

	if len(plain.AffectedLevels) != 2 {
		t.Fatalf("expected 2 affected levels, got %d", len(plain.AffectedLevels))
	}
	if plain.AffectedLevels[0] != "AUDIT_SEVERITY_LOW" {
		t.Errorf("AffectedLevels[0] = %q, want AUDIT_SEVERITY_LOW", plain.AffectedLevels[0])
	}
	if plain.AffectedLevels[1] != "AUDIT_SEVERITY_CRITICAL" {
		t.Errorf("AffectedLevels[1] = %q, want AUDIT_SEVERITY_CRITICAL", plain.AffectedLevels[1])
	}
}

// TestIntoPlainRepeatedEnumEmpty verifies empty repeated enum → non-nil []string.
func TestIntoPlainRepeatedEnumEmpty(t *testing.T) {
	pb := &AuditLog{Id: 1, Action: "test"}
	plain := pb.IntoPlain()

	if plain.AffectedLevels == nil {
		t.Error("AffectedLevels is nil, expected empty non-nil slice")
	}
	if len(plain.AffectedLevels) != 0 {
		t.Errorf("expected empty AffectedLevels, got %v", plain.AffectedLevels)
	}
}

// TestIntoPbRepeatedEnumFromString verifies []string → repeated enum conversion.
func TestIntoPbRepeatedEnumFromString(t *testing.T) {
	plain := &AuditLogScanner{
		Id:             1,
		Action:         "test",
		AffectedLevels: []string{"AUDIT_SEVERITY_MEDIUM", "AUDIT_SEVERITY_HIGH"},
	}
	pb := plain.IntoPb()

	if len(pb.AffectedLevels) != 2 {
		t.Fatalf("expected 2 affected levels, got %d", len(pb.AffectedLevels))
	}
	if pb.AffectedLevels[0] != AuditSeverity_AUDIT_SEVERITY_MEDIUM {
		t.Errorf("AffectedLevels[0] = %v, want MEDIUM", pb.AffectedLevels[0])
	}
	if pb.AffectedLevels[1] != AuditSeverity_AUDIT_SEVERITY_HIGH {
		t.Errorf("AffectedLevels[1] = %v, want HIGH", pb.AffectedLevels[1])
	}
}
