// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: appsender/appsender.proto

package appsender

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SendAppRequestMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The nodes to send this request to
	NodeIds [][]byte `protobuf:"bytes,1,rep,name=node_ids,json=nodeIds,proto3" json:"node_ids,omitempty"`
	// The ID of this request
	RequestId uint32 `protobuf:"varint,2,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The request body
	Request []byte `protobuf:"bytes,3,opt,name=request,proto3" json:"request,omitempty"`
}

func (x *SendAppRequestMsg) Reset() {
	*x = SendAppRequestMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendAppRequestMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendAppRequestMsg) ProtoMessage() {}

func (x *SendAppRequestMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendAppRequestMsg.ProtoReflect.Descriptor instead.
func (*SendAppRequestMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{0}
}

func (x *SendAppRequestMsg) GetNodeIds() [][]byte {
	if x != nil {
		return x.NodeIds
	}
	return nil
}

func (x *SendAppRequestMsg) GetRequestId() uint32 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *SendAppRequestMsg) GetRequest() []byte {
	if x != nil {
		return x.Request
	}
	return nil
}

type SendAppResponseMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The node to send a response to
	NodeId []byte `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	// ID of this request
	RequestId uint32 `protobuf:"varint,2,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The response body
	Response []byte `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *SendAppResponseMsg) Reset() {
	*x = SendAppResponseMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendAppResponseMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendAppResponseMsg) ProtoMessage() {}

func (x *SendAppResponseMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendAppResponseMsg.ProtoReflect.Descriptor instead.
func (*SendAppResponseMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{1}
}

func (x *SendAppResponseMsg) GetNodeId() []byte {
	if x != nil {
		return x.NodeId
	}
	return nil
}

func (x *SendAppResponseMsg) GetRequestId() uint32 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *SendAppResponseMsg) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

type SendAppGossipMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The message body
	Msg []byte `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *SendAppGossipMsg) Reset() {
	*x = SendAppGossipMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendAppGossipMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendAppGossipMsg) ProtoMessage() {}

func (x *SendAppGossipMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendAppGossipMsg.ProtoReflect.Descriptor instead.
func (*SendAppGossipMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{2}
}

func (x *SendAppGossipMsg) GetMsg() []byte {
	if x != nil {
		return x.Msg
	}
	return nil
}

type SendAppGossipSpecificMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The nodes to send this request to
	NodeIds [][]byte `protobuf:"bytes,1,rep,name=node_ids,json=nodeIds,proto3" json:"node_ids,omitempty"`
	// The message body
	Msg []byte `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *SendAppGossipSpecificMsg) Reset() {
	*x = SendAppGossipSpecificMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendAppGossipSpecificMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendAppGossipSpecificMsg) ProtoMessage() {}

func (x *SendAppGossipSpecificMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendAppGossipSpecificMsg.ProtoReflect.Descriptor instead.
func (*SendAppGossipSpecificMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{3}
}

func (x *SendAppGossipSpecificMsg) GetNodeIds() [][]byte {
	if x != nil {
		return x.NodeIds
	}
	return nil
}

func (x *SendAppGossipSpecificMsg) GetMsg() []byte {
	if x != nil {
		return x.Msg
	}
	return nil
}

type SendCrossChainAppRequestMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The chain to send this request to
	ChainId []byte `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	// the ID of this request
	RequestId uint32 `protobuf:"varint,2,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The request body
	Request []byte `protobuf:"bytes,3,opt,name=request,proto3" json:"request,omitempty"`
}

func (x *SendCrossChainAppRequestMsg) Reset() {
	*x = SendCrossChainAppRequestMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendCrossChainAppRequestMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendCrossChainAppRequestMsg) ProtoMessage() {}

func (x *SendCrossChainAppRequestMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendCrossChainAppRequestMsg.ProtoReflect.Descriptor instead.
func (*SendCrossChainAppRequestMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{4}
}

func (x *SendCrossChainAppRequestMsg) GetChainId() []byte {
	if x != nil {
		return x.ChainId
	}
	return nil
}

func (x *SendCrossChainAppRequestMsg) GetRequestId() uint32 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *SendCrossChainAppRequestMsg) GetRequest() []byte {
	if x != nil {
		return x.Request
	}
	return nil
}

type SendCrossChainAppResponseMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The chain this request was sent from
	ChainId []byte `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	// the ID of this request
	RequestId uint32 `protobuf:"varint,2,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The response body
	Response []byte `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *SendCrossChainAppResponseMsg) Reset() {
	*x = SendCrossChainAppResponseMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_appsender_appsender_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendCrossChainAppResponseMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendCrossChainAppResponseMsg) ProtoMessage() {}

func (x *SendCrossChainAppResponseMsg) ProtoReflect() protoreflect.Message {
	mi := &file_appsender_appsender_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendCrossChainAppResponseMsg.ProtoReflect.Descriptor instead.
func (*SendCrossChainAppResponseMsg) Descriptor() ([]byte, []int) {
	return file_appsender_appsender_proto_rawDescGZIP(), []int{5}
}

func (x *SendCrossChainAppResponseMsg) GetChainId() []byte {
	if x != nil {
		return x.ChainId
	}
	return nil
}

func (x *SendCrossChainAppResponseMsg) GetRequestId() uint32 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *SendCrossChainAppResponseMsg) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

var File_appsender_appsender_proto protoreflect.FileDescriptor

var file_appsender_appsender_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x70, 0x73,
	0x65, 0x6e, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x61, 0x70, 0x70,
	0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x67, 0x0a, 0x11, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65,
	0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x6e, 0x6f, 0x64, 0x65,
	0x49, 0x64, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x68, 0x0a, 0x12,
	0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d,
	0x73, 0x67, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x24, 0x0a, 0x10, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70,
	0x70, 0x47, 0x6f, 0x73, 0x73, 0x69, 0x70, 0x4d, 0x73, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x47, 0x0a, 0x18,
	0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x47, 0x6f, 0x73, 0x73, 0x69, 0x70, 0x53, 0x70, 0x65,
	0x63, 0x69, 0x66, 0x69, 0x63, 0x4d, 0x73, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65,
	0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x6e, 0x6f, 0x64, 0x65,
	0x49, 0x64, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x71, 0x0a, 0x1b, 0x53, 0x65, 0x6e, 0x64, 0x43, 0x72, 0x6f,
	0x73, 0x73, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x4d, 0x73, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x12,
	0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x18,
	0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x74, 0x0a, 0x1c, 0x53, 0x65, 0x6e, 0x64,
	0x43, 0x72, 0x6f, 0x73, 0x73, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x70, 0x70, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x73, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xf3,
	0x03, 0x0a, 0x09, 0x41, 0x70, 0x70, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x46, 0x0a, 0x0e,
	0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c,
	0x2e, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x41,
	0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x12, 0x48, 0x0a, 0x0f, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d, 0x2e, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x4d, 0x73, 0x67, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x44,
	0x0a, 0x0d, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x47, 0x6f, 0x73, 0x73, 0x69, 0x70, 0x12,
	0x1b, 0x2e, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x6e, 0x64,
	0x41, 0x70, 0x70, 0x47, 0x6f, 0x73, 0x73, 0x69, 0x70, 0x4d, 0x73, 0x67, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x12, 0x54, 0x0a, 0x15, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x47,
	0x6f, 0x73, 0x73, 0x69, 0x70, 0x53, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x63, 0x12, 0x23, 0x2e,
	0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x70,
	0x70, 0x47, 0x6f, 0x73, 0x73, 0x69, 0x70, 0x53, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x63, 0x4d,
	0x73, 0x67, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x5a, 0x0a, 0x18, 0x53, 0x65,
	0x6e, 0x64, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x70, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x2e, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x43, 0x68, 0x61, 0x69,
	0x6e, 0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x1a, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x5c, 0x0a, 0x19, 0x53, 0x65, 0x6e, 0x64, 0x43, 0x72,
	0x6f, 0x73, 0x73, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x27, 0x2e, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x2e,
	0x53, 0x65, 0x6e, 0x64, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x70,
	0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x73, 0x67, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x61, 0x76, 0x61, 0x2d, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x61, 0x76, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x68, 0x65, 0x67, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x62,
	0x2f, 0x61, 0x70, 0x70, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_appsender_appsender_proto_rawDescOnce sync.Once
	file_appsender_appsender_proto_rawDescData = file_appsender_appsender_proto_rawDesc
)

func file_appsender_appsender_proto_rawDescGZIP() []byte {
	file_appsender_appsender_proto_rawDescOnce.Do(func() {
		file_appsender_appsender_proto_rawDescData = protoimpl.X.CompressGZIP(file_appsender_appsender_proto_rawDescData)
	})
	return file_appsender_appsender_proto_rawDescData
}

var file_appsender_appsender_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_appsender_appsender_proto_goTypes = []interface{}{
	(*SendAppRequestMsg)(nil),            // 0: appsender.SendAppRequestMsg
	(*SendAppResponseMsg)(nil),           // 1: appsender.SendAppResponseMsg
	(*SendAppGossipMsg)(nil),             // 2: appsender.SendAppGossipMsg
	(*SendAppGossipSpecificMsg)(nil),     // 3: appsender.SendAppGossipSpecificMsg
	(*SendCrossChainAppRequestMsg)(nil),  // 4: appsender.SendCrossChainAppRequestMsg
	(*SendCrossChainAppResponseMsg)(nil), // 5: appsender.SendCrossChainAppResponseMsg
	(*emptypb.Empty)(nil),                // 6: google.protobuf.Empty
}
var file_appsender_appsender_proto_depIdxs = []int32{
	0, // 0: appsender.AppSender.SendAppRequest:input_type -> appsender.SendAppRequestMsg
	1, // 1: appsender.AppSender.SendAppResponse:input_type -> appsender.SendAppResponseMsg
	2, // 2: appsender.AppSender.SendAppGossip:input_type -> appsender.SendAppGossipMsg
	3, // 3: appsender.AppSender.SendAppGossipSpecific:input_type -> appsender.SendAppGossipSpecificMsg
	4, // 4: appsender.AppSender.SendCrossChainAppRequest:input_type -> appsender.SendCrossChainAppRequestMsg
	5, // 5: appsender.AppSender.SendCrossChainAppResponse:input_type -> appsender.SendCrossChainAppResponseMsg
	6, // 6: appsender.AppSender.SendAppRequest:output_type -> google.protobuf.Empty
	6, // 7: appsender.AppSender.SendAppResponse:output_type -> google.protobuf.Empty
	6, // 8: appsender.AppSender.SendAppGossip:output_type -> google.protobuf.Empty
	6, // 9: appsender.AppSender.SendAppGossipSpecific:output_type -> google.protobuf.Empty
	6, // 10: appsender.AppSender.SendCrossChainAppRequest:output_type -> google.protobuf.Empty
	6, // 11: appsender.AppSender.SendCrossChainAppResponse:output_type -> google.protobuf.Empty
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_appsender_appsender_proto_init() }
func file_appsender_appsender_proto_init() {
	if File_appsender_appsender_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_appsender_appsender_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendAppRequestMsg); i {
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
		file_appsender_appsender_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendAppResponseMsg); i {
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
		file_appsender_appsender_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendAppGossipMsg); i {
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
		file_appsender_appsender_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendAppGossipSpecificMsg); i {
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
		file_appsender_appsender_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendCrossChainAppRequestMsg); i {
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
		file_appsender_appsender_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendCrossChainAppResponseMsg); i {
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
			RawDescriptor: file_appsender_appsender_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_appsender_appsender_proto_goTypes,
		DependencyIndexes: file_appsender_appsender_proto_depIdxs,
		MessageInfos:      file_appsender_appsender_proto_msgTypes,
	}.Build()
	File_appsender_appsender_proto = out.File
	file_appsender_appsender_proto_rawDesc = nil
	file_appsender_appsender_proto_goTypes = nil
	file_appsender_appsender_proto_depIdxs = nil
}
