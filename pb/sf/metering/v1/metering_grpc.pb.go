// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.0
// source: sf/metering/v1/metering.proto

package pbmetering

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MeteringClient is the client API for Metering service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MeteringClient interface {
	Emit(ctx context.Context, in *Events, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type meteringClient struct {
	cc grpc.ClientConnInterface
}

func NewMeteringClient(cc grpc.ClientConnInterface) MeteringClient {
	return &meteringClient{cc}
}

func (c *meteringClient) Emit(ctx context.Context, in *Events, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/sf.metering.v1.Metering/Emit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MeteringServer is the server API for Metering service.
// All implementations should embed UnimplementedMeteringServer
// for forward compatibility
type MeteringServer interface {
	Emit(context.Context, *Events) (*emptypb.Empty, error)
}

// UnimplementedMeteringServer should be embedded to have forward compatible implementations.
type UnimplementedMeteringServer struct {
}

func (UnimplementedMeteringServer) Emit(context.Context, *Events) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Emit not implemented")
}

// UnsafeMeteringServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MeteringServer will
// result in compilation errors.
type UnsafeMeteringServer interface {
	mustEmbedUnimplementedMeteringServer()
}

func RegisterMeteringServer(s grpc.ServiceRegistrar, srv MeteringServer) {
	s.RegisterService(&Metering_ServiceDesc, srv)
}

func _Metering_Emit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Events)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeteringServer).Emit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sf.metering.v1.Metering/Emit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeteringServer).Emit(ctx, req.(*Events))
	}
	return interceptor(ctx, in, info, handler)
}

// Metering_ServiceDesc is the grpc.ServiceDesc for Metering service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metering_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sf.metering.v1.Metering",
	HandlerType: (*MeteringServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Emit",
			Handler:    _Metering_Emit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sf/metering/v1/metering.proto",
}
