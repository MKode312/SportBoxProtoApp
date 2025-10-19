package register

import (
	"context"	
	"log/slog"
	"net/http"
	ssogrpc "sport-box-api/internal/clients/sso/grpc"
	"sport-box-api/internal/lib/api/response"
	ssoerrors "sport-box-api/internal/lib/errors/sso"
	"sport-box-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	response.Response
}

func New(ctx context.Context, log *slog.Logger, ssoclient ssogrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.Error("invalid request"))
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		userID, err := ssoclient.Register(ctx, req.Email, req.Password) 
		if err != nil {
			if err.Error() == ssoerrors.ErrUserExists.Error() {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("This user already exists"))
				return
			}

			log.Error("failed to register user", sl.Err(err))

			
			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error("failed to register"))

			return
		}

		log.Info("user was successfully registered", slog.Int64("uid", userID))

		render.JSON(w, r, Response{
			Response: response.OK(),
		})
		
	}
}
