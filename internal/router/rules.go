package router

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	trimPunctRE  = regexp.MustCompile(`^[\s\p{P}\p{S}]+|[\s\p{P}\p{S}]+$`)
	tokenPunctRE = regexp.MustCompile(`^[\p{P}\p{S}]+|[\p{P}\p{S}]+$`)

	metaPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^what can you do\??$`),
		regexp.MustCompile(`(?i)^what do you do\??$`),
		regexp.MustCompile(`(?i)^who are you\??$`),
		regexp.MustCompile(`(?i)^how does this app work\??$`),
		regexp.MustCompile(`(?i)^how does this work\??$`),
		regexp.MustCompile(`(?i)^what is this (for|about)\??$`),
		regexp.MustCompile(`(?i)^what is edulearn\??$`),
		regexp.MustCompile(`(?i)^who made you\??$`),
	}

	continuationPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(go on|keep going|keep it going|tell me more|explain more|dig deeper|elaborate|give me another example)\b`),
		regexp.MustCompile(`(?i)\b(more detail|more details)\b`),
		regexp.MustCompile(`(?i)^(more|details|examples|another example|continue|continuation)\??$`),
	}

	vaguePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^help$`),
		regexp.MustCompile(`(?i)^explain$`),
		regexp.MustCompile(`(?i)^explain this$`),
		regexp.MustCompile(`(?i)^teach me$`),
		regexp.MustCompile(`(?i)^tell me$`),
		regexp.MustCompile(`(?i)^tell me about$`),
		regexp.MustCompile(`(?i)^what is this$`),
		regexp.MustCompile(`(?i)^i don'?t understand$`),
		regexp.MustCompile(`(?i)^i dont understand$`),
	}

	soloInterrogativeRE = regexp.MustCompile(`(?i)^(why|how|what|when|where|who)(\?)?$`)
)

var stopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "but": {}, "by": {},
	"can": {}, "could": {}, "do": {}, "for": {}, "from": {}, "get": {}, "give": {}, "help": {},
	"hi": {}, "how": {}, "i": {}, "i'm": {}, "im": {}, "in": {}, "is": {}, "it": {}, "like": {},
	"me": {}, "my": {}, "of": {}, "on": {}, "please": {}, "tell": {}, "teach": {}, "thanks": {},
	"thank": {}, "that": {}, "the": {}, "this": {}, "to": {}, "u": {}, "uh": {}, "um": {},
	"what": {}, "when": {}, "where": {}, "who": {}, "why": {}, "you": {}, "your": {},
}

var greetings = stringSet{
	"hi": true, "hey": true, "hello": true, "yo": true, "sup": true, "gm": true, "gn": true,
	"wagmi": true, "good morning": true, "good afternoon": true, "good evening": true,
	"good night": true, "good day": true, "morning": true, "afternoon": true, "evening": true,
	"hola": true, "howdy": true, "greetings": true, "hey there": true, "hello there": true, "hi there": true,
}

var thanks = stringSet{
	"thanks": true, "thank you": true, "thank you!": true, "thx": true, "ty": true, "ty!": true,
	"appreciate it": true, "much appreciated": true, "much obliged": true, "cheers": true, "you rock": true,
}

var chitchatExact = stringSet{
	"lol": true, "lmao": true, "lolol": true, "haha": true, "hahaha": true, "hehe": true, "hehehe": true,
	"rofl": true, "bru": true, "bruh": true, "oof": true, "meh": true, "nm": true, "nvm": true,
	"relatable": true, "i am bored": true, "i'm bored": true, "im bored": true, "idc": true, "wtf": true,
}

var ackExact = stringSet{
	"ok": true, "okay": true, "k": true, "kk": true, "okie": true, "okie dokie": true, "cool": true,
	"nice": true, "sweet": true, "great": true, "awesome": true, "perfect": true, "rad": true, "sick": true,
	"fire": true, "got it": true, "gotcha": true, "understood": true, "i see": true, "isee": true,
	"i got it": true, "makes sense": true, "that makes sense": true, "fair enough": true, "fair point": true,
	"right on": true, "noted": true, "roger": true, "copy that": true, "will do": true, "alright": true,
	"all right": true, "aight": true,
}

var awaitingAffirmationExact = stringSet{
	"ok": true, "okay": true, "k": true, "kk": true, "yes": true, "yeah": true, "yep": true,
	"yup": true, "sure": true, "y": true, "yea": true, "please": true, "yes please": true, "go ahead": true, "do it": true,
	"sounds good": true, "sound good": true, "that works": true, "let's do it": true, "lets do it": true,
	"go for it": true, "why not": true, "absolutely": true, "definitely": true, "of course": true,
	"for sure": true, "okie": true, "aight": true, "alright": true, "all right": true, "correct": true,
	"right": true, "perfect": true, "great": true, "awesome": true, "cool": true, "nice": true, "sweet": true,
	"got it": true, "understood": true, "i see": true, "isee": true, "makes sense": true, "that makes sense": true,
	"will do": true, "copy that": true, "roger": true,
}

var awaitingRejectionExact = stringSet{
	"no": true, "nope": true, "nah": true, "not now": true, "not yet": true, "wait": true, "stop": true,
	"cancel": true, "never mind": true, "nevermind": true, "dont": true, "don't": true, "do not": true,
	"hold on": true, "not really": true,
}

const loosePhraseMaxExtra = 8

var ackStems = []string{
	"that makes sense", "makes sense", "fair enough", "fair point", "all right", "okie dokie",
	"copy that", "got it", "i got it", "right on", "will do", "gotcha", "understood", "awesome",
	"perfect", "alright", "okay", "okie", "cool", "nice", "sweet", "great", "noted", "roger",
	"aight", "fire", "sick", "rad", "isee", "i see",
}

var awaitingAffirmationStems = []string{
	"that makes sense", "makes sense", "sounds good", "sound good", "let's do it", "lets do it",
	"go ahead", "that works", "go for it", "of course", "for sure", "why not", "do it", "copy that",
	"got it", "will do", "understood", "absolutely", "definitely", "awesome", "perfect", "alright",
	"please", "okay", "okie", "yeah", "sure", "cool", "nice", "sweet", "great", "right", "yes", "yep",
	"yup", "yea", "roger", "aight", "isee", "i see",
}

var awaitingRejectionStems = []string{
	"not really", "never mind", "nevermind", "not now", "not yet", "hold on", "do not", "don't",
	"dont", "cancel", "stop", "wait", "nope", "nah",
}

var shortOkRE = regexp.MustCompile(`^ok+$`)
var shortKRE = regexp.MustCompile(`^k{1,4}$`)
var shortNoRE = regexp.MustCompile(`^no+$`)

type stringSet map[string]bool

// Rule decides routing for one case. Returning ok=false means try the next rule.
type Rule func(in EvaluateInput) (RouteDecision, bool)

// defaultRules order matches api/src/ai/ai-cost-router.ts.
var defaultRules = []Rule{
	ruleEmpty,
	ruleContinuation,
	ruleNeedsContinuation,
	ruleAwaitingUserReply,
	ruleMeta,
	ruleGreeting,
	ruleThanks,
	ruleChitchat,
	ruleAcknowledgement,
	ruleTooVague,
	ruleTooShort,
}

// RouteForUserMessage applies the rule engine (same semantics as Nest aiCostRouteForUserMessage).
func RouteForUserMessage(userText string, cfg RouterConfig, ctx ConversationContext) RouteDecision {
	if !cfg.Enabled {
		return RouteDecision{Route: RouteCallModel}
	}

	normalized := NormalizeUserText(userText)
	words := splitWords(normalized)

	in := EvaluateInput{
		UserText:   userText,
		RawText:    userText,
		Normalized: normalized,
		Words:      words,
		Config:     cfg,
		Context:    ctx,
	}

	for _, rule := range defaultRules {
		if decision, ok := rule(in); ok {
			return decision
		}
	}

	return RouteDecision{Route: RouteCallModel}
}

// NormalizeUserText mirrors normalizeUserText in TypeScript.
func NormalizeUserText(input string) string {
	collapsed := strings.TrimSpace(strings.ToLower(input))
	collapsed = collapseSpaces(collapsed)
	return strings.TrimSpace(trimPunctRE.ReplaceAllString(collapsed, ""))
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return b.String()
}

func splitWords(normalized string) []string {
	if normalized == "" {
		return nil
	}
	return strings.Fields(normalized)
}

func bypass(reason RouteReason, normalized, reply string) RouteDecision {
	return RouteDecision{
		Route:          RouteBypassModel,
		Reason:         reason,
		NormalizedText: normalized,
		ReplyText:      reply,
	}
}

func ruleEmpty(in EvaluateInput) (RouteDecision, bool) {
	if in.Normalized == "" {
		return bypass(ReasonEmpty, in.Normalized, cannedReply(string(ReasonEmpty))), true
	}
	return RouteDecision{}, false
}

func ruleContinuation(in EvaluateInput) (RouteDecision, bool) {
	if looksLikeContinuationPrompt(in.Normalized) {
		return RouteDecision{Route: RouteCallModel}, true
	}
	return RouteDecision{}, false
}

func ruleNeedsContinuation(in EvaluateInput) (RouteDecision, bool) {
	if needsContinuationModel(in.Normalized, in.Context) {
		return RouteDecision{Route: RouteCallModel}, true
	}
	return RouteDecision{}, false
}

func ruleAwaitingUserReply(in EvaluateInput) (RouteDecision, bool) {
	if shouldCallModelForAwaitingReply(in.Normalized, in.Context) {
		return RouteDecision{Route: RouteCallModel}, true
	}
	return RouteDecision{}, false
}

func ruleMeta(in EvaluateInput) (RouteDecision, bool) {
	if isMetaQuery(in.Normalized) {
		return bypass(ReasonMeta, in.Normalized, cannedReply(string(ReasonMeta))), true
	}
	return RouteDecision{}, false
}

func ruleGreeting(in EvaluateInput) (RouteDecision, bool) {
	if isGreeting(in.Normalized) {
		return bypass(ReasonGreeting, in.Normalized, cannedReply(string(ReasonGreeting))), true
	}
	return RouteDecision{}, false
}

func ruleThanks(in EvaluateInput) (RouteDecision, bool) {
	if isThanks(in.Normalized) {
		return bypass(ReasonThanks, in.Normalized, cannedReply(string(ReasonThanks))), true
	}
	return RouteDecision{}, false
}

func ruleChitchat(in EvaluateInput) (RouteDecision, bool) {
	if isChitchat(in.Normalized) {
		return bypass(ReasonChitchat, in.Normalized, cannedReply(string(ReasonChitchat))), true
	}
	return RouteDecision{}, false
}

func ruleAcknowledgement(in EvaluateInput) (RouteDecision, bool) {
	if !isStandaloneAcknowledgement(in.Normalized) {
		return RouteDecision{}, false
	}
	reply := cannedReply(string(ReasonAcknowledgement))
	if in.Config.FollowupBypass && in.Context.HasRecentSubstantiveAssistant {
		reply = cannedReply("acknowledgement_followup")
	}
	return bypass(ReasonAcknowledgement, in.Normalized, reply), true
}

func ruleTooVague(in EvaluateInput) (RouteDecision, bool) {
	if in.Context.HasRecentSubstantiveAssistant {
		return RouteDecision{}, false
	}
	if isTooVague(in.Normalized) {
		return bypass(ReasonTooVague, in.Normalized, cannedReply(string(ReasonTooVague))), true
	}
	return RouteDecision{}, false
}

func ruleTooShort(in EvaluateInput) (RouteDecision, bool) {
	shortBypass := len(in.Normalized) < in.Config.MinChars || len(in.Words) < in.Config.MinWords
	if shortBypass {
		return bypass(ReasonTooShort, in.Normalized, cannedReply(string(ReasonTooShort))), true
	}
	return RouteDecision{}, false
}

func hasTopicWord(normalized string) bool {
	for _, t := range splitWords(normalized) {
		clean := tokenPunctRE.ReplaceAllString(t, "")
		if len(clean) < 4 {
			continue
		}
		if _, stop := stopwords[clean]; !stop {
			return true
		}
	}
	return false
}

func looksLikeContinuationPrompt(normalized string) bool {
	for _, re := range continuationPatterns {
		if re.MatchString(normalized) {
			return true
		}
	}
	return false
}

func isMetaQuery(normalized string) bool {
	for _, re := range metaPatterns {
		if re.MatchString(normalized) {
			return true
		}
	}
	return false
}

func isGreeting(normalized string) bool {
	if normalized == "" || len(normalized) > 48 {
		return false
	}
	return greetings[normalized]
}

func isThanks(normalized string) bool {
	if normalized == "" || len(normalized) > 72 {
		return false
	}
	return thanks[normalized]
}

func isChitchat(normalized string) bool {
	if normalized == "" || len(normalized) > 56 {
		return false
	}
	return chitchatExact[normalized]
}

func isStandaloneAcknowledgement(normalized string) bool {
	return matchesLoosePhrase(normalized, ackExact, ackStems, 72, matchesShortOk, matchesShortK)
}

func isTooVague(normalized string) bool {
	if normalized == "" {
		return false
	}
	for _, re := range vaguePatterns {
		if re.MatchString(normalized) && !hasTopicWord(normalized) {
			return true
		}
	}
	if soloInterrogativeRE.MatchString(normalized) && !hasTopicWord(normalized) {
		return true
	}
	return false
}

func needsContinuationModel(normalized string, ctx ConversationContext) bool {
	if !ctx.HasRecentSubstantiveAssistant {
		return false
	}
	if soloInterrogativeRE.MatchString(normalized) {
		return true
	}
	for _, re := range vaguePatterns {
		if re.MatchString(normalized) && !hasTopicWord(normalized) {
			return true
		}
	}
	return false
}

func shouldCallModelForAwaitingReply(normalized string, ctx ConversationContext) bool {
	if !ctx.AssistantAwaitingUserReply {
		return false
	}
	return isAwaitingAffirmation(normalized) || isAwaitingRejection(normalized)
}

func isAwaitingAffirmation(normalized string) bool {
	return matchesLoosePhrase(
		normalized,
		awaitingAffirmationExact,
		awaitingAffirmationStems,
		72,
		matchesShortOk,
		matchesShortK,
	)
}

func isAwaitingRejection(normalized string) bool {
	return matchesLoosePhrase(
		normalized,
		awaitingRejectionExact,
		awaitingRejectionStems,
		72,
		matchesShortNo,
	)
}

func matchesShortOk(normalized string) bool {
	return shortOkRE.MatchString(normalized)
}

func matchesShortK(normalized string) bool {
	return shortKRE.MatchString(normalized)
}

func matchesShortNo(normalized string) bool {
	return shortNoRE.MatchString(normalized)
}

func looseTokenMatch(token, stem string) bool {
	if token == stem {
		return true
	}
	if !strings.HasPrefix(token, stem) {
		return false
	}
	if len(token) > len(stem)+loosePhraseMaxExtra {
		return false
	}
	extra := token[len(stem):]
	if extra == "" {
		return true
	}
	last := stem[len(stem)-1]
	for i := 0; i < len(extra); i++ {
		if extra[i] != last {
			return false
		}
	}
	return true
}

func matchesLooseStem(normalized, stem string) bool {
	stemParts := strings.Fields(stem)
	msgParts := strings.Fields(normalized)
	maxLen := len(stem) + loosePhraseMaxExtra*len(stemParts)
	if len(normalized) > maxLen {
		return false
	}

	if len(stemParts) == 1 {
		if len(msgParts) != 1 {
			return false
		}
		return looseTokenMatch(msgParts[0], stemParts[0])
	}

	if len(msgParts) != len(stemParts) {
		return false
	}
	for i, part := range stemParts {
		if !looseTokenMatch(msgParts[i], part) {
			return false
		}
	}
	return true
}

func matchesLoosePhrase(
	normalized string,
	exact stringSet,
	stems []string,
	maxLen int,
	tinyChecks ...func(string) bool,
) bool {
	if normalized == "" || len(normalized) > maxLen {
		return false
	}
	if exact[normalized] {
		return true
	}
	for _, check := range tinyChecks {
		if check(normalized) {
			return true
		}
	}
	for _, stem := range stems {
		if matchesLooseStem(normalized, stem) {
			return true
		}
	}
	return false
}

func cannedReply(reason string) string {
	switch reason {
	case string(ReasonGreeting):
		return "Hey 👋 What skill are you building today? Tell me what you're working on and whether you want " +
			"an explanation, a quiz, or a practice plan — e.g. \"quiz me on TypeScript basics\" or \"7-day plan to get better at public speaking.\""
	case string(ReasonThanks):
		return "You're welcome 😊 Happy to keep going — what skill should we work on next?"
	case string(ReasonChitchat):
		return "Haha fair 😄 Whenever you're ready, drop a skill you're building and what you need — " +
			"explain it, quiz me, or map out a practice plan."
	case string(ReasonMeta):
		return "I'm your EduLearn tutor 📚 I help you build real-world skills — explanations, quizzes, practice plans, " +
			"flashcards, and study schedules. Tell me which skill you're working on and how you want to practice, and we'll go from there."
	case "acknowledgement_followup":
		return "Nice — want to go deeper on that skill? Ask something specific, or try " +
			"\"quiz me on React hooks\" or \"make me a week-long plan for SQL interviews.\""
	case string(ReasonAcknowledgement):
		return "Glad that helped 🙂 What skill should we level up next?"
	case string(ReasonEmpty):
		return "Hmm, I didn't catch any text — mind sending it again? " +
			"Name the skill and whether you want an explanation, quiz, or practice plan."
	case string(ReasonTooShort):
		return "Hey 👋 What skill are you building today?"
	case string(ReasonTooVague):
		return "I'd love to help — which skill are we working on, and do you want an explanation, quiz, or a practice plan? " +
			"Something like \"explain async/await for beginners\" or \"quiz me on UX research basics\" is perfect."
	default:
		return "Could you add a bit more detail? Something like \"quiz me on Python data structures, medium difficulty\" " +
			"helps me give you a way better answer."
	}
}
