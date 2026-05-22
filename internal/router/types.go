package router

import (
	"os"
	"strconv"
	"strings"
)

const (
	defaultMinChars = 6
	defaultMinWords = 2
)

// RouteReason is set when route is bypass_model.
type RouteReason string

const (
	ReasonGreeting        RouteReason = "greeting"
	ReasonThanks          RouteReason = "thanks"
	ReasonEmpty           RouteReason = "empty"
	ReasonTooShort        RouteReason = "too_short"
	ReasonTooVague        RouteReason = "too_vague"
	ReasonChitchat        RouteReason = "chitchat"
	ReasonAcknowledgement RouteReason = "acknowledgement"
	ReasonMeta            RouteReason = "meta"
)

// RouteName is the routing outcome.
type RouteName string

const (
	RouteCallModel   RouteName = "call_model"
	RouteBypassModel RouteName = "bypass_model"
)

// ConversationContext mirrors the Nest ai-cost-router context.
type ConversationContext struct {
	HasRecentSubstantiveAssistant bool `json:"hasRecentSubstantiveAssistant"`
}

// RouterConfig controls bypass thresholds and feature flags.
type RouterConfig struct {
	Enabled        bool `json:"enabled"`
	MinChars       int  `json:"minChars"`
	MinWords       int  `json:"minWords"`
	FollowupBypass bool `json:"followupBypass"`
}

// RouterConfigOverride allows per-request tuning without sending a full config block.
type RouterConfigOverride struct {
	Enabled        *bool `json:"enabled,omitempty"`
	MinChars       *int  `json:"minChars,omitempty"`
	MinWords       *int  `json:"minWords,omitempty"`
	FollowupBypass *bool `json:"followupBypass,omitempty"`
}

// RouteRequest is the POST /v1/route body.
type RouteRequest struct {
	UserText            *string               `json:"userText"`
	ConversationContext *ConversationContext  `json:"conversationContext,omitempty"`
	Config              *RouterConfigOverride `json:"config,omitempty"`
}

// RouteDecision is returned from POST /v1/route.
type RouteDecision struct {
	Route          RouteName   `json:"route"`
	Reason         RouteReason `json:"reason,omitempty"`
	NormalizedText string      `json:"normalizedText,omitempty"`
	ReplyText      string      `json:"replyText,omitempty"`
}

// HealthResponse is returned from GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// ErrorResponse is returned for client and server errors.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// EvaluateInput is passed through the rule engine.
type EvaluateInput struct {
	UserText   string
	RawText    string
	Normalized string
	Words      []string
	Config     RouterConfig
	Context    ConversationContext
}

// ConfigFromEnv loads defaults from environment variables (same names as Nest).
func ConfigFromEnv() RouterConfig {
	return RouterConfig{
		Enabled:        envBool("AI_COST_ROUTER_ENABLED", true),
		FollowupBypass: envBool("AI_COST_ROUTER_FOLLOWUP", true),
		MinChars:       envInt("AI_COST_ROUTER_MIN_CHARS", defaultMinChars),
		MinWords:       envInt("AI_COST_ROUTER_MIN_WORDS", defaultMinWords),
	}
}

func envBool(key string, defaultVal bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	lower := strings.ToLower(raw)
	if lower == "false" || lower == "0" {
		return false
	}
	return true
}

func envInt(key string, defaultVal int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func mergeConfig(base RouterConfig, override *RouterConfigOverride) RouterConfig {
	if override == nil {
		return base
	}
	out := base
	if override.Enabled != nil {
		out.Enabled = *override.Enabled
	}
	if override.FollowupBypass != nil {
		out.FollowupBypass = *override.FollowupBypass
	}
	if override.MinChars != nil && *override.MinChars > 0 {
		out.MinChars = *override.MinChars
	}
	if override.MinWords != nil && *override.MinWords > 0 {
		out.MinWords = *override.MinWords
	}
	return out
}

func defaultContext(ctx *ConversationContext) ConversationContext {
	if ctx == nil {
		return ConversationContext{}
	}
	return *ctx
}
