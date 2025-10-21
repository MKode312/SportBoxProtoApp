package book

import (
	"context"
	"log/slog"
	"net/http"
	bookgrpc "sport-box-api/internal/clients/booking/grpc"
	"sport-box-api/internal/lib/api/response"
)

type BoxesResponse struct {
	Boxes []Box `json:"boxes"`
}

type Box struct {
	Name         string  `json:"name"`
	PricePerHour float64 `json:"pricePerHour"`
	Available    bool    `json:"available"`
}

func GetBoxes(ctx context.Context, log *slog.Logger, client bookgrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Temporary mock data until backend service is ready
		boxes := []Box{
			{
				Name:         "LeninaBox",
				PricePerHour: 600.0,
				Available:    true,
			},
			{
				Name:         "SibirskayaBox",
				PricePerHour: 600.0,
				Available:    true,
			},
			{
				Name:         "LunacharskogoBox",
				PricePerHour: 600.0,
				Available:    false,
			},
		}

		response.JSON(w, http.StatusOK, BoxesResponse{Boxes: boxes})
	}
}
