package ratelcast

import (
	"encoding/json"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

// Int64ValueToInt64 converts *wrapperspb.Int64Value to int64
func Int64ValueToInt64(v *wrapperspb.Int64Value) int64 {
	if v == nil {
		return 0
	}
	return v.GetValue()
}

// Int64ToInt64Value converts int64 to *wrapperspb.Int64Value
func Int64ToInt64Value(v int64) *wrapperspb.Int64Value {
	return wrapperspb.Int64(v)
}

// TimestampToNullableTime converts *timestamppb.Timestamp to *time.Time
func TimestampToNullableTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// NullableTimeToTimestamp converts *time.Time to *timestamppb.Timestamp
func NullableTimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// DurationPbToNullableDuration converts *durationpb.Duration to *time.Duration
func DurationPbToNullableDuration(d *durationpb.Duration) *time.Duration {
	if d == nil {
		return nil
	}
	dur := d.AsDuration()
	return &dur
}

// NullableDurationToDurationPb converts *time.Duration to *durationpb.Duration
func NullableDurationToDurationPb(d *time.Duration) *durationpb.Duration {
	if d == nil {
		return nil
	}
	return durationpb.New(*d)
}

// StructToRawMessage converts *structpb.Struct to json.RawMessage via JSON marshaling
func StructToRawMessage(s *structpb.Struct) json.RawMessage {
	if s == nil {
		return nil
	}
	data, err := protojson.Marshal(s)
	if err != nil {
		return nil
	}
	return data
}

// RawMessageToStruct converts json.RawMessage to *structpb.Struct via JSON unmarshaling
func RawMessageToStruct(data json.RawMessage) *structpb.Struct {
	if len(data) == 0 {
		return nil
	}
	s := &structpb.Struct{}
	if err := protojson.Unmarshal(data, s); err != nil {
		return nil
	}
	return s
}
