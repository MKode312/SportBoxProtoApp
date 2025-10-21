package book

import (
	"context"
	"log/slog"
	"net/http"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	"sport-box-api/internal/lib/api/response"
)

type BookingsResponse struct {
	Bookings []Booking `json:"bookings"`
}

type Booking struct {
	ID           int64  `json:"id"`
	BoxName      string `json:"boxName"`
	TimeStart    string `json:"timeStart"`
	TimeHrs      int64  `json:"timeHrs"`
	TimeMins     int64  `json:"timeMins"`
	PeopleAmount int64  `json:"peopleAmount"`
}

func GetBookings(ctx context.Context, log *slog.Logger, client bookgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Temporary mock data until backend service is ready
		bookings := []Booking{
			{
				ID:           1,
				BoxName:      "Box 1",
				TimeStart:    "14:00",
				TimeHrs:      2,
				TimeMins:     0,
				PeopleAmount: 4,
			},
		}

		response.JSON(w, http.StatusOK, BookingsResponse{Bookings: bookings})
	}
}