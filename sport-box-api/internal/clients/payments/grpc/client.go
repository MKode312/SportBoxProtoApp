package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"sport-box-api/internal/lib/logger/sl"
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
	defaultEmptyBalance = -1
)

type Client struct {
	log        *slog.Logger
	client     paymentsv1.PaymentClient
	connection *grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.payments.grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	connection, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		log.Error("failed to connect to payments service", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := paymentsv1.NewPaymentClient(connection)

	return &Client{
		log:        log,
		client:     client,
		connection: connection,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Close() error {
	return c.connection.Close()
}

func (c *Client) AddCard(ctx context.Context, email string, cardNumber int64, cvc int64, phoneNumber int64) (bool, error) {
	const op = "clients.payments.grpc.AddCard"

	resp, err := c.client.AddCard(ctx, &paymentsv1.AddCardRequest{
		Email:       email,
		CardNumber:  fmt.Sprint(cardNumber),
		Cvc:         fmt.Sprint(cvc),
		PhoneNumber: fmt.Sprint(phoneNumber),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.AlreadyExists {
				return false, fmt.Errorf("%s: card already exists", op)
			}
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Success, nil
}

func (c *Client) AddFunds(ctx context.Context, email string, amount int64) (int64, bool, error) {
	const op = "clients.payments.grpc.AddFunds"

	resp, err := c.client.AddFunds(ctx, &paymentsv1.AddFundsRequest{
		Email:  email,
		Amount: amount,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				return defaultEmptyBalance, false, fmt.Errorf("%s: card not found", op)
			}
		}
		return defaultEmptyBalance, false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Balance, resp.Success, nil
}

func (c *Client) GetCard(ctx context.Context, email string) (string, string, error) {
	const op = "clients.payments.grpc.GetCard"

	resp, err := c.client.GetCard(ctx, &paymentsv1.GetCardRequest{
		Email: email,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return "", "", fmt.Errorf("%s: card not found", op)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаете номера как строки
	return resp.CardNumber, resp.PhoneNumber, nil
}
