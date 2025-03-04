// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.2
// source: mapreduce.proto

package mrpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MasterService_RequestMapTask_FullMethodName    = "/mr.MasterService/RequestMapTask"
	MasterService_ReportMapTask_FullMethodName     = "/mr.MasterService/ReportMapTask"
	MasterService_RequestReduceTask_FullMethodName = "/mr.MasterService/RequestReduceTask"
	MasterService_ReportReduceTask_FullMethodName  = "/mr.MasterService/ReportReduceTask"
)

// MasterServiceClient is the client API for MasterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The Master service that coordinates MapReduce tasks.
type MasterServiceClient interface {
	RequestMapTask(ctx context.Context, in *MapTaskRequest, opts ...grpc.CallOption) (*MapTaskResponse, error)
	ReportMapTask(ctx context.Context, in *ReportMapTaskRequest, opts ...grpc.CallOption) (*ReportMapTaskResponse, error)
	RequestReduceTask(ctx context.Context, in *ReduceTaskRequest, opts ...grpc.CallOption) (*ReduceTaskResponse, error)
	ReportReduceTask(ctx context.Context, in *ReportReduceTaskRequest, opts ...grpc.CallOption) (*ReportReduceTaskResponse, error)
}

type masterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMasterServiceClient(cc grpc.ClientConnInterface) MasterServiceClient {
	return &masterServiceClient{cc}
}

func (c *masterServiceClient) RequestMapTask(ctx context.Context, in *MapTaskRequest, opts ...grpc.CallOption) (*MapTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MapTaskResponse)
	err := c.cc.Invoke(ctx, MasterService_RequestMapTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterServiceClient) ReportMapTask(ctx context.Context, in *ReportMapTaskRequest, opts ...grpc.CallOption) (*ReportMapTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReportMapTaskResponse)
	err := c.cc.Invoke(ctx, MasterService_ReportMapTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterServiceClient) RequestReduceTask(ctx context.Context, in *ReduceTaskRequest, opts ...grpc.CallOption) (*ReduceTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReduceTaskResponse)
	err := c.cc.Invoke(ctx, MasterService_RequestReduceTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterServiceClient) ReportReduceTask(ctx context.Context, in *ReportReduceTaskRequest, opts ...grpc.CallOption) (*ReportReduceTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReportReduceTaskResponse)
	err := c.cc.Invoke(ctx, MasterService_ReportReduceTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MasterServiceServer is the server API for MasterService service.
// All implementations must embed UnimplementedMasterServiceServer
// for forward compatibility.
//
// The Master service that coordinates MapReduce tasks.
type MasterServiceServer interface {
	RequestMapTask(context.Context, *MapTaskRequest) (*MapTaskResponse, error)
	ReportMapTask(context.Context, *ReportMapTaskRequest) (*ReportMapTaskResponse, error)
	RequestReduceTask(context.Context, *ReduceTaskRequest) (*ReduceTaskResponse, error)
	ReportReduceTask(context.Context, *ReportReduceTaskRequest) (*ReportReduceTaskResponse, error)
	mustEmbedUnimplementedMasterServiceServer()
}

// UnimplementedMasterServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMasterServiceServer struct{}

func (UnimplementedMasterServiceServer) RequestMapTask(context.Context, *MapTaskRequest) (*MapTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestMapTask not implemented")
}
func (UnimplementedMasterServiceServer) ReportMapTask(context.Context, *ReportMapTaskRequest) (*ReportMapTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportMapTask not implemented")
}
func (UnimplementedMasterServiceServer) RequestReduceTask(context.Context, *ReduceTaskRequest) (*ReduceTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestReduceTask not implemented")
}
func (UnimplementedMasterServiceServer) ReportReduceTask(context.Context, *ReportReduceTaskRequest) (*ReportReduceTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportReduceTask not implemented")
}
func (UnimplementedMasterServiceServer) mustEmbedUnimplementedMasterServiceServer() {}
func (UnimplementedMasterServiceServer) testEmbeddedByValue()                       {}

// UnsafeMasterServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MasterServiceServer will
// result in compilation errors.
type UnsafeMasterServiceServer interface {
	mustEmbedUnimplementedMasterServiceServer()
}

func RegisterMasterServiceServer(s grpc.ServiceRegistrar, srv MasterServiceServer) {
	// If the following call pancis, it indicates UnimplementedMasterServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MasterService_ServiceDesc, srv)
}

func _MasterService_RequestMapTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MapTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServiceServer).RequestMapTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MasterService_RequestMapTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServiceServer).RequestMapTask(ctx, req.(*MapTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MasterService_ReportMapTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportMapTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServiceServer).ReportMapTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MasterService_ReportMapTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServiceServer).ReportMapTask(ctx, req.(*ReportMapTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MasterService_RequestReduceTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReduceTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServiceServer).RequestReduceTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MasterService_RequestReduceTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServiceServer).RequestReduceTask(ctx, req.(*ReduceTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MasterService_ReportReduceTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportReduceTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServiceServer).ReportReduceTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MasterService_ReportReduceTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServiceServer).ReportReduceTask(ctx, req.(*ReportReduceTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MasterService_ServiceDesc is the grpc.ServiceDesc for MasterService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MasterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mr.MasterService",
	HandlerType: (*MasterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestMapTask",
			Handler:    _MasterService_RequestMapTask_Handler,
		},
		{
			MethodName: "ReportMapTask",
			Handler:    _MasterService_ReportMapTask_Handler,
		},
		{
			MethodName: "RequestReduceTask",
			Handler:    _MasterService_RequestReduceTask_Handler,
		},
		{
			MethodName: "ReportReduceTask",
			Handler:    _MasterService_ReportReduceTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mapreduce.proto",
}
