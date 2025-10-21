package book

import (
	"encoding/json"
	"net/http"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	"sport-box-api/internal/lib/api/response"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type CancelBookingRequest struct {
	Email string `json:"email"`
}

type CancelResponse struct {
	Success        bool  `json:"success"`
	RefundedAmount int64 `json:"refundedAmount"`
	Balance        int64 `json:"balance"`
}

// @Summary Cancel booking
// @Description Cancel a booking by ID
// @Tags booking
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param request body CancelBookingRequest true "Cancel booking request"
// @Success 200 {object} CancelResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /bookings/{id} [delete]
func Cancel(booker *bookgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.book.Cancel"

		bookingIDStr := chi.URLParam(r, "id")
		bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, response.Response{Error: "invalid booking ID"})
			return
		}

		var req CancelBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.JSON(w, http.StatusBadRequest, response.Response{Error: "invalid request"})
			return
		}

		if strings.TrimSpace(req.Email) == "" {
			response.JSON(w, http.StatusBadRequest, response.Response{Error: "email is required"})
			return
		}

		refundedAmount, balance, success, err := booker.CancelBooking(r.Context(), req.Email, bookingID)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "booking not found"):
				response.JSON(w, http.StatusNotFound, response.Response{Error: "booking not found"})
			case strings.Contains(err.Error(), "belongs to another user"):
				response.JSON(w, http.StatusForbidden, response.Response{Error: "this booking belongs to another user"})
			default:
				response.JSON(w, http.StatusInternalServerError, response.Response{Error: "failed to cancel booking"})
			}
			return
		}

		response.JSON(w, http.StatusOK, CancelResponse{
			Success:        success,
			RefundedAmount: refundedAmount,
			Balance:        balance,
		})
	}
}