// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.2
// source: payment.proto

package paymentpb

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
	PaymentGatewayService_ProcessPayment_FullMethodName = "/payment.PaymentGatewayService/ProcessPayment"
)

// PaymentGatewayServiceClient is the client API for PaymentGatewayService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Service exposed by the Payment Gateway for clients.
type PaymentGatewayServiceClient interface {
	ProcessPayment(ctx context.Context, in *PaymentRequest, opts ...grpc.CallOption) (*PaymentResponse, error)
}

type paymentGatewayServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPaymentGatewayServiceClient(cc grpc.ClientConnInterface) PaymentGatewayServiceClient {
	return &paymentGatewayServiceClient{cc}
}

func (c *paymentGatewayServiceClient) ProcessPayment(ctx context.Context, in *PaymentRequest, opts ...grpc.CallOption) (*PaymentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PaymentResponse)
	err := c.cc.Invoke(ctx, PaymentGatewayService_ProcessPayment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PaymentGatewayServiceServer is the server API for PaymentGatewayService service.
// All implementations must embed UnimplementedPaymentGatewayServiceServer
// for forward compatibility.
//
// Service exposed by the Payment Gateway for clients.
type PaymentGatewayServiceServer interface {
	ProcessPayment(context.Context, *PaymentRequest) (*PaymentResponse, error)
	mustEmbedUnimplementedPaymentGatewayServiceServer()
}

// UnimplementedPaymentGatewayServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPaymentGatewayServiceServer struct{}

func (UnimplementedPaymentGatewayServiceServer) ProcessPayment(context.Context, *PaymentRequest) (*PaymentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessPayment not implemented")
}
func (UnimplementedPaymentGatewayServiceServer) mustEmbedUnimplementedPaymentGatewayServiceServer() {}
func (UnimplementedPaymentGatewayServiceServer) testEmbeddedByValue()                               {}

// UnsafePaymentGatewayServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PaymentGatewayServiceServer will
// result in compilation errors.
type UnsafePaymentGatewayServiceServer interface {
	mustEmbedUnimplementedPaymentGatewayServiceServer()
}

func RegisterPaymentGatewayServiceServer(s grpc.ServiceRegistrar, srv PaymentGatewayServiceServer) {
	// If the following call pancis, it indicates UnimplementedPaymentGatewayServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PaymentGatewayService_ServiceDesc, srv)
}

func _PaymentGatewayService_ProcessPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentGatewayServiceServer).ProcessPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PaymentGatewayService_ProcessPayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentGatewayServiceServer).ProcessPayment(ctx, req.(*PaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PaymentGatewayService_ServiceDesc is the grpc.ServiceDesc for PaymentGatewayService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PaymentGatewayService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentGatewayService",
	HandlerType: (*PaymentGatewayServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessPayment",
			Handler:    _PaymentGatewayService_ProcessPayment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}

const (
	BankService_PreparePayment_FullMethodName = "/payment.BankService/PreparePayment"
	BankService_CommitPayment_FullMethodName  = "/payment.BankService/CommitPayment"
	BankService_AbortPayment_FullMethodName   = "/payment.BankService/AbortPayment"
)

// BankServiceClient is the client API for BankService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Service implemented by Bank Servers.
type BankServiceClient interface {
	PreparePayment(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error)
	CommitPayment(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error)
	AbortPayment(ctx context.Context, in *AbortRequest, opts ...grpc.CallOption) (*AbortResponse, error)
}

type bankServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBankServiceClient(cc grpc.ClientConnInterface) BankServiceClient {
	return &bankServiceClient{cc}
}

func (c *bankServiceClient) PreparePayment(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PrepareResponse)
	err := c.cc.Invoke(ctx, BankService_PreparePayment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bankServiceClient) CommitPayment(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CommitResponse)
	err := c.cc.Invoke(ctx, BankService_CommitPayment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bankServiceClient) AbortPayment(ctx context.Context, in *AbortRequest, opts ...grpc.CallOption) (*AbortResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AbortResponse)
	err := c.cc.Invoke(ctx, BankService_AbortPayment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BankServiceServer is the server API for BankService service.
// All implementations must embed UnimplementedBankServiceServer
// for forward compatibility.
//
// Service implemented by Bank Servers.
type BankServiceServer interface {
	PreparePayment(context.Context, *PrepareRequest) (*PrepareResponse, error)
	CommitPayment(context.Context, *CommitRequest) (*CommitResponse, error)
	AbortPayment(context.Context, *AbortRequest) (*AbortResponse, error)
	mustEmbedUnimplementedBankServiceServer()
}

// UnimplementedBankServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBankServiceServer struct{}

func (UnimplementedBankServiceServer) PreparePayment(context.Context, *PrepareRequest) (*PrepareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreparePayment not implemented")
}
func (UnimplementedBankServiceServer) CommitPayment(context.Context, *CommitRequest) (*CommitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitPayment not implemented")
}
func (UnimplementedBankServiceServer) AbortPayment(context.Context, *AbortRequest) (*AbortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AbortPayment not implemented")
}
func (UnimplementedBankServiceServer) mustEmbedUnimplementedBankServiceServer() {}
func (UnimplementedBankServiceServer) testEmbeddedByValue()                     {}

// UnsafeBankServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BankServiceServer will
// result in compilation errors.
type UnsafeBankServiceServer interface {
	mustEmbedUnimplementedBankServiceServer()
}

func RegisterBankServiceServer(s grpc.ServiceRegistrar, srv BankServiceServer) {
	// If the following call pancis, it indicates UnimplementedBankServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&BankService_ServiceDesc, srv)
}

func _BankService_PreparePayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).PreparePayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_PreparePayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).PreparePayment(ctx, req.(*PrepareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BankService_CommitPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).CommitPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_CommitPayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).CommitPayment(ctx, req.(*CommitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BankService_AbortPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AbortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).AbortPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_AbortPayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).AbortPayment(ctx, req.(*AbortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BankService_ServiceDesc is the grpc.ServiceDesc for BankService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BankService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.BankService",
	HandlerType: (*BankServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PreparePayment",
			Handler:    _BankService_PreparePayment_Handler,
		},
		{
			MethodName: "CommitPayment",
			Handler:    _BankService_CommitPayment_Handler,
		},
		{
			MethodName: "AbortPayment",
			Handler:    _BankService_AbortPayment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}
