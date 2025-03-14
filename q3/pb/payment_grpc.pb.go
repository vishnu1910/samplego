// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.2
// source: payment.proto

package pb

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
	PaymentGateway_RegisterClient_FullMethodName = "/payment.PaymentGateway/RegisterClient"
	PaymentGateway_ProcessPayment_FullMethodName = "/payment.PaymentGateway/ProcessPayment"
	PaymentGateway_GetBalance_FullMethodName     = "/payment.PaymentGateway/GetBalance"
)

// PaymentGatewayClient is the client API for PaymentGateway service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ------------------ Payment Gateway Service ------------------
type PaymentGatewayClient interface {
	// Client registration (authentication/authorization)
	RegisterClient(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	// Process a payment using 2PC
	ProcessPayment(ctx context.Context, in *PaymentRequest, opts ...grpc.CallOption) (*PaymentResponse, error)
	// Query client balance
	GetBalance(ctx context.Context, in *BalanceRequest, opts ...grpc.CallOption) (*BalanceResponse, error)
}

type paymentGatewayClient struct {
	cc grpc.ClientConnInterface
}

func NewPaymentGatewayClient(cc grpc.ClientConnInterface) PaymentGatewayClient {
	return &paymentGatewayClient{cc}
}

func (c *paymentGatewayClient) RegisterClient(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, PaymentGateway_RegisterClient_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentGatewayClient) ProcessPayment(ctx context.Context, in *PaymentRequest, opts ...grpc.CallOption) (*PaymentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PaymentResponse)
	err := c.cc.Invoke(ctx, PaymentGateway_ProcessPayment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentGatewayClient) GetBalance(ctx context.Context, in *BalanceRequest, opts ...grpc.CallOption) (*BalanceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BalanceResponse)
	err := c.cc.Invoke(ctx, PaymentGateway_GetBalance_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PaymentGatewayServer is the server API for PaymentGateway service.
// All implementations must embed UnimplementedPaymentGatewayServer
// for forward compatibility.
//
// ------------------ Payment Gateway Service ------------------
type PaymentGatewayServer interface {
	// Client registration (authentication/authorization)
	RegisterClient(context.Context, *RegisterRequest) (*RegisterResponse, error)
	// Process a payment using 2PC
	ProcessPayment(context.Context, *PaymentRequest) (*PaymentResponse, error)
	// Query client balance
	GetBalance(context.Context, *BalanceRequest) (*BalanceResponse, error)
	mustEmbedUnimplementedPaymentGatewayServer()
}

// UnimplementedPaymentGatewayServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPaymentGatewayServer struct{}

func (UnimplementedPaymentGatewayServer) RegisterClient(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterClient not implemented")
}
func (UnimplementedPaymentGatewayServer) ProcessPayment(context.Context, *PaymentRequest) (*PaymentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessPayment not implemented")
}
func (UnimplementedPaymentGatewayServer) GetBalance(context.Context, *BalanceRequest) (*BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (UnimplementedPaymentGatewayServer) mustEmbedUnimplementedPaymentGatewayServer() {}
func (UnimplementedPaymentGatewayServer) testEmbeddedByValue()                        {}

// UnsafePaymentGatewayServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PaymentGatewayServer will
// result in compilation errors.
type UnsafePaymentGatewayServer interface {
	mustEmbedUnimplementedPaymentGatewayServer()
}

func RegisterPaymentGatewayServer(s grpc.ServiceRegistrar, srv PaymentGatewayServer) {
	// If the following call pancis, it indicates UnimplementedPaymentGatewayServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PaymentGateway_ServiceDesc, srv)
}

func _PaymentGateway_RegisterClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentGatewayServer).RegisterClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PaymentGateway_RegisterClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentGatewayServer).RegisterClient(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentGateway_ProcessPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentGatewayServer).ProcessPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PaymentGateway_ProcessPayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentGatewayServer).ProcessPayment(ctx, req.(*PaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentGateway_GetBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentGatewayServer).GetBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PaymentGateway_GetBalance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentGatewayServer).GetBalance(ctx, req.(*BalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PaymentGateway_ServiceDesc is the grpc.ServiceDesc for PaymentGateway service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PaymentGateway_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentGateway",
	HandlerType: (*PaymentGatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterClient",
			Handler:    _PaymentGateway_RegisterClient_Handler,
		},
		{
			MethodName: "ProcessPayment",
			Handler:    _PaymentGateway_ProcessPayment_Handler,
		},
		{
			MethodName: "GetBalance",
			Handler:    _PaymentGateway_GetBalance_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}

const (
	BankService_Prepare_FullMethodName    = "/payment.BankService/Prepare"
	BankService_Commit_FullMethodName     = "/payment.BankService/Commit"
	BankService_Abort_FullMethodName      = "/payment.BankService/Abort"
	BankService_GetAccount_FullMethodName = "/payment.BankService/GetAccount"
)

// BankServiceClient is the client API for BankService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ------------------ Bank Service (used for 2PC) ------------------
type BankServiceClient interface {
	// 2PC Prepare phase
	Prepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error)
	// 2PC Commit phase
	Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error)
	// 2PC Abort phase
	Abort(ctx context.Context, in *AbortRequest, opts ...grpc.CallOption) (*AbortResponse, error)
	// Query account balance
	GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*GetAccountResponse, error)
}

type bankServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBankServiceClient(cc grpc.ClientConnInterface) BankServiceClient {
	return &bankServiceClient{cc}
}

func (c *bankServiceClient) Prepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PrepareResponse)
	err := c.cc.Invoke(ctx, BankService_Prepare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bankServiceClient) Commit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*CommitResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CommitResponse)
	err := c.cc.Invoke(ctx, BankService_Commit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bankServiceClient) Abort(ctx context.Context, in *AbortRequest, opts ...grpc.CallOption) (*AbortResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AbortResponse)
	err := c.cc.Invoke(ctx, BankService_Abort_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bankServiceClient) GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*GetAccountResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAccountResponse)
	err := c.cc.Invoke(ctx, BankService_GetAccount_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BankServiceServer is the server API for BankService service.
// All implementations must embed UnimplementedBankServiceServer
// for forward compatibility.
//
// ------------------ Bank Service (used for 2PC) ------------------
type BankServiceServer interface {
	// 2PC Prepare phase
	Prepare(context.Context, *PrepareRequest) (*PrepareResponse, error)
	// 2PC Commit phase
	Commit(context.Context, *CommitRequest) (*CommitResponse, error)
	// 2PC Abort phase
	Abort(context.Context, *AbortRequest) (*AbortResponse, error)
	// Query account balance
	GetAccount(context.Context, *GetAccountRequest) (*GetAccountResponse, error)
	mustEmbedUnimplementedBankServiceServer()
}

// UnimplementedBankServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBankServiceServer struct{}

func (UnimplementedBankServiceServer) Prepare(context.Context, *PrepareRequest) (*PrepareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Prepare not implemented")
}
func (UnimplementedBankServiceServer) Commit(context.Context, *CommitRequest) (*CommitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedBankServiceServer) Abort(context.Context, *AbortRequest) (*AbortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Abort not implemented")
}
func (UnimplementedBankServiceServer) GetAccount(context.Context, *GetAccountRequest) (*GetAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
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

func _BankService_Prepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).Prepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_Prepare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).Prepare(ctx, req.(*PrepareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BankService_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_Commit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).Commit(ctx, req.(*CommitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BankService_Abort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AbortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).Abort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_Abort_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).Abort(ctx, req.(*AbortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BankService_GetAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BankServiceServer).GetAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BankService_GetAccount_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BankServiceServer).GetAccount(ctx, req.(*GetAccountRequest))
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
			MethodName: "Prepare",
			Handler:    _BankService_Prepare_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _BankService_Commit_Handler,
		},
		{
			MethodName: "Abort",
			Handler:    _BankService_Abort_Handler,
		},
		{
			MethodName: "GetAccount",
			Handler:    _BankService_GetAccount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}
