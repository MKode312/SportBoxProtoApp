package payments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	paymentsv1 "github.com/MKode312/protos/gen/go/payments"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	emptyBalanceValue = -1
)

type Client struct {
	api paymentsv1.PaymentClient
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
		api: paymentsv1.NewPaymentClient(cc),
		log: log,
	}, nil
}

func (c *Client) Pay(ctx context.Context, email string, amount int64) (balance int64, success bool, err error) {
	const op = "paymgrpc.Pay"

	resp, err := c.api.Pay(ctx, &paymentsv1.PayRequest{
		Email:  email,
		Amount: amount,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				return emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			}
			if st.Code() == codes.InvalidArgument {
				return emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			}
			if st.Code() == codes.Unavailable {
				return emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			}
			if st.Code() == codes.Canceled {
				return emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			}
		}
		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Balance, resp.Success, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

