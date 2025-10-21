package bookgrpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	bookingv1 "github.com/MKode312/protos/gen/go/booking"
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
	api bookingv1.BookClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "bookgrpc.New"

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
		api: bookingv1.NewBookClient(cc),
		log: log,
	}, nil
}

func (c *Client) Book(ctx context.Context, email string, boxName string, peopleAmount int64, timeStart string, timeHrs int64, timeMins int64) (balance int64, resID int64, success bool, err error) {
	const op = "bookgrpc.Book"

	resp, err := c.api.Book(ctx, &bookingv1.BookRequest{
		Email:        email,
		BoxName:      boxName,
		PeopleAmount: peopleAmount,
		TimeStart:    timeStart,
		TimeHrs:      timeHrs,
		TimeMins:     timeMins,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Canceled:
				return emptyBalanceValue,  0, false, fmt.Errorf("%s", st.Message())
			case codes.AlreadyExists:
				return emptyBalanceValue, 0, false, fmt.Errorf("%s", st.Message())
			case codes.NotFound:
				return emptyBalanceValue, 0, false, fmt.Errorf("%s", st.Message())
			case codes.InvalidArgument:
				return emptyBalanceValue,  0, false, fmt.Errorf("%s", st.Message())
			case codes.OutOfRange:
				return emptyBalanceValue, 0, false, fmt.Errorf("%s", st.Message())
			case codes.Internal:
				return emptyBalanceValue, 0, false, fmt.Errorf("%s", st.Message())
			}
		}

		return emptyBalanceValue, 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Balance, resp.ReserveId, resp.Success, nil
}

func (c *Client) CancelBooking(ctx context.Context, email string, bookingID int64) (refundedAmount int64, balance int64, success bool, err error) {
	const op = "bookgrpc.CancelBooking"

	resp, err := c.api.(ctx, &bookingv1.CancelBookingRequest{
		Email:     email,
		BookingId: bookingID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return 0, emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			case codes.PermissionDenied:
				return 0, emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			case codes.Internal:
				return 0, emptyBalanceValue, false, fmt.Errorf("%s", st.Message())
			}
		}

		return 0, emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.RefundedAmount, resp.Balance, resp.Success, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
