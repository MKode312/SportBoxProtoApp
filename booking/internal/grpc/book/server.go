package bookgrpc

import (
	"booking/internal/clients/payments"
	"booking/internal/domain/boxes"
	"context"
	"errors"
	"fmt"
	"time"

	bookingv1 "github.com/MKode312/protos/gen/go/booking"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrNotEnoughFundsToPay = errors.New("not enough funds to pay")
	ErrNotFound            = errors.New("card not found")
)

const (
	emptyBalanceValue = -1
)

type Book interface {
	Book(ctx context.Context, email string, boxName string, timeStart string, timeHrs int64, timeMins int64) (reserveID int64, success bool, err error)
}

type serverAPI struct {
	bookingv1.UnimplementedBookServer
	book Book
}

type bookingServerAdapter struct {
	bookingv1.UnimplementedBookServer
    originalServer *serverAPI
    paymentsClient payments.Client
}


func Register(gRPC *grpc.Server, book Book, paymclient payments.Client) {
    realSrv := &serverAPI{book: book}
    wrapped := &bookingServerAdapter{

        originalServer: realSrv,
        paymentsClient: paymclient,
    }
    bookingv1.RegisterBookServer(gRPC, wrapped)
}

func (b *bookingServerAdapter) Book(ctx context.Context, req *bookingv1.BookRequest) (*bookingv1.BookResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	balance, paysuccess, err := compilePayment(ctx, req, b.paymentsClient)
	if err != nil {
		if err.Error() == ErrInvalidCredentials.Error() {
			return nil, status.Error(codes.InvalidArgument, "invalid email")
		}

		if err.Error() == ErrNotEnoughFundsToPay.Error() {
			return nil, status.Error(codes.OutOfRange, "not enough funds to pay for the booking")
		}

		if err.Error() == ErrNotFound.Error() {
			return nil, status.Error(codes.NotFound, "card not found")
		}

		st, _ := status.FromError(err)

		return nil, status.Error(st.Code(), "failed to pay for the booking")
	}

	if paysuccess {
		reserveID, success, err := b.originalServer.book.Book(ctx, req.GetEmail(), req.GetBoxName(), req.GetTimeStart(), req.GetTimeHrs(), req.GetTimeMins())
		if err != nil {
			return nil, status.Error(codes.Canceled, "booking cancelled")
		}

		return &bookingv1.BookResponse{
			ReserveId: reserveID,
			Balance:   balance,
			Success:   success,
		}, nil
	} else {
		return nil, status.Error(codes.Canceled, "payment failed")
	}
}

func compilePayment(ctx context.Context, req *bookingv1.BookRequest, paymentsClient payments.Client) (int64, bool, error) {
	const op = "book.CompilePayment"

	amount := (req.GetTimeHrs()*int64(boxes.HrsAmount) + req.GetTimeMins()*int64(boxes.MinAmount)) * req.GetPeopleAmount()

	balance, paysuccess, err := paymentsClient.Pay(ctx, req.GetEmail(), amount)
	if err != nil {
		if err.Error() == ErrNotEnoughFundsToPay.Error() {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrNotEnoughFundsToPay)
		}

		if err.Error() == ErrInvalidCredentials.Error() {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		if err.Error() == ErrNotFound.Error() {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	return balance, paysuccess, nil

}

func validate(req *bookingv1.BookRequest) error {
	if req.GetBoxName() == "" {
		return status.Error(codes.InvalidArgument, "boxName is required")
	}

	if _, ok := boxes.Boxes[req.GetBoxName()]; !ok {
		return status.Error(codes.NotFound, "boxName not found")
	}

	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPeopleAmount() <= 0 {
		return status.Error(codes.InvalidArgument, "invalid amount of people")
	}

	if req.GetPeopleAmount() > 4 {
		return status.Error(codes.InvalidArgument, "the amount of people is greater than 4")
	}

	timeStart, err := time.Parse(time.UnixDate, req.GetTimeStart())
	if err != nil {
		fmt.Println(err)
		return status.Error(codes.Internal, "internal error occured")
	}

	if timeStart.IsZero() {
		return status.Error(codes.InvalidArgument, "invalid time")
	}

	if req.GetTimeHrs() <= 0 && req.GetTimeMins() <= 0 {
		return status.Error(codes.InvalidArgument, "invalid time")
	}

	return nil
}
