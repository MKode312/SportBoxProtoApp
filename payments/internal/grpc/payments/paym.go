package paymgrpc

import (
	"context"
	"errors"
	"payments/internal/services/payment"


	paymentsv1 "github.com/MKode312/protos/gen/go/payments"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Payment interface {
	AddCard(ctx context.Context, email string, cardNumber string, cvc string, phoneNumber string) (success bool, err error)
	AddFunds(ctx context.Context, email string, amount int64) (balance int64, success bool, err error)
	Pay(ctx context.Context, email string, amount int64) (balance int64, success bool, err error)
	GetCard(ctx context.Context, email string) (cardNumber string, phoneNumber string, err error)
}

type serverAPI struct {
	paymentsv1.UnimplementedPaymentServer
	payment Payment
}

func Register(gRPC *grpc.Server, payment Payment) {
	paymentsv1.RegisterPaymentServer(gRPC, &serverAPI{payment: payment})
}

func (s *serverAPI) AddCard(ctx context.Context, req *paymentsv1.AddCardRequest) (*paymentsv1.AddCardResponse, error) {
	if req.GetCardNumber() == "" || req.GetCvc() == "" || req.GetEmail() == "" || req.GetPhoneNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "card details, phone number and email are required")
	}

	success, err := s.payment.AddCard(ctx, req.GetEmail(), req.GetCardNumber(), req.GetCvc(), req.GetPhoneNumber())
	if err != nil {
		if errors.Is(err, payment.ErrInvalidCredentials) {
			return nil, status.Error(codes.AlreadyExists, "card already exists")
		}
		return nil, status.Error(codes.Canceled, "unsuccessfully added a card")
	}

	return &paymentsv1.AddCardResponse{
		Success: success,
	}, nil
}

func (s *serverAPI) Pay(ctx context.Context, req *paymentsv1.PayRequest) (*paymentsv1.PayResponse, error) {
	if err := validateEmail(req.GetEmail()); err != nil {
		return nil, err
	} else if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	balance, success, err := s.payment.Pay(ctx, req.GetEmail(), req.GetAmount())
	if err != nil {
		if errors.Is(err, payment.ErrNotEnoughFundsToPay) {
			return nil, status.Error(codes.OutOfRange, "not enough funds to pay")
		}
		if errors.Is(err, payment.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Canceled, "unsuccessful payment, operation cancelled")
	}

	return &paymentsv1.PayResponse{
		Balance: balance,
		Success: success,
	}, nil
}

func (s *serverAPI) AddFunds(ctx context.Context, req *paymentsv1.AddFundsRequest) (*paymentsv1.AddFundsResponse, error) {
	if err := validateEmail(req.GetEmail()); err != nil {
		return nil, err
	} else if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	balance, success, err := s.payment.AddFunds(ctx, req.GetEmail(), req.GetAmount())
	if err != nil {
		if errors.Is(err, payment.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Canceled, "unsuccessful adding funds")
	}

	return &paymentsv1.AddFundsResponse{
		Balance: balance,
		Success: success,
	}, nil

}

func (s *serverAPI) GetCard(ctx context.Context, req *paymentsv1.GetCardRequest) (*paymentsv1.GetCardResponse, error) {
	if err := validateEmail(req.GetEmail()); err != nil {
		return nil, err
	}

	cardNumber, phoneNumber, err := s.payment.GetCard(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, payment.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Internal, "failed to get card information")
	}

	return &paymentsv1.GetCardResponse{
		CardNumber: cardNumber,
		PhoneNumber: phoneNumber,
		Success: true,
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	return nil
}
