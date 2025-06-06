//
// (C) Copyright 2022 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v3.5.0
// source: chk/faults.proto

package chk

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Fault struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Class   CheckInconsistClass `protobuf:"varint,1,opt,name=class,proto3,enum=chk.CheckInconsistClass" json:"class,omitempty"`
	Strings []string            `protobuf:"bytes,2,rep,name=strings,proto3" json:"strings,omitempty"`
	Uints   []uint32            `protobuf:"varint,3,rep,packed,name=uints,proto3" json:"uints,omitempty"`
	Ints    []int32             `protobuf:"varint,4,rep,packed,name=ints,proto3" json:"ints,omitempty"`
}

func (x *Fault) Reset() {
	*x = Fault{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chk_faults_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Fault) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fault) ProtoMessage() {}

func (x *Fault) ProtoReflect() protoreflect.Message {
	mi := &file_chk_faults_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fault.ProtoReflect.Descriptor instead.
func (*Fault) Descriptor() ([]byte, []int) {
	return file_chk_faults_proto_rawDescGZIP(), []int{0}
}

func (x *Fault) GetClass() CheckInconsistClass {
	if x != nil {
		return x.Class
	}
	return CheckInconsistClass_CIC_NONE
}

func (x *Fault) GetStrings() []string {
	if x != nil {
		return x.Strings
	}
	return nil
}

func (x *Fault) GetUints() []uint32 {
	if x != nil {
		return x.Uints
	}
	return nil
}

func (x *Fault) GetInts() []int32 {
	if x != nil {
		return x.Ints
	}
	return nil
}

var File_chk_faults_proto protoreflect.FileDescriptor

var file_chk_faults_proto_rawDesc = []byte{
	0x0a, 0x10, 0x63, 0x68, 0x6b, 0x2f, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x03, 0x63, 0x68, 0x6b, 0x1a, 0x0d, 0x63, 0x68, 0x6b, 0x2f, 0x63, 0x68, 0x6b,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7b, 0x0a, 0x05, 0x46, 0x61, 0x75, 0x6c, 0x74, 0x12,
	0x2e, 0x0a, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18,
	0x2e, 0x63, 0x68, 0x6b, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x49, 0x6e, 0x63, 0x6f, 0x6e, 0x73,
	0x69, 0x73, 0x74, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x52, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x07, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x69, 0x6e,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x05, 0x75, 0x69, 0x6e, 0x74, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x05, 0x52, 0x04, 0x69,
	0x6e, 0x74, 0x73, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x64, 0x61, 0x6f, 0x73, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x64, 0x61, 0x6f,
	0x73, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x68, 0x6b, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chk_faults_proto_rawDescOnce sync.Once
	file_chk_faults_proto_rawDescData = file_chk_faults_proto_rawDesc
)

func file_chk_faults_proto_rawDescGZIP() []byte {
	file_chk_faults_proto_rawDescOnce.Do(func() {
		file_chk_faults_proto_rawDescData = protoimpl.X.CompressGZIP(file_chk_faults_proto_rawDescData)
	})
	return file_chk_faults_proto_rawDescData
}

var file_chk_faults_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_chk_faults_proto_goTypes = []interface{}{
	(*Fault)(nil),            // 0: chk.Fault
	(CheckInconsistClass)(0), // 1: chk.CheckInconsistClass
}
var file_chk_faults_proto_depIdxs = []int32{
	1, // 0: chk.Fault.class:type_name -> chk.CheckInconsistClass
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_chk_faults_proto_init() }
func file_chk_faults_proto_init() {
	if File_chk_faults_proto != nil {
		return
	}
	file_chk_chk_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_chk_faults_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Fault); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chk_faults_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chk_faults_proto_goTypes,
		DependencyIndexes: file_chk_faults_proto_depIdxs,
		MessageInfos:      file_chk_faults_proto_msgTypes,
	}.Build()
	File_chk_faults_proto = out.File
	file_chk_faults_proto_rawDesc = nil
	file_chk_faults_proto_goTypes = nil
	file_chk_faults_proto_depIdxs = nil
}
