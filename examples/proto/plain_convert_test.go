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
