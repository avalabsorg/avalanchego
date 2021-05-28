// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.15.8
// source: rpcdb.proto

package rpcdbproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DatabaseClient is the client API for Database service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DatabaseClient interface {
	Has(ctx context.Context, in *HasRequest, opts ...grpc.CallOption) (*HasResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	Stat(ctx context.Context, in *StatRequest, opts ...grpc.CallOption) (*StatResponse, error)
	Compact(ctx context.Context, in *CompactRequest, opts ...grpc.CallOption) (*CompactResponse, error)
	Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseResponse, error)
	WriteBatch(ctx context.Context, in *WriteBatchRequest, opts ...grpc.CallOption) (*WriteBatchResponse, error)
	NewIteratorWithStartAndPrefix(ctx context.Context, in *NewIteratorWithStartAndPrefixRequest, opts ...grpc.CallOption) (*NewIteratorWithStartAndPrefixResponse, error)
	IteratorNext(ctx context.Context, in *IteratorNextRequest, opts ...grpc.CallOption) (*IteratorNextResponse, error)
	IteratorError(ctx context.Context, in *IteratorErrorRequest, opts ...grpc.CallOption) (*IteratorErrorResponse, error)
	IteratorRelease(ctx context.Context, in *IteratorReleaseRequest, opts ...grpc.CallOption) (*IteratorReleaseResponse, error)
}

type databaseClient struct {
	cc grpc.ClientConnInterface
}

func NewDatabaseClient(cc grpc.ClientConnInterface) DatabaseClient {
	return &databaseClient{cc}
}

func (c *databaseClient) Has(ctx context.Context, in *HasRequest, opts ...grpc.CallOption) (*HasResponse, error) {
	out := new(HasResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Has", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Put", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Stat(ctx context.Context, in *StatRequest, opts ...grpc.CallOption) (*StatResponse, error) {
	out := new(StatResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Stat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Compact(ctx context.Context, in *CompactRequest, opts ...grpc.CallOption) (*CompactResponse, error) {
	out := new(CompactResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Compact", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseResponse, error) {
	out := new(CloseResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/Close", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) WriteBatch(ctx context.Context, in *WriteBatchRequest, opts ...grpc.CallOption) (*WriteBatchResponse, error) {
	out := new(WriteBatchResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/WriteBatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) NewIteratorWithStartAndPrefix(ctx context.Context, in *NewIteratorWithStartAndPrefixRequest, opts ...grpc.CallOption) (*NewIteratorWithStartAndPrefixResponse, error) {
	out := new(NewIteratorWithStartAndPrefixResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/NewIteratorWithStartAndPrefix", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) IteratorNext(ctx context.Context, in *IteratorNextRequest, opts ...grpc.CallOption) (*IteratorNextResponse, error) {
	out := new(IteratorNextResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/IteratorNext", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) IteratorError(ctx context.Context, in *IteratorErrorRequest, opts ...grpc.CallOption) (*IteratorErrorResponse, error) {
	out := new(IteratorErrorResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/IteratorError", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *databaseClient) IteratorRelease(ctx context.Context, in *IteratorReleaseRequest, opts ...grpc.CallOption) (*IteratorReleaseResponse, error) {
	out := new(IteratorReleaseResponse)
	err := c.cc.Invoke(ctx, "/rpcdbproto.Database/IteratorRelease", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DatabaseServer is the server API for Database service.
// All implementations must embed UnimplementedDatabaseServer
// for forward compatibility
type DatabaseServer interface {
	Has(context.Context, *HasRequest) (*HasResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Put(context.Context, *PutRequest) (*PutResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Stat(context.Context, *StatRequest) (*StatResponse, error)
	Compact(context.Context, *CompactRequest) (*CompactResponse, error)
	Close(context.Context, *CloseRequest) (*CloseResponse, error)
	WriteBatch(context.Context, *WriteBatchRequest) (*WriteBatchResponse, error)
	NewIteratorWithStartAndPrefix(context.Context, *NewIteratorWithStartAndPrefixRequest) (*NewIteratorWithStartAndPrefixResponse, error)
	IteratorNext(context.Context, *IteratorNextRequest) (*IteratorNextResponse, error)
	IteratorError(context.Context, *IteratorErrorRequest) (*IteratorErrorResponse, error)
	IteratorRelease(context.Context, *IteratorReleaseRequest) (*IteratorReleaseResponse, error)
	mustEmbedUnimplementedDatabaseServer()
}

// UnimplementedDatabaseServer must be embedded to have forward compatible implementations.
type UnimplementedDatabaseServer struct {
}

func (UnimplementedDatabaseServer) Has(context.Context, *HasRequest) (*HasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Has not implemented")
}
func (UnimplementedDatabaseServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedDatabaseServer) Put(context.Context, *PutRequest) (*PutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (UnimplementedDatabaseServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedDatabaseServer) Stat(context.Context, *StatRequest) (*StatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (UnimplementedDatabaseServer) Compact(context.Context, *CompactRequest) (*CompactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Compact not implemented")
}
func (UnimplementedDatabaseServer) Close(context.Context, *CloseRequest) (*CloseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedDatabaseServer) WriteBatch(context.Context, *WriteBatchRequest) (*WriteBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WriteBatch not implemented")
}
func (UnimplementedDatabaseServer) NewIteratorWithStartAndPrefix(context.Context, *NewIteratorWithStartAndPrefixRequest) (*NewIteratorWithStartAndPrefixResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewIteratorWithStartAndPrefix not implemented")
}
func (UnimplementedDatabaseServer) IteratorNext(context.Context, *IteratorNextRequest) (*IteratorNextResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorNext not implemented")
}
func (UnimplementedDatabaseServer) IteratorError(context.Context, *IteratorErrorRequest) (*IteratorErrorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorError not implemented")
}
func (UnimplementedDatabaseServer) IteratorRelease(context.Context, *IteratorReleaseRequest) (*IteratorReleaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorRelease not implemented")
}
func (UnimplementedDatabaseServer) mustEmbedUnimplementedDatabaseServer() {}

// UnsafeDatabaseServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DatabaseServer will
// result in compilation errors.
type UnsafeDatabaseServer interface {
	mustEmbedUnimplementedDatabaseServer()
}

func RegisterDatabaseServer(s grpc.ServiceRegistrar, srv DatabaseServer) {
	s.RegisterService(&Database_ServiceDesc, srv)
}

func _Database_Has_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Has(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Has",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Has(ctx, req.(*HasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Stat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Stat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Stat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Stat(ctx, req.(*StatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Compact_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompactRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Compact(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Compact",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Compact(ctx, req.(*CompactRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/Close",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).Close(ctx, req.(*CloseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_WriteBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).WriteBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/WriteBatch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).WriteBatch(ctx, req.(*WriteBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_NewIteratorWithStartAndPrefix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewIteratorWithStartAndPrefixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).NewIteratorWithStartAndPrefix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/NewIteratorWithStartAndPrefix",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).NewIteratorWithStartAndPrefix(ctx, req.(*NewIteratorWithStartAndPrefixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_IteratorNext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorNextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).IteratorNext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/IteratorNext",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).IteratorNext(ctx, req.(*IteratorNextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_IteratorError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorErrorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).IteratorError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/IteratorError",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).IteratorError(ctx, req.(*IteratorErrorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Database_IteratorRelease_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatabaseServer).IteratorRelease(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcdbproto.Database/IteratorRelease",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatabaseServer).IteratorRelease(ctx, req.(*IteratorReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Database_ServiceDesc is the grpc.ServiceDesc for Database service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Database_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpcdbproto.Database",
	HandlerType: (*DatabaseServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Has",
			Handler:    _Database_Has_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Database_Get_Handler,
		},
		{
			MethodName: "Put",
			Handler:    _Database_Put_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Database_Delete_Handler,
		},
		{
			MethodName: "Stat",
			Handler:    _Database_Stat_Handler,
		},
		{
			MethodName: "Compact",
			Handler:    _Database_Compact_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _Database_Close_Handler,
		},
		{
			MethodName: "WriteBatch",
			Handler:    _Database_WriteBatch_Handler,
		},
		{
			MethodName: "NewIteratorWithStartAndPrefix",
			Handler:    _Database_NewIteratorWithStartAndPrefix_Handler,
		},
		{
			MethodName: "IteratorNext",
			Handler:    _Database_IteratorNext_Handler,
		},
		{
			MethodName: "IteratorError",
			Handler:    _Database_IteratorError_Handler,
		},
		{
			MethodName: "IteratorRelease",
			Handler:    _Database_IteratorRelease_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpcdb.proto",
}
