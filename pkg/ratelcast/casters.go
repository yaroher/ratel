package ratelcast

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimestampToTime converts *timestamppb.Timestamp to time.Time
func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// TimeToTimestamp converts time.Time to *timestamppb.Timestamp
func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// DurationPbToDuration converts *durationpb.Duration to time.Duration
func DurationPbToDuration(d *durationpb.Duration) time.Duration {
	if d == nil {
		return 0
	}
	return d.AsDuration()
}

// DurationToDurationPb converts time.Duration to *durationpb.Duration
func DurationToDurationPb(d time.Duration) *durationpb.Duration {
	if d == 0 {
		return nil
	}
	return durationpb.New(d)
}

// NullableInt64 converts *int64 to int64 (for wrappers)
func NullableInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// Int64ToNullable converts int64 to *int64
func Int64ToNullable(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// NullableString converts *string to string
func NullableString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// StringToNullable converts string to *string
func StringToNullable(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// NullableBool converts *bool to bool
func NullableBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// BoolToNullable converts bool to *bool
func BoolToNullable(v bool) *bool {
	return &v
}
