package main

import (
	"github.com/yaroher/protoc-gen-go-plain/generator"
	"github.com/yaroher/protoc-gen-go-plain/goplain"
)

// getRatelTypeOverrides returns type overrides for protoc-gen-go-plain
// These convert protobuf well-known types to PostgreSQL-compatible Go types
func getRatelTypeOverrides() []*goplain.TypeOverride {
	return []*goplain.TypeOverride{
		// google.protobuf.Timestamp -> time.Time
		{
			Selector: &goplain.OverrideSelector{
				FieldTypeUrl: strPtr("google.protobuf.Timestamp"),
			},
			TargetGoType: &goplain.GoIdent{
				Name:       "Time",
				ImportPath: "time",
			},
		},
		// google.protobuf.Duration -> time.Duration
		{
			Selector: &goplain.OverrideSelector{
				FieldTypeUrl: strPtr("google.protobuf.Duration"),
			},
			TargetGoType: &goplain.GoIdent{
				Name:       "Duration",
				ImportPath: "time",
			},
		},
		// google.protobuf.Struct -> json.RawMessage (for JSONB)
		{
			Selector: &goplain.OverrideSelector{
				FieldTypeUrl: strPtr("google.protobuf.Struct"),
			},
			TargetGoType: &goplain.GoIdent{
				Name:       "RawMessage",
				ImportPath: "encoding/json",
			},
		},
		// google.protobuf.Int64Value -> *int64
		{
			Selector: &goplain.OverrideSelector{
				FieldTypeUrl: strPtr("google.protobuf.Int64Value"),
			},
			TargetGoType: &goplain.GoIdent{
				Name:       "int64",
				ImportPath: "",
			},
		},
	}
}

// getRatelCasters returns existing casters for protoc-gen-go-plain
// These provide conversion functions between protobuf and Go types
func getRatelCasters() []*generator.ExistingCaster {
	return []*generator.ExistingCaster{
		// timestamppb.Timestamp -> time.Time
		{
			SourceType: generator.GoType{
				Name:       "Timestamp",
				ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Time",
				ImportPath: "time",
			},
			CasterIdent: generator.GoIdent{
				Name:       "TimestampToTime",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// time.Time -> timestamppb.Timestamp
		{
			SourceType: generator.GoType{
				Name:       "Time",
				ImportPath: "time",
			},
			TargetType: generator.GoType{
				Name:       "Timestamp",
				ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "TimeToTimestamp",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// durationpb.Duration -> time.Duration
		{
			SourceType: generator.GoType{
				Name:       "Duration",
				ImportPath: "google.golang.org/protobuf/types/known/durationpb",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Duration",
				ImportPath: "time",
			},
			CasterIdent: generator.GoIdent{
				Name:       "DurationPbToDuration",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// time.Duration -> durationpb.Duration
		{
			SourceType: generator.GoType{
				Name:       "Duration",
				ImportPath: "time",
			},
			TargetType: generator.GoType{
				Name:       "Duration",
				ImportPath: "google.golang.org/protobuf/types/known/durationpb",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "DurationToDurationPb",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// wrapperspb.Int64Value -> int64
		{
			SourceType: generator.GoType{
				Name:       "Int64Value",
				ImportPath: "google.golang.org/protobuf/types/known/wrapperspb",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "int64",
				ImportPath: "",
			},
			CasterIdent: generator.GoIdent{
				Name:       "Int64ValueToInt64",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// int64 -> wrapperspb.Int64Value
		{
			SourceType: generator.GoType{
				Name:       "int64",
				ImportPath: "",
			},
			TargetType: generator.GoType{
				Name:       "Int64Value",
				ImportPath: "google.golang.org/protobuf/types/known/wrapperspb",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "Int64ToInt64Value",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// *timestamppb.Timestamp -> *time.Time (nullable)
		{
			SourceType: generator.GoType{
				Name:       "Timestamp",
				ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Time",
				ImportPath: "time",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "TimestampToNullableTime",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// *time.Time -> *timestamppb.Timestamp (nullable)
		{
			SourceType: generator.GoType{
				Name:       "Time",
				ImportPath: "time",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Timestamp",
				ImportPath: "google.golang.org/protobuf/types/known/timestamppb",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "NullableTimeToTimestamp",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// *durationpb.Duration -> *time.Duration (nullable)
		{
			SourceType: generator.GoType{
				Name:       "Duration",
				ImportPath: "google.golang.org/protobuf/types/known/durationpb",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Duration",
				ImportPath: "time",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "DurationPbToNullableDuration",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
		// *time.Duration -> *durationpb.Duration (nullable)
		{
			SourceType: generator.GoType{
				Name:       "Duration",
				ImportPath: "time",
				IsPointer:  true,
			},
			TargetType: generator.GoType{
				Name:       "Duration",
				ImportPath: "google.golang.org/protobuf/types/known/durationpb",
				IsPointer:  true,
			},
			CasterIdent: generator.GoIdent{
				Name:       "NullableDurationToDurationPb",
				ImportPath: "github.com/yaroher/ratel/pkg/ratelcast",
			},
			IsFunc: true,
		},
	}
}

func strPtr(s string) *string {
	return &s
}
