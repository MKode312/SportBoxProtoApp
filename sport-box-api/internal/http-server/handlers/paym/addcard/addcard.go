package addcard

import (
	"context"
	"log/slog"
	"net/http"
	paymgrpc "sport-box-api/internal/clients/payments/grpc"
	"sport-box-api/internal/lib/api/response"
	paymerrors "sport-box-api/internal/lib/errors/payments"
	"sport-box-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email       string `json:"email" validate:"required"`
	PhoneNumber int64  `json:"phoneNumber" validate:"required"`
	CardNumber  int64  `json:"cardNumber" validate:"required"`
	Cvc         int64  `json:"cvc" validate:"required"`
}

type Response struct {
	Success bool `json:"success"`
	response.Response
}

func New(ctx context.Context, log *slog.Logger, paymentsclient paymgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.paym.addCard.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		success, err := paymentsclient.AddCard(ctx, req.Email, req.CardNumber, req.Cvc, req.PhoneNumber)
		if err != nil {
			if err.Error() == paymerrors.ErrInvalidCredentials.Error() {
				log.Error("invalid credentials")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid card data"))
				return
			}

			log.Error("failed to add a card", sl.Err(err))

			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error("Failed to add a card"))

			return
		}

		log.Info("successfully added card")

		render.JSON(w, r, Response{
			Success:  success,
			Response: response.OK(),
		})
	}
}
