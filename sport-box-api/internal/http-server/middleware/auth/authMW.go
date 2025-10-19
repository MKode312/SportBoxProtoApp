package authMW

import (
	"log"
	"net/http"
	"sport-box-api/internal/lib/api/response"
	jwtValidation "sport-box-api/internal/lib/jwt/validation"

	"github.com/go-chi/render"
)

func AuthorizeJWTToken(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Unauthorized"))
			return
		}

		tokenString := cookie.Value

		err = jwtValidation.VerifyJWTToken(tokenString)
		if err != nil {
			render.Status(r, http.StatusForbidden)
			log.Printf("%v", err)
			render.JSON(w, r, response.Error("Invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
