package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// --- Authentication / Authorization Interceptor ---
// In this example, we assume that registered clients have been stored in a map.
// In production, this would likely be a secure database or file-backed store.
var validUsers = map[string]string{
	"alice":   "alicepassword",
	"bob":     "bobpassword",
	"gateway": "gatewaypassword", // Added so that bank servers accept calls from the gateway.
}


// AuthInterceptor verifies that the metadata contains valid credentials.
func AuthInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("missing metadata")
	}

	// Expect metadata "username" and "password"
	usernames := md.Get("username")
	passwords := md.Get("password")
	if len(usernames) == 0 || len(passwords) == 0 {
		return ctx, fmt.Errorf("missing credentials")
	}

	username := usernames[0]
	password := passwords[0]

	if expected, exists := validUsers[username]; !exists || expected != password {
		return ctx, fmt.Errorf("invalid credentials")
	}

	// Add client id (username) to context for later use
	newCtx := context.WithValue(ctx, "clientID", username)
	return newCtx, nil
}

// UnaryAuthInterceptor is a gRPC interceptor for authentication and authorization.
func UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Authenticate the request
	newCtx, err := AuthInterceptor(ctx)
	if err != nil {
		log.Printf("[AUTH ERROR] %v", err)
		return nil, err
	}
	// Optionally, add authorization checks here based on method name in info.FullMethod.
	// For example, only allow balance queries if authenticated.
	return handler(newCtx, req)
}

// --- Logging Interceptor ---
// Logs the details of each request, response and any errors.
func UnaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	startTime := time.Now()
	clientID, _ := ctx.Value("clientID").(string)

	// Log incoming request details
	log.Printf("[REQUEST] Method:%s | Client:%s | Payload:%+v", info.FullMethod, clientID, req)

	// Process request
	resp, err := handler(ctx, req)

	// Log response or error details
	duration := time.Since(startTime)
	if err != nil {
		log.Printf("[ERROR] Method:%s | Client:%s | Duration:%v | Error:%v", info.FullMethod, clientID, duration, err)
	} else {
		log.Printf("[RESPONSE] Method:%s | Client:%s | Duration:%v | Payload:%+v", info.FullMethod, clientID, duration, resp)
	}

	return resp, err
}

// ChainUnaryInterceptors chains multiple interceptors.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Wrap the handler with all interceptors
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			next := chain
			currentInterceptor := interceptors[i]
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, next)
			}
		}
		return chain(ctx, req)
	}
}

