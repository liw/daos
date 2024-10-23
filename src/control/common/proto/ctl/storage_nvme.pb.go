//
// (C) Copyright 2019-2023 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.5.0
// source: ctl/storage_nvme.proto

package ctl

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

// NvmeControllerResult represents state of operation performed on controller.
type NvmeControllerResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PciAddr  string         `protobuf:"bytes,1,opt,name=pci_addr,json=pciAddr,proto3" json:"pci_addr,omitempty"`     // PCI address of NVMe controller
	State    *ResponseState `protobuf:"bytes,2,opt,name=state,proto3" json:"state,omitempty"`                        // state of current operation
	RoleBits uint32         `protobuf:"varint,3,opt,name=role_bits,json=roleBits,proto3" json:"role_bits,omitempty"` // Device active roles (bitmask)
}

func (x *NvmeControllerResult) Reset() {
	*x = NvmeControllerResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_storage_nvme_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NvmeControllerResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NvmeControllerResult) ProtoMessage() {}

func (x *NvmeControllerResult) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_storage_nvme_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NvmeControllerResult.ProtoReflect.Descriptor instead.
func (*NvmeControllerResult) Descriptor() ([]byte, []int) {
	return file_ctl_storage_nvme_proto_rawDescGZIP(), []int{0}
}

func (x *NvmeControllerResult) GetPciAddr() string {
	if x != nil {
		return x.PciAddr
	}
	return ""
}

func (x *NvmeControllerResult) GetState() *ResponseState {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *NvmeControllerResult) GetRoleBits() uint32 {
	if x != nil {
		return x.RoleBits
	}
	return 0
}

type ScanNvmeReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Health    bool   `protobuf:"varint,1,opt,name=Health,proto3" json:"Health,omitempty"`       // Retrieve NVMe device health statistics
	Meta      bool   `protobuf:"varint,2,opt,name=Meta,proto3" json:"Meta,omitempty"`           // Retrieve metadata relating to NVMe device
	Basic     bool   `protobuf:"varint,3,opt,name=Basic,proto3" json:"Basic,omitempty"`         // Strip NVMe device details to only basic
	MetaSize  uint64 `protobuf:"varint,4,opt,name=MetaSize,proto3" json:"MetaSize,omitempty"`   // Size of the metadata blob
	RdbSize   uint64 `protobuf:"varint,5,opt,name=RdbSize,proto3" json:"RdbSize,omitempty"`     // Size of the RDB blob
	LinkStats bool   `protobuf:"varint,6,opt,name=LinkStats,proto3" json:"LinkStats,omitempty"` // Populate PCIe link info in health statistics
}

func (x *ScanNvmeReq) Reset() {
	*x = ScanNvmeReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_storage_nvme_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScanNvmeReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScanNvmeReq) ProtoMessage() {}

func (x *ScanNvmeReq) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_storage_nvme_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScanNvmeReq.ProtoReflect.Descriptor instead.
func (*ScanNvmeReq) Descriptor() ([]byte, []int) {
	return file_ctl_storage_nvme_proto_rawDescGZIP(), []int{1}
}

func (x *ScanNvmeReq) GetHealth() bool {
	if x != nil {
		return x.Health
	}
	return false
}

func (x *ScanNvmeReq) GetMeta() bool {
	if x != nil {
		return x.Meta
	}
	return false
}

func (x *ScanNvmeReq) GetBasic() bool {
	if x != nil {
		return x.Basic
	}
	return false
}

func (x *ScanNvmeReq) GetMetaSize() uint64 {
	if x != nil {
		return x.MetaSize
	}
	return 0
}

func (x *ScanNvmeReq) GetRdbSize() uint64 {
	if x != nil {
		return x.RdbSize
	}
	return 0
}

func (x *ScanNvmeReq) GetLinkStats() bool {
	if x != nil {
		return x.LinkStats
	}
	return false
}

type ScanNvmeResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ctrlrs []*NvmeController `protobuf:"bytes,1,rep,name=ctrlrs,proto3" json:"ctrlrs,omitempty"`
	State  *ResponseState    `protobuf:"bytes,2,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *ScanNvmeResp) Reset() {
	*x = ScanNvmeResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_storage_nvme_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScanNvmeResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScanNvmeResp) ProtoMessage() {}

func (x *ScanNvmeResp) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_storage_nvme_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScanNvmeResp.ProtoReflect.Descriptor instead.
func (*ScanNvmeResp) Descriptor() ([]byte, []int) {
	return file_ctl_storage_nvme_proto_rawDescGZIP(), []int{2}
}

func (x *ScanNvmeResp) GetCtrlrs() []*NvmeController {
	if x != nil {
		return x.Ctrlrs
	}
	return nil
}

func (x *ScanNvmeResp) GetState() *ResponseState {
	if x != nil {
		return x.State
	}
	return nil
}

type FormatNvmeReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FormatNvmeReq) Reset() {
	*x = FormatNvmeReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ctl_storage_nvme_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FormatNvmeReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FormatNvmeReq) ProtoMessage() {}

func (x *FormatNvmeReq) ProtoReflect() protoreflect.Message {
	mi := &file_ctl_storage_nvme_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FormatNvmeReq.ProtoReflect.Descriptor instead.
func (*FormatNvmeReq) Descriptor() ([]byte, []int) {
	return file_ctl_storage_nvme_proto_rawDescGZIP(), []int{3}
}

var File_ctl_storage_nvme_proto protoreflect.FileDescriptor

var file_ctl_storage_nvme_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x74, 0x6c, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x76,
	0x6d, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x63, 0x74, 0x6c, 0x1a, 0x10, 0x63,
	0x74, 0x6c, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0d, 0x63, 0x74, 0x6c, 0x2f, 0x73, 0x6d, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x78,
	0x0a, 0x14, 0x4e, 0x76, 0x6d, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x63, 0x69, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x63, 0x69, 0x41, 0x64, 0x64,
	0x72, 0x12, 0x28, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x63, 0x74, 0x6c, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x72,
	0x6f, 0x6c, 0x65, 0x5f, 0x62, 0x69, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08,
	0x72, 0x6f, 0x6c, 0x65, 0x42, 0x69, 0x74, 0x73, 0x22, 0xa3, 0x01, 0x0a, 0x0b, 0x53, 0x63, 0x61,
	0x6e, 0x4e, 0x76, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68,
	0x12, 0x12, 0x0a, 0x04, 0x4d, 0x65, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04,
	0x4d, 0x65, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x61, 0x73, 0x69, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x05, 0x42, 0x61, 0x73, 0x69, 0x63, 0x12, 0x1a, 0x0a, 0x08, 0x4d, 0x65,
	0x74, 0x61, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x4d, 0x65,
	0x74, 0x61, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x64, 0x62, 0x53, 0x69, 0x7a,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x52, 0x64, 0x62, 0x53, 0x69, 0x7a, 0x65,
	0x12, 0x1c, 0x0a, 0x09, 0x4c, 0x69, 0x6e, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x73, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x09, 0x4c, 0x69, 0x6e, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x73, 0x22, 0x65,
	0x0a, 0x0c, 0x53, 0x63, 0x61, 0x6e, 0x4e, 0x76, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x2b,
	0x0a, 0x06, 0x63, 0x74, 0x72, 0x6c, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x63, 0x74, 0x6c, 0x2e, 0x4e, 0x76, 0x6d, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x52, 0x06, 0x63, 0x74, 0x72, 0x6c, 0x72, 0x73, 0x12, 0x28, 0x0a, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x74, 0x6c,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x0f, 0x0a, 0x0d, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x4e,
	0x76, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x61, 0x6f, 0x73, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f,
	0x64, 0x61, 0x6f, 0x73, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x74,
	0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ctl_storage_nvme_proto_rawDescOnce sync.Once
	file_ctl_storage_nvme_proto_rawDescData = file_ctl_storage_nvme_proto_rawDesc
)

func file_ctl_storage_nvme_proto_rawDescGZIP() []byte {
	file_ctl_storage_nvme_proto_rawDescOnce.Do(func() {
		file_ctl_storage_nvme_proto_rawDescData = protoimpl.X.CompressGZIP(file_ctl_storage_nvme_proto_rawDescData)
	})
	return file_ctl_storage_nvme_proto_rawDescData
}

var file_ctl_storage_nvme_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_ctl_storage_nvme_proto_goTypes = []interface{}{
	(*NvmeControllerResult)(nil), // 0: ctl.NvmeControllerResult
	(*ScanNvmeReq)(nil),          // 1: ctl.ScanNvmeReq
	(*ScanNvmeResp)(nil),         // 2: ctl.ScanNvmeResp
	(*FormatNvmeReq)(nil),        // 3: ctl.FormatNvmeReq
	(*ResponseState)(nil),        // 4: ctl.ResponseState
	(*NvmeController)(nil),       // 5: ctl.NvmeController
}
var file_ctl_storage_nvme_proto_depIdxs = []int32{
	4, // 0: ctl.NvmeControllerResult.state:type_name -> ctl.ResponseState
	5, // 1: ctl.ScanNvmeResp.ctrlrs:type_name -> ctl.NvmeController
	4, // 2: ctl.ScanNvmeResp.state:type_name -> ctl.ResponseState
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_ctl_storage_nvme_proto_init() }
func file_ctl_storage_nvme_proto_init() {
	if File_ctl_storage_nvme_proto != nil {
		return
	}
	file_ctl_common_proto_init()
	file_ctl_smd_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ctl_storage_nvme_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NvmeControllerResult); i {
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
		file_ctl_storage_nvme_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScanNvmeReq); i {
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
		file_ctl_storage_nvme_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScanNvmeResp); i {
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
		file_ctl_storage_nvme_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FormatNvmeReq); i {
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
			RawDescriptor: file_ctl_storage_nvme_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ctl_storage_nvme_proto_goTypes,
		DependencyIndexes: file_ctl_storage_nvme_proto_depIdxs,
		MessageInfos:      file_ctl_storage_nvme_proto_msgTypes,
	}.Build()
	File_ctl_storage_nvme_proto = out.File
	file_ctl_storage_nvme_proto_rawDesc = nil
	file_ctl_storage_nvme_proto_goTypes = nil
	file_ctl_storage_nvme_proto_depIdxs = nil
}
