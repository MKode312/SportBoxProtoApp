package addfunds

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
	Email  string `json:"email" validate:"required"`
	Amount int64  `json:"amount"`
}

type Response struct {
	Success bool  `json:"success"`
	Balance int64 `json:"balance"`
	response.Response
}

func New(ctx context.Context, log *slog.Logger, paymentsclient paymgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.paym.addFunds.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("Failed to decode request"))

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

		balance, success, err := paymentsclient.AddFunds(ctx, req.Email, req.Amount)
		if err != nil {
			if err.Error() == paymerrors.ErrNotFound.Error() {
				log.Error("invalid credentials")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("Not found"))
				return
			}

			log.Error("failed to add some funds", sl.Err(err))

			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error("Failed to add some funds to your card"))

			return
		}

		log.Info("successfully added funds")

		render.JSON(w, r, Response{
			Balance: balance,
			Success:  success,
			Response: response.OK(),
		})
	}
}
