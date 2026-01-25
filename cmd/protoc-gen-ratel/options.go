package main

import (
	"github.com/yaroher/ratel/ratelproto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// getRatelTableOptions extracts ratel.table option from a message
func getRatelTableOptions(msg *protogen.Message) *ratelproto.Table {
	if msg == nil {
		return nil
	}
	opts := msg.Desc.Options()
	if opts == nil {
		return nil
	}
	ext := proto.GetExtension(opts, ratelproto.E_Table)
	if ext == nil {
		return nil
	}
	return ext.(*ratelproto.Table)
}

// getRatelColumnOptions extracts ratel.column option from a field
func getRatelColumnOptions(field *protogen.Field) *ratelproto.Column {
	if field == nil {
		return nil
	}
	opts := field.Desc.Options()
	if opts == nil {
		return nil
	}
	ext := proto.GetExtension(opts, ratelproto.E_Column)
	if ext == nil {
		return nil
	}
	return ext.(*ratelproto.Column)
}

// getRatelRelationOptions extracts ratel.relation option from a field
func getRatelRelationOptions(field *protogen.Field) *ratelproto.Relation {
	if field == nil {
		return nil
	}
	opts := field.Desc.Options()
	if opts == nil {
		return nil
	}
	ext := proto.GetExtension(opts, ratelproto.E_Relation)
	if ext == nil {
		return nil
	}
	return ext.(*ratelproto.Relation)
}
