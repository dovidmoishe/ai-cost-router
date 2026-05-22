package router

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"
)

const maxRouteBodyBytes = 64 << 10 // 64 KiB

var (
	defaultConfig   RouterConfig
	defaultConfigMu sync.RWMutex
)

// Init loads env-based defaults once at process start.
func Init() {
	defaultConfigMu.Lock()
	defaultConfig = ConfigFromEnv()
	defaultConfigMu.Unlock()
	loadInternalServiceToken()
}

func currentConfig() RouterConfig {
	defaultConfigMu.RLock()
	defer defaultConfigMu.RUnlock()
	return defaultConfig
}

// HealthHandler reports service readiness.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "ai-cost-router",
	})
}

// RouteHandler decides whether to call the LLM or return a canned reply.
func RouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
			return
		}
	} else {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}

	body := http.MaxBytesReader(w, r.Body, maxRouteBodyBytes)
	defer body.Close()

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	var req RouteRequest
	if err := dec.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body exceeds limit")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be a single JSON object")
		return
	}

	if err := validateRouteRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	cfg := mergeConfig(currentConfig(), req.Config)
	ctx := defaultContext(req.ConversationContext)
	decision := RouteForUserMessage(*req.UserText, cfg, ctx)

	writeJSON(w, http.StatusOK, decision)
}

func validateRouteRequest(req *RouteRequest) error {
	if req.UserText == nil {
		return errors.New("userText is required")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: code, Message: message})
}
