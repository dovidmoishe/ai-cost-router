package router

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const responseTimeHeader = "X-Response-Time-Ms"

// ResponseTime records request duration for monitoring.
func ResponseTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		ms := time.Since(start).Milliseconds()
		ww.Header().Set(responseTimeHeader, strconv.FormatInt(ms, 10))

		log.Printf(
			"http_request method=%s path=%s status=%d duration_ms=%d request_id=%s",
			r.Method,
			r.URL.Path,
			ww.Status(),
			ms,
			middleware.GetReqID(r.Context()),
		)
	})
}
