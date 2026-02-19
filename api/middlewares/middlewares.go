package middlewares

import (
	"context"
	apiErr "eavesdropper/api/error"
	"eavesdropper/errs"
	"eavesdropper/services/auth"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const AuthTokenKey = "authToken"

func GetAuthToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("No authorization header found.")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("Invalid auth token format")
	}
	return strings.TrimPrefix(authHeader, prefix), nil
}

func ValidateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := GetAuthToken(r)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Invalid auth headerformat")
			return
		}

		ctx := context.WithValue(r.Context(), AuthTokenKey, token)

		err = auth.ValidateToken(ctx, token)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Invalid auth token")
			return
		}

		next(w, r.WithContext(ctx))
	}
}

func ValidateOwnership(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

		userId := r.PathValue("id")
		if userId == "" {
			apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Missing the id path param value")
			return
		}

		token, err := GetAuthToken(r)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Invalid auth headerformat")
			return
		}

		user, err := auth.GetUserRecord(ctx, token)
		if err != nil {
			apiErr.WriteJSONError(
				w,
				http.StatusInternalServerError,
				"",
				fmt.Sprintf("Faile to get the user auth record: %s", err),
			)
			return
		}
		if !user.EmailVerified {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, errs.ErrUnverifiedEmailAccount.Error(), "")
			return
		}

		if user.UID != userId {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, errs.ErrAuthTokenDoesNotMatchAcessedUser.Error(), "")
			return
		}

		next(w, r)
	}
}

// Checks if a user has persmission to read a transcript
func ValidateReadTranscript(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := GetAuthToken(r)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Invalid auth headerformat")
			return
		}

		ctx := context.WithValue(r.Context(), AuthTokenKey, token)

		err = auth.ValidateToken(ctx, token)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Invalid auth token")
			return
		}

		next(w, r.WithContext(ctx))
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow your dev and prod UI origins
		allowedOrigins := map[string]bool{
			"capacitor://localhost":              true, // iOS Capacitor
			"http://localhost":                   true, // Android Capacitor default
			"http://localhost:8100":              true,
			"https://eavesdropper-4f10b.web.app": true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// Allow whatever headers the browser requested, or a superset
			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders == "" {
				reqHeaders = "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control"
			}
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24h
		}

		// Preflight: answer and stop here
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
