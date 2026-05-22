# ai-cost-router

Small Go HTTP service that decides whether a tutor chat message should call an LLM or return a canned reply. Rule logic matches [`api/src/ai/ai-cost-router.ts`](../api/src/ai/ai-cost-router.ts) so the Nest API can delegate routing without duplicating heuristics in TypeScript.

## Requirements

- Go 1.22+

## Run locally

```bash
cd ai-cost-router
go run ./cmd/api
```

Optional `.env` (loaded via `godotenv`):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `INTERNAL_SERVICE_TOKEN` | *(required for routing)* | Shared secret; Nest sends `Authorization: Bearer <token>` on `POST /v1/route` |
| `AI_COST_ROUTER_ENABLED` | `true` | When `false`, always returns `call_model` |
| `AI_COST_ROUTER_FOLLOWUP` | `true` | Richer ack replies after substantive assistant turns |
| `AI_COST_ROUTER_MIN_CHARS` | `6` | Min normalized length before `too_short` bypass |
| `AI_COST_ROUTER_MIN_WORDS` | `2` | Min word count before `too_short` bypass |

Generate a strong internal token:

```bash
go run ./cmd/token
```

## API

### `GET /health`

```bash
curl -s http://localhost:8080/health
```

```json
{"status":"ok","service":"ai-cost-router"}
```

Every response includes `X-Response-Time-Ms` (server-side duration in milliseconds) and logs `duration_ms` for monitoring.

### `POST /v1/route`

Requires internal service auth (not required on `/health`):

```
Authorization: Bearer <INTERNAL_SERVICE_TOKEN>
```

`Content-Type: application/json`

**Request**

```json
{
  "userText": "hey",
  "conversationContext": {
    "hasRecentSubstantiveAssistant": false,
    "assistantAwaitingUserReply": false
  },
  "config": {
    "enabled": true,
    "minChars": 6,
    "minWords": 2,
    "followupBypass": true
  }
}
```

| Field | Required | Notes |
|-------|----------|-------|
| `userText` | yes | Raw user message; may be empty/whitespace (routes to `empty`) |
| `conversationContext` | no | Nest derives from chat history. `assistantAwaitingUserReply: true` when the last assistant turn expects yes/no (roadmap/quiz confirmations); short replies like `"okay"` then route to `call_model`. |
| `config` | no | Per-request overrides; server env is the base |

**Response — call model**

```json
{ "route": "call_model" }
```

**Response — bypass**

```json
{
  "route": "bypass_model",
  "reason": "greeting",
  "normalizedText": "hey",
  "replyText": "Hey! What are you studying right now?\n\n..."
}
```

`reason` is one of: `greeting`, `thanks`, `empty`, `too_short`, `too_vague`, `chitchat`, `acknowledgement`, `meta`.

**Errors**

```json
{ "error": "validation_error", "message": "userText is required" }
```

| Status | `error` |
|--------|---------|
| 400 | `invalid_json`, `validation_error` |
| 401 | `unauthorized` |
| 405 | `method_not_allowed` |
| 413 | `payload_too_large` |
| 415 | `unsupported_media_type` |

Example:

```bash
curl -s -X POST http://localhost:8080/v1/route \
  -H "Authorization: Bearer $INTERNAL_SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userText":"thanks"}'
```

```bash
curl -s -X POST http://localhost:8080/v1/route \
  -H "Authorization: Bearer $INTERNAL_SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userText":"quiz me on TypeScript basics"}'
```

## Tests

```bash
go test ./...
```

## Layout

```
cmd/api/main.go          # HTTP server
internal/router/
  types.go               # request/response types, env config
  rules.go               # ordered rule engine
  handler.go             # /health and /v1/route
  auth.go                # Bearer token middleware for /v1/route
  rules_test.go          # rule engine tests
  auth_test.go           # auth middleware tests
```

## Integrating with Nest

Today the API uses in-process `aiCostRouteForUserMessage`. To use this service, HTTP-call `POST /v1/route` with the `Authorization` header and the same `userText` and `conversationContext` you already derive in `ai-tutor-chat.service.ts`, then branch on `route` / `replyText` the same way as the TypeScript decision type.
