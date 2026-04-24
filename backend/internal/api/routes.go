package api

import (
	"net/http"
	"net/url"
)

type routeHandler interface {
	handleJobs(http.ResponseWriter, *http.Request)
	handleJobByID(http.ResponseWriter, *http.Request)
	handleSubtitleAssets(http.ResponseWriter, *http.Request)
	handleSubtitleFile(http.ResponseWriter, *http.Request)
}

func Routes(handler routeHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", handler.handleJobs)
	mux.HandleFunc("/jobs/", handler.handleJobByID)
	mux.HandleFunc("/subtitle-assets", handler.handleSubtitleAssets)
	mux.HandleFunc("/subtitle-files/", handler.handleSubtitleFile)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedLocalOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedLocalOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme != "http" {
		return false
	}

	if parsed.Port() == "" {
		return false
	}

	hostname := parsed.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1"
}
