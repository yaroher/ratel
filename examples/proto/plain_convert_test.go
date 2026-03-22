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
