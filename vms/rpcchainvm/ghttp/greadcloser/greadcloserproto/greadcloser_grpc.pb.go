// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package greadcloserproto

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

// ReaderClient is the client API for Reader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReaderClient interface {
	Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error)
	Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseResponse, error)
}

type readerClient struct {
	cc grpc.ClientConnInterface
}

func NewReaderClient(cc grpc.ClientConnInterface) ReaderClient {
	return &readerClient{cc}
}

func (c *readerClient) Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error) {
	out := new(ReadResponse)
	err := c.cc.Invoke(ctx, "/greadcloserproto.Reader/Read", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *readerClient) Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseResponse, error) {
	out := new(CloseResponse)
	err := c.cc.Invoke(ctx, "/greadcloserproto.Reader/Close", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReaderServer is the server API for Reader service.
// All implementations must embed UnimplementedReaderServer
// for forward compatibility
type ReaderServer interface {
	Read(context.Context, *ReadRequest) (*ReadResponse, error)
	Close(context.Context, *CloseRequest) (*CloseResponse, error)
	mustEmbedUnimplementedReaderServer()
}

// UnimplementedReaderServer must be embedded to have forward compatible implementations.
type UnimplementedReaderServer struct {
}

func (UnimplementedReaderServer) Read(context.Context, *ReadRequest) (*ReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedReaderServer) Close(context.Context, *CloseRequest) (*CloseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedReaderServer) mustEmbedUnimplementedReaderServer() {}

// UnsafeReaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReaderServer will
// result in compilation errors.
type UnsafeReaderServer interface {
	mustEmbedUnimplementedReaderServer()
}

func RegisterReaderServer(s grpc.ServiceRegistrar, srv ReaderServer) {
	s.RegisterService(&Reader_ServiceDesc, srv)
}

func _Reader_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReaderServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/greadcloserproto.Reader/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReaderServer).Read(ctx, req.(*ReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Reader_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReaderServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/greadcloserproto.Reader/Close",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReaderServer).Close(ctx, req.(*CloseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Reader_ServiceDesc is the grpc.ServiceDesc for Reader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Reader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "greadcloserproto.Reader",
	HandlerType: (*ReaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Read",
			Handler:    _Reader_Read_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _Reader_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "greadcloser.proto",
}
