package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/yaroher/ratel/ratelproto"
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
	t, _ := ext.(*ratelproto.Table)
	return t
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
	c, _ := ext.(*ratelproto.Column)
	return c
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
	r, _ := ext.(*ratelproto.Relation)
	return r
}
