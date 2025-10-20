package book

import (
	"context"
	"log/slog"
	"net/http"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	"sport-box-api/internal/lib/api/response"
	bookerrors "sport-box-api/internal/lib/errors/booking"
	"sport-box-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email        string `json:"email" validate:"required"`
	BoxName      string `json:"boxName" validate:"required"`
	PeopleAmount int64  `json:"peopleAmount" validate:"required"`
	TimeStart    string `json:"timeStart" validate:"required"`
	TimeHrs      int64  `json:"timeHrs" validate:"required"`
	TimeMins     int64  `json:"timeMins" validate:"required"`
}

type Response struct {
	Success bool  `json:"success"`
	Balance int64 `json:"balance"`
	ResID   int64 `json:"reserveID"`
	response.Response
}

func New(ctx context.Context, log *slog.Logger, bookingclient bookgrpc.Client) http.HandlerFunc {
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

		balance, resID, success, err := bookingclient.Book(ctx, req.Email, req.BoxName, req.PeopleAmount, req.TimeStart, req.TimeHrs, req.TimeMins)
		if err != nil {
			if err.Error() == bookerrors.ErrInvalidCredentials.Error() {
				log.Error("invalid credentials")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid credentials"))
				return
			}

			if err.Error() == bookerrors.ErrAlreadyBooked.Error() {
				log.Error("already booked")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("This sport box is already booked for this time, try to book it later"))
				return
			}

			if err.Error() == bookerrors.ErrBookingNotFound.Error() {
				log.Error("booking not found")
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.Error("Something went wrong"))
				return
			}

			if err.Error() == bookerrors.ErrCardNotFound.Error() {
				log.Error("card not found")
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, response.Error("We haven't found your credit card in the system. Make sure you have added it as a paying method"))
				return
			}

			if err.Error() == bookerrors.ErrNotEnoughFundsToPay.Error() {
				log.Error("not enough funds to pay")
				render.Status(r, http.StatusPaymentRequired)
				render.JSON(w, r, response.Error("You don't have enough funds at your account's wallet"))
				return
			}

			log.Error("failed to book a box and pay for it", sl.Err(err))

			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error("Failed to book a box and pay for it"))

			return
		}

		log.Info("successfully booked a box and paid for it", slog.Int64("reservationID", resID))

		render.JSON(w, r, Response{
			Balance:  balance,
			Success:  success,
			ResID:    resID,
			Response: response.OK(),
		})
	}
}
