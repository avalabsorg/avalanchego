// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.15.8
// source: gconn.proto

package gconnproto

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

type ReadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length int32 `protobuf:"varint,1,opt,name=length,proto3" json:"length,omitempty"`
}

func (x *ReadRequest) Reset() {
	*x = ReadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadRequest) ProtoMessage() {}

func (x *ReadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadRequest.ProtoReflect.Descriptor instead.
func (*ReadRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{0}
}

func (x *ReadRequest) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type ReadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Read    []byte `protobuf:"bytes,1,opt,name=read,proto3" json:"read,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Errored bool   `protobuf:"varint,3,opt,name=errored,proto3" json:"errored,omitempty"`
}

func (x *ReadResponse) Reset() {
	*x = ReadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadResponse) ProtoMessage() {}

func (x *ReadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadResponse.ProtoReflect.Descriptor instead.
func (*ReadResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{1}
}

func (x *ReadResponse) GetRead() []byte {
	if x != nil {
		return x.Read
	}
	return nil
}

func (x *ReadResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *ReadResponse) GetErrored() bool {
	if x != nil {
		return x.Errored
	}
	return false
}

type WriteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *WriteRequest) Reset() {
	*x = WriteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteRequest) ProtoMessage() {}

func (x *WriteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteRequest.ProtoReflect.Descriptor instead.
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{2}
}

func (x *WriteRequest) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type WriteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length  int32  `protobuf:"varint,1,opt,name=length,proto3" json:"length,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Errored bool   `protobuf:"varint,3,opt,name=errored,proto3" json:"errored,omitempty"`
}

func (x *WriteResponse) Reset() {
	*x = WriteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteResponse) ProtoMessage() {}

func (x *WriteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteResponse.ProtoReflect.Descriptor instead.
func (*WriteResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{3}
}

func (x *WriteResponse) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *WriteResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *WriteResponse) GetErrored() bool {
	if x != nil {
		return x.Errored
	}
	return false
}

type CloseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CloseRequest) Reset() {
	*x = CloseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseRequest) ProtoMessage() {}

func (x *CloseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseRequest.ProtoReflect.Descriptor instead.
func (*CloseRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{4}
}

type CloseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CloseResponse) Reset() {
	*x = CloseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseResponse) ProtoMessage() {}

func (x *CloseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseResponse.ProtoReflect.Descriptor instead.
func (*CloseResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{5}
}

type SetDeadlineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time []byte `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *SetDeadlineRequest) Reset() {
	*x = SetDeadlineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetDeadlineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetDeadlineRequest) ProtoMessage() {}

func (x *SetDeadlineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetDeadlineRequest.ProtoReflect.Descriptor instead.
func (*SetDeadlineRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{6}
}

func (x *SetDeadlineRequest) GetTime() []byte {
	if x != nil {
		return x.Time
	}
	return nil
}

type SetDeadlineResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetDeadlineResponse) Reset() {
	*x = SetDeadlineResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetDeadlineResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetDeadlineResponse) ProtoMessage() {}

func (x *SetDeadlineResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetDeadlineResponse.ProtoReflect.Descriptor instead.
func (*SetDeadlineResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{7}
}

type SetReadDeadlineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time []byte `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *SetReadDeadlineRequest) Reset() {
	*x = SetReadDeadlineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetReadDeadlineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetReadDeadlineRequest) ProtoMessage() {}

func (x *SetReadDeadlineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetReadDeadlineRequest.ProtoReflect.Descriptor instead.
func (*SetReadDeadlineRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{8}
}

func (x *SetReadDeadlineRequest) GetTime() []byte {
	if x != nil {
		return x.Time
	}
	return nil
}

type SetReadDeadlineResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetReadDeadlineResponse) Reset() {
	*x = SetReadDeadlineResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetReadDeadlineResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetReadDeadlineResponse) ProtoMessage() {}

func (x *SetReadDeadlineResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetReadDeadlineResponse.ProtoReflect.Descriptor instead.
func (*SetReadDeadlineResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{9}
}

type SetWriteDeadlineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time []byte `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *SetWriteDeadlineRequest) Reset() {
	*x = SetWriteDeadlineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetWriteDeadlineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetWriteDeadlineRequest) ProtoMessage() {}

func (x *SetWriteDeadlineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetWriteDeadlineRequest.ProtoReflect.Descriptor instead.
func (*SetWriteDeadlineRequest) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{10}
}

func (x *SetWriteDeadlineRequest) GetTime() []byte {
	if x != nil {
		return x.Time
	}
	return nil
}

type SetWriteDeadlineResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetWriteDeadlineResponse) Reset() {
	*x = SetWriteDeadlineResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gconn_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetWriteDeadlineResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetWriteDeadlineResponse) ProtoMessage() {}

func (x *SetWriteDeadlineResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gconn_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetWriteDeadlineResponse.ProtoReflect.Descriptor instead.
func (*SetWriteDeadlineResponse) Descriptor() ([]byte, []int) {
	return file_gconn_proto_rawDescGZIP(), []int{11}
}

var File_gconn_proto protoreflect.FileDescriptor

var file_gconn_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x67,
	0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x25, 0x0a, 0x0b, 0x52, 0x65, 0x61,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67,
	0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68,
	0x22, 0x52, 0x0a, 0x0c, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x72, 0x65, 0x61, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x65, 0x64, 0x22, 0x28, 0x0a, 0x0c, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x57,
	0x0a, 0x0d, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x65, 0x64, 0x22, 0x0e, 0x0a, 0x0c, 0x43, 0x6c, 0x6f, 0x73, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0f, 0x0a, 0x0d, 0x43, 0x6c, 0x6f, 0x73, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x28, 0x0a, 0x12, 0x53, 0x65, 0x74, 0x44,
	0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x74, 0x69,
	0x6d, 0x65, 0x22, 0x15, 0x0a, 0x13, 0x53, 0x65, 0x74, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2c, 0x0a, 0x16, 0x53, 0x65, 0x74,
	0x52, 0x65, 0x61, 0x64, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x19, 0x0a, 0x17, 0x53, 0x65, 0x74, 0x52, 0x65,
	0x61, 0x64, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x2d, 0x0a, 0x17, 0x53, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x44, 0x65,
	0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x22, 0x1a, 0x0a, 0x18, 0x53, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x44, 0x65, 0x61,
	0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xc8, 0x03,
	0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x6e, 0x12, 0x39, 0x0a, 0x04, 0x52, 0x65, 0x61, 0x64, 0x12, 0x17,
	0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x61, 0x64,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x3c, 0x0a, 0x05, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x18, 0x2e, 0x67, 0x63, 0x6f,
	0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x3c, 0x0a, 0x05, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x18, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x19, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4e, 0x0a,
	0x0b, 0x53, 0x65, 0x74, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x1e, 0x2e, 0x67,
	0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x74, 0x44, 0x65, 0x61,
	0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x67,
	0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x74, 0x44, 0x65, 0x61,
	0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5a, 0x0a,
	0x0f, 0x53, 0x65, 0x74, 0x52, 0x65, 0x61, 0x64, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65,
	0x12, 0x22, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65,
	0x74, 0x52, 0x65, 0x61, 0x64, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x53, 0x65, 0x74, 0x52, 0x65, 0x61, 0x64, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x10, 0x53, 0x65, 0x74,
	0x57, 0x72, 0x69, 0x74, 0x65, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x23, 0x2e,
	0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x74, 0x57, 0x72,
	0x69, 0x74, 0x65, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x24, 0x2e, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x53, 0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x65, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x76, 0x61, 0x2d, 0x6c, 0x61, 0x62, 0x73, 0x2f,
	0x61, 0x76, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x67, 0x6f, 0x2f, 0x72, 0x70, 0x63, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x76, 0x6d, 0x2f, 0x67, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x67, 0x63, 0x6f,
	0x6e, 0x6e, 0x2f, 0x67, 0x63, 0x6f, 0x6e, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gconn_proto_rawDescOnce sync.Once
	file_gconn_proto_rawDescData = file_gconn_proto_rawDesc
)

func file_gconn_proto_rawDescGZIP() []byte {
	file_gconn_proto_rawDescOnce.Do(func() {
		file_gconn_proto_rawDescData = protoimpl.X.CompressGZIP(file_gconn_proto_rawDescData)
	})
	return file_gconn_proto_rawDescData
}

var file_gconn_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_gconn_proto_goTypes = []interface{}{
	(*ReadRequest)(nil),              // 0: gconnproto.ReadRequest
	(*ReadResponse)(nil),             // 1: gconnproto.ReadResponse
	(*WriteRequest)(nil),             // 2: gconnproto.WriteRequest
	(*WriteResponse)(nil),            // 3: gconnproto.WriteResponse
	(*CloseRequest)(nil),             // 4: gconnproto.CloseRequest
	(*CloseResponse)(nil),            // 5: gconnproto.CloseResponse
	(*SetDeadlineRequest)(nil),       // 6: gconnproto.SetDeadlineRequest
	(*SetDeadlineResponse)(nil),      // 7: gconnproto.SetDeadlineResponse
	(*SetReadDeadlineRequest)(nil),   // 8: gconnproto.SetReadDeadlineRequest
	(*SetReadDeadlineResponse)(nil),  // 9: gconnproto.SetReadDeadlineResponse
	(*SetWriteDeadlineRequest)(nil),  // 10: gconnproto.SetWriteDeadlineRequest
	(*SetWriteDeadlineResponse)(nil), // 11: gconnproto.SetWriteDeadlineResponse
}
var file_gconn_proto_depIdxs = []int32{
	0,  // 0: gconnproto.Conn.Read:input_type -> gconnproto.ReadRequest
	2,  // 1: gconnproto.Conn.Write:input_type -> gconnproto.WriteRequest
	4,  // 2: gconnproto.Conn.Close:input_type -> gconnproto.CloseRequest
	6,  // 3: gconnproto.Conn.SetDeadline:input_type -> gconnproto.SetDeadlineRequest
	8,  // 4: gconnproto.Conn.SetReadDeadline:input_type -> gconnproto.SetReadDeadlineRequest
	10, // 5: gconnproto.Conn.SetWriteDeadline:input_type -> gconnproto.SetWriteDeadlineRequest
	1,  // 6: gconnproto.Conn.Read:output_type -> gconnproto.ReadResponse
	3,  // 7: gconnproto.Conn.Write:output_type -> gconnproto.WriteResponse
	5,  // 8: gconnproto.Conn.Close:output_type -> gconnproto.CloseResponse
	7,  // 9: gconnproto.Conn.SetDeadline:output_type -> gconnproto.SetDeadlineResponse
	9,  // 10: gconnproto.Conn.SetReadDeadline:output_type -> gconnproto.SetReadDeadlineResponse
	11, // 11: gconnproto.Conn.SetWriteDeadline:output_type -> gconnproto.SetWriteDeadlineResponse
	6,  // [6:12] is the sub-list for method output_type
	0,  // [0:6] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_gconn_proto_init() }
func file_gconn_proto_init() {
	if File_gconn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gconn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadRequest); i {
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
		file_gconn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadResponse); i {
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
		file_gconn_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteRequest); i {
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
		file_gconn_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteResponse); i {
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
		file_gconn_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseRequest); i {
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
		file_gconn_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseResponse); i {
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
		file_gconn_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetDeadlineRequest); i {
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
		file_gconn_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetDeadlineResponse); i {
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
		file_gconn_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetReadDeadlineRequest); i {
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
		file_gconn_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetReadDeadlineResponse); i {
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
		file_gconn_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetWriteDeadlineRequest); i {
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
		file_gconn_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetWriteDeadlineResponse); i {
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
			RawDescriptor: file_gconn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_gconn_proto_goTypes,
		DependencyIndexes: file_gconn_proto_depIdxs,
		MessageInfos:      file_gconn_proto_msgTypes,
	}.Build()
	File_gconn_proto = out.File
	file_gconn_proto_rawDesc = nil
	file_gconn_proto_goTypes = nil
	file_gconn_proto_depIdxs = nil
}
