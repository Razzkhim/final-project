package middleware

import (
	"final-project/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var token string
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			token = cookie.Value

			jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				return config.JWTKey, nil
			})

			if !jwtToken.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
