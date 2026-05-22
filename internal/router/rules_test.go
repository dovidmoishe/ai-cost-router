package router

import (
	"strings"
	"testing"
)

func defaultTestConfig() RouterConfig {
	return RouterConfig{
		Enabled:        true,
		MinChars:       6,
		MinWords:       2,
		FollowupBypass: true,
	}
}

func TestNormalizeUserText(t *testing.T) {
	if got := NormalizeUserText("  Hello!!!  "); got != "hello" {
		t.Fatalf("normalize: got %q want hello", got)
	}
	if got := NormalizeUserText("  "); got != "" {
		t.Fatalf("normalize empty: got %q", got)
	}
}

func TestRouteGreeting(t *testing.T) {
	d := RouteForUserMessage("hey", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonGreeting {
		t.Fatalf("greeting: %+v", d)
	}
	if !strings.Contains(strings.ToLower(d.ReplyText), "hey") {
		t.Fatalf("greeting reply missing hey: %q", d.ReplyText)
	}
}

func TestRouteThanks(t *testing.T) {
	d := RouteForUserMessage("thanks", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonThanks {
		t.Fatalf("thanks: %+v", d)
	}
	if !strings.Contains(strings.ToLower(d.ReplyText), "you're welcome") {
		t.Fatalf("thanks reply: %q", d.ReplyText)
	}
}

func TestRouteEmpty(t *testing.T) {
	d := RouteForUserMessage("   ", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonEmpty {
		t.Fatalf("empty: %+v", d)
	}
}

func TestRouteTooShort(t *testing.T) {
	d := RouteForUserMessage("ab", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonTooShort {
		t.Fatalf("too_short: %+v", d)
	}
}

func TestRouteAcknowledgementColdStart(t *testing.T) {
	d := RouteForUserMessage("ok", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonAcknowledgement {
		t.Fatalf("ack: %+v", d)
	}
}

func TestRouteTooVague(t *testing.T) {
	d := RouteForUserMessage("help", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonTooVague {
		t.Fatalf("too_vague: %+v", d)
	}
}

func TestRouteSoloInterrogativeColdStart(t *testing.T) {
	d := RouteForUserMessage("why?", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonTooVague {
		t.Fatalf("solo why cold: %+v", d)
	}
}

func TestRouteRealQuestion(t *testing.T) {
	d := RouteForUserMessage(
		"Explain Solana PDAs like I am new to Anchor.",
		defaultTestConfig(),
		ConversationContext{},
	)
	if d.Route != RouteCallModel {
		t.Fatalf("real question: %+v", d)
	}
}

func TestRouteContinuation(t *testing.T) {
	d := RouteForUserMessage(
		"tell me more about that step",
		defaultTestConfig(),
		ConversationContext{
			HasRecentSubstantiveAssistant: true,
			AssistantAwaitingUserReply:    false,
		},
	)
	if d.Route != RouteCallModel {
		t.Fatalf("continuation: %+v", d)
	}
}

func TestRouteSoloWhyAfterSubstantive(t *testing.T) {
	d := RouteForUserMessage(
		"why",
		defaultTestConfig(),
		ConversationContext{HasRecentSubstantiveAssistant: true},
	)
	if d.Route != RouteCallModel {
		t.Fatalf("why after substantive: %+v", d)
	}
}

func TestRouteHelpAfterSubstantive(t *testing.T) {
	d := RouteForUserMessage(
		"help",
		defaultTestConfig(),
		ConversationContext{HasRecentSubstantiveAssistant: true},
	)
	if d.Route != RouteCallModel {
		t.Fatalf("help after substantive: %+v", d)
	}
}

func TestRouteCoolAfterSubstantive(t *testing.T) {
	d := RouteForUserMessage(
		"cool",
		defaultTestConfig(),
		ConversationContext{
			HasRecentSubstantiveAssistant: true,
			AssistantAwaitingUserReply:    false,
		},
	)
	if d.Route != RouteBypassModel || d.Reason != ReasonAcknowledgement {
		t.Fatalf("cool follow-up: %+v", d)
	}
}

func TestRouteAwaitingAffirmation(t *testing.T) {
	ctx := ConversationContext{
		HasRecentSubstantiveAssistant: true,
		AssistantAwaitingUserReply:    true,
	}
	for _, text := range []string{"okay", "yes", "sure", "go ahead", "sounds good", "yes please"} {
		d := RouteForUserMessage(text, defaultTestConfig(), ctx)
		if d.Route != RouteCallModel {
			t.Fatalf("%q awaiting affirmation: %+v", text, d)
		}
	}
}

func TestRouteAwaitingRejection(t *testing.T) {
	d := RouteForUserMessage(
		"not now",
		defaultTestConfig(),
		ConversationContext{
			HasRecentSubstantiveAssistant: true,
			AssistantAwaitingUserReply:    true,
		},
	)
	if d.Route != RouteCallModel {
		t.Fatalf("awaiting rejection: %+v", d)
	}
}

func TestRouteOkayColdStart(t *testing.T) {
	d := RouteForUserMessage("ok", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonAcknowledgement {
		t.Fatalf("cold ok: %+v", d)
	}
}

func TestIsAwaitingAffirmation(t *testing.T) {
	for _, text := range []string{"yes please", "okayy", "okiee", "yesss", "yeahh"} {
		if !isAwaitingAffirmation(text) {
			t.Fatalf("expected %q to affirm", text)
		}
	}
	if isAwaitingAffirmation("maybe later") {
		t.Fatal("expected maybe later to not affirm")
	}
	if isAwaitingAffirmation("book") {
		t.Fatal("expected book to not affirm")
	}
}

func TestLooseAcknowledgement(t *testing.T) {
	for _, text := range []string{"okayy", "okiee", "coolll"} {
		if !isStandaloneAcknowledgement(text) {
			t.Fatalf("expected %q to ack", text)
		}
	}
	if isStandaloneAcknowledgement("i seek help") {
		t.Fatal("expected i seek help to not ack")
	}
}

func TestRouteMeta(t *testing.T) {
	d := RouteForUserMessage("what can you do?", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonMeta {
		t.Fatalf("meta: %+v", d)
	}
}

func TestRouteChitchat(t *testing.T) {
	d := RouteForUserMessage("lol", defaultTestConfig(), ConversationContext{})
	if d.Route != RouteBypassModel || d.Reason != ReasonChitchat {
		t.Fatalf("chitchat: %+v", d)
	}
}

func TestRouteDisabled(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Enabled = false
	d := RouteForUserMessage("hey", cfg, ConversationContext{})
	if d.Route != RouteCallModel {
		t.Fatalf("disabled: %+v", d)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("AI_COST_ROUTER_ENABLED", "false")
	t.Setenv("AI_COST_ROUTER_MIN_CHARS", "10")
	cfg := ConfigFromEnv()
	if cfg.Enabled {
		t.Fatal("expected enabled false")
	}
	if cfg.MinChars != 10 {
		t.Fatalf("minChars: %d", cfg.MinChars)
	}
}
