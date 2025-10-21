package authMW

import (
	"net/http"
	"sport-box-api/internal/lib/api/response"
	jwtValidation "sport-box-api/internal/lib/jwt/validation"

	"github.com/go-chi/render"
)

func AuthorizeJWTToken(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			render.JSON(w, r, response.Error("Unauthorized"))
			return
		}

		tokenString := cookie.Value

		err = jwtValidation.VerifyJWTToken(tokenString)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			render.JSON(w, r, response.Error("Invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
