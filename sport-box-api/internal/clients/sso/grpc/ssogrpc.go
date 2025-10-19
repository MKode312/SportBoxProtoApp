package ssogrpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ssov1 "github.com/MKode312/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api ssov1.AuthClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "ssogrpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
	    grpc.WithTransportCredentials(insecure.NewCredentials()),
	    grpc.WithChainUnaryInterceptor(
		grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...), 
		grpcretry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: ssov1.NewAuthClient(cc),
		log: log,
	}, nil
}

func (c *Client) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "ssogrpc.Register"

	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
		if st.Code() == codes.AlreadyExists {
			return 0, fmt.Errorf("%s", st.Message())
		} 
		if st.Code() == codes.InvalidArgument {
			return 0, fmt.Errorf("%s", st.Message())
		}
	}
	return 0, fmt.Errorf("%s: %w", op, err)
	}

	return resp.UserId, nil
}

func (c *Client) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "ssogrpc.Login"

	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
		Email: email,
		Password: password,
		AppId: int32(appID),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
		if st.Code() == codes.NotFound {
			return "", fmt.Errorf("%s", st.Message())
		}
	}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.Token, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
