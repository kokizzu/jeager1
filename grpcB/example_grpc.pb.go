// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: example.proto

package grpcB

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

// GrpcBClient is the client API for GrpcB service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GrpcBClient interface {
	GetSomething(ctx context.Context, in *GetSomethingRequest, opts ...grpc.CallOption) (*GetSomethingResponse, error)
	PostAnything(ctx context.Context, in *PostAnythingRequest, opts ...grpc.CallOption) (*PostAnythingResponse, error)
}

type grpcBClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcBClient(cc grpc.ClientConnInterface) GrpcBClient {
	return &grpcBClient{cc}
}

func (c *grpcBClient) GetSomething(ctx context.Context, in *GetSomethingRequest, opts ...grpc.CallOption) (*GetSomethingResponse, error) {
	out := new(GetSomethingResponse)
	err := c.cc.Invoke(ctx, "/GrpcB/GetSomething", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcBClient) PostAnything(ctx context.Context, in *PostAnythingRequest, opts ...grpc.CallOption) (*PostAnythingResponse, error) {
	out := new(PostAnythingResponse)
	err := c.cc.Invoke(ctx, "/GrpcB/PostAnything", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GrpcBServer is the server API for GrpcB service.
// All implementations must embed UnimplementedGrpcBServer
// for forward compatibility
type GrpcBServer interface {
	GetSomething(context.Context, *GetSomethingRequest) (*GetSomethingResponse, error)
	PostAnything(context.Context, *PostAnythingRequest) (*PostAnythingResponse, error)
	mustEmbedUnimplementedGrpcBServer()
}

// UnimplementedGrpcBServer must be embedded to have forward compatible implementations.
type UnimplementedGrpcBServer struct {
}

func (UnimplementedGrpcBServer) GetSomething(context.Context, *GetSomethingRequest) (*GetSomethingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSomething not implemented")
}
func (UnimplementedGrpcBServer) PostAnything(context.Context, *PostAnythingRequest) (*PostAnythingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostAnything not implemented")
}
func (UnimplementedGrpcBServer) mustEmbedUnimplementedGrpcBServer() {}

// UnsafeGrpcBServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GrpcBServer will
// result in compilation errors.
type UnsafeGrpcBServer interface {
	mustEmbedUnimplementedGrpcBServer()
}

func RegisterGrpcBServer(s grpc.ServiceRegistrar, srv GrpcBServer) {
	s.RegisterService(&GrpcB_ServiceDesc, srv)
}

func _GrpcB_GetSomething_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSomethingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcBServer).GetSomething(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GrpcB/GetSomething",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcBServer).GetSomething(ctx, req.(*GetSomethingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcB_PostAnything_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostAnythingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcBServer).PostAnything(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GrpcB/PostAnything",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcBServer).PostAnything(ctx, req.(*PostAnythingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GrpcB_ServiceDesc is the grpc.ServiceDesc for GrpcB service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GrpcB_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "GrpcB",
	HandlerType: (*GrpcBServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSomething",
			Handler:    _GrpcB_GetSomething_Handler,
		},
		{
			MethodName: "PostAnything",
			Handler:    _GrpcB_PostAnything_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example.proto",
}
