package getcard

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
)

type Request struct {
	Email string `json:"email" validate:"required,email"`
}

type Response struct {
	CardNumber  int64  `json:"cardNumber"`
	PhoneNumber int64  `json:"phoneNumber"`
	Status      string `json:"status"`
	response.Response
}

func New(ctx context.Context, log *slog.Logger, paymentsclient paymgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.paym.getcard.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		email := r.URL.Query().Get("email")
		if email == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("Email is required"))
			return
		}

		cardNumber, phoneNumber, err := paymentsclient.GetCard(ctx, email)
		if err != nil {
			if err.Error() == paymerrors.ErrInvalidCredentials.Error() {
				log.Error("invalid credentials", sl.Err(err))
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("Invalid credentials"))
				return
			}

			if err.Error() == paymerrors.ErrNotFound.Error() {
				log.Error("card not found", sl.Err(err))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.Error("Card not found"))
				return
			}

			log.Error("failed to get card", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Failed to get card"))
			return
		}

		log.Info("successfully retrieved card", slog.Int64("cardNumber", cardNumber))

		render.JSON(w, r, Response{
			CardNumber:  cardNumber,
			PhoneNumber: phoneNumber,
			Status:      "Card retrieved successfully",
			Response:    response.OK(),
		})
	}
}