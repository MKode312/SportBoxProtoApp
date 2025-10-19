package paymgrpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	paymentsv1 "github.com/MKode312/protos/gen/go/payments"
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

func (c *Client) AddCard(ctx context.Context, email string, cardNumber string, cvc string, phoneNumber string) (success bool, err error) {
	const op = "paymgrpc.AddCard"

	resp, err := c.api.AddCard(ctx, &paymentsv1.AddCardRequest{
		Email:       email,
		PhoneNumber: phoneNumber,
		CardNumber:  cardNumber,
		Cvc:         cvc,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.AlreadyExists {
				return false, fmt.Errorf("%s", st.Message())
			}
			if st.Code() == codes.Canceled {
				return false, fmt.Errorf("%s", st.Message())
			}
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.Success, nil
}

func (c *Client) AddFunds(ctx context.Context, email string, amount int64) (balance int64, success bool, err error) {
	const op = "paymgrpc.AddFunds"

	resp, err := c.api.AddFunds(ctx, &paymentsv1.AddFundsRequest{
		Email:  email,
		Amount: amount,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
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
