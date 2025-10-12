package middleware

import (
	"github.com/raufhm/fairflow/shared/logger"
	"go.uber.org/zap"
	"net/http"
)

// CORS middleware adds CORS headers

func CORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger.Log.Info("CORS middleware", zap.String("method", r.Method), zap.String("path", r.URL.Path))

		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key")

		if r.Method == "OPTIONS" {

			w.WriteHeader(http.StatusOK)

			return

		}

		next.ServeHTTP(w, r)

	})

}
