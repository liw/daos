//
// (C) Copyright 2020-2022 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v3.5.0
// source: ctl/ranks.proto

package ctl

import (
	shared "github.com/daos-stack/daos/src/control/common/proto/shared"
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

// Generic request indicating which ranks to operate on.
// Used in gRPC fanout to operate on hosts with multiple ranks.
type RanksReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Force     bool   `protobuf:"varint,3,opt,name=force,proto3" json:"force,omitempty"`                          // force operation
	Ranks     string `protobuf:"bytes,4,opt,name=ranks,proto3" json:"ranks,omitempty"`                           // rankset to operate over
	CheckMode bool   `protobuf:"varint,5,opt,name=check_mode,json=checkMode,proto3" json:"check_mode,omitempty"` // start in check mode
}

func (x *RanksReq) Reset() {
	*x = RanksReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_ranks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RanksReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RanksReq) ProtoMessage() {}

func (x *RanksReq) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_ranks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RanksReq.ProtoReflect.Descriptor instead.
func (*RanksReq) Descriptor() ([]byte, []int) {
	return file_ctl_ranks_proto_rawDescGZIP(), []int{0}
}

func (x *RanksReq) GetForce() bool {
	if x != nil {
		return x.Force
	}
	return false
}

func (x *RanksReq) GetRanks() string {
	if x != nil {
		return x.Ranks
	}
	return ""
}

func (x *RanksReq) GetCheckMode() bool {
	if x != nil {
		return x.CheckMode
	}
	return false
}

// Generic response containing DER result from multiple ranks.
// Used in gRPC fanout to operate on hosts with multiple ranks.
type RanksResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Results []*shared.RankResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}

func (x *RanksResp) Reset() {
	*x = RanksResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_ranks_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RanksResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RanksResp) ProtoMessage() {}

func (x *RanksResp) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_ranks_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RanksResp.ProtoReflect.Descriptor instead.
func (*RanksResp) Descriptor() ([]byte, []int) {
	return file_ctl_ranks_proto_rawDescGZIP(), []int{1}
}

func (x *RanksResp) GetResults() []*shared.RankResult {
	if x != nil {
		return x.Results
	}
	return nil
}

var File_ctl_ranks_proto protoreflect.FileDescriptor

var file_ctl_ranks_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x74, 0x6c, 0x2f, 0x72, 0x61, 0x6e, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x63, 0x74, 0x6c, 0x1a, 0x12, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2f, 0x72,
	0x61, 0x6e, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x55, 0x0a, 0x08, 0x52, 0x61,
	0x6e, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x72, 0x61, 0x6e, 0x6b, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x61, 0x6e,
	0x6b, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x4d, 0x6f, 0x64,
	0x65, 0x22, 0x39, 0x0a, 0x09, 0x52, 0x61, 0x6e, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x2c,
	0x0a, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x52, 0x61, 0x6e, 0x6b, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x42, 0x39, 0x5a, 0x37,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x61, 0x6f, 0x73, 0x2d,
	0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x64, 0x61, 0x6f, 0x73, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x74, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ctl_ranks_proto_rawDescOnce sync.Once
	file_ctl_ranks_proto_rawDescData = file_ctl_ranks_proto_rawDesc
)

func file_ctl_ranks_proto_rawDescGZIP() []byte {
	file_ctl_ranks_proto_rawDescOnce.Do(func() {
		file_ctl_ranks_proto_rawDescData = protoimpl.X.CompressGZIP(file_ctl_ranks_proto_rawDescData)
	})
	return file_ctl_ranks_proto_rawDescData
}

var file_ctl_ranks_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_ctl_ranks_proto_goTypes = []interface{}{
	(*RanksReq)(nil),          // 0: ctl.RanksReq
	(*RanksResp)(nil),         // 1: ctl.RanksResp
	(*shared.RankResult)(nil), // 2: shared.RankResult
}
var file_ctl_ranks_proto_depIdxs = []int32{
	2, // 0: ctl.RanksResp.results:type_name -> shared.RankResult
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ctl_ranks_proto_init() }
func file_ctl_ranks_proto_init() {
	if File_ctl_ranks_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ctl_ranks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RanksReq); i {
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
		file_ctl_ranks_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RanksResp); i {
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
			RawDescriptor: file_ctl_ranks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ctl_ranks_proto_goTypes,
		DependencyIndexes: file_ctl_ranks_proto_depIdxs,
		MessageInfos:      file_ctl_ranks_proto_msgTypes,
	}.Build()
	File_ctl_ranks_proto = out.File
	file_ctl_ranks_proto_rawDesc = nil
	file_ctl_ranks_proto_goTypes = nil
	file_ctl_ranks_proto_depIdxs = nil
}
