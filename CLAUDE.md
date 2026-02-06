# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a webhook container for the Mythic C2 framework (v3.0.0+). It listens for events via RabbitMQ and forwards formatted notifications to Slack, Discord, and Pushover channels. It is part of the MythicC2Profiles ecosystem.

## Build & Run Commands

All Go source lives in `C2_Profiles/basic_webhook/`. Run commands from that directory.

```bash
# Build (used inside Docker)
cd C2_Profiles/basic_webhook && make build

# Build and run locally with custom env vars
cd C2_Profiles/basic_webhook && make run_custom

# Build Go binary directly
cd C2_Profiles/basic_webhook && CGO_ENABLED=0 go build -o main .

# Start within Mythic
sudo ./mythic-cli start basic_webhook
```

There are no tests in this project.

## Architecture

**Entry point**: `C2_Profiles/basic_webhook/main.go` — calls `my_webhooks.Initialize()` then starts the Mythic container service in webhook mode via `MythicContainer.StartAndRunForever()`.

**Webhook handlers** (`C2_Profiles/basic_webhook/my_webhooks/`):
- `initialize.go` — Registers a `WebhookDefinition` with the Mythic framework, binding all event handler functions
- `new_callback.go` — Formats new agent callback events (IP, hostname, integrity level, etc.)
- `new_alert.go` — Forwards alerts with a 60-second throttle between messages
- `new_feedback.go` — Handles feedback events (bugs, feature requests, detection events); uses `mythicrpc.SendMythicRPCTaskSearch()` to fetch task context
- `new_startup.go` — Startup confirmation messages
- `new_custom.go` — Flexible key-value custom messages

**Message routing** (`sendingUtils.go`): Inspects the webhook URL to determine the destination — URLs containing "discord.com" route to `sendDiscordMessage()`, URLs containing "slack.com" send directly via HTTP POST to the Slack API, and URLs containing "pushover.net" route to `sendPushoverMessage()`.

**Discord support** (`discord.go`): Converts Slack Block Kit message format into Discord embed format, including hex color conversion.

**Pushover support** (`pushover.go`): Converts Slack Block Kit message format into a structured JSON payload with `title`, `message`, `event_type`, and `color` fields. Strips Slack markdown formatting for cleaner display. Users configure extraction templates on the Pushover side (`{{title}}`, `{{message}}`, etc.).

## Key Dependencies

- `github.com/MythicMeta/MythicContainer` (v1.6.2) — Core framework providing webhook structs, RPC client, logging, and the container runtime
- RabbitMQ — Event transport from Mythic server
- gRPC — Communication with Mythic server for task lookups and event logging

## Configuration

Webhook URL and channel routing are configured via environment variables (see Makefile for defaults):
- `WEBHOOK_DEFAULT_URL` — Slack, Discord, or Pushover webhook URL
- `WEBHOOK_DEFAULT_CHANNEL`, `WEBHOOK_DEFAULT_CALLBACK_CHANNEL`, `WEBHOOK_DEFAULT_FEEDBACK_CHANNEL`, `WEBHOOK_DEFAULT_STARTUP_CHANNEL` — Per-event-type channel overrides
- `RABBITMQ_HOST`, `RABBITMQ_PASSWORD`, `MYTHIC_SERVER_HOST`, `MYTHIC_SERVER_GRPC_PORT` — Infrastructure connectivity

The webhook URL can also be set per-operation in Mythic's UI.

## Docker & CI

The Dockerfile (`C2_Profiles/basic_webhook/Dockerfile`) uses a multi-stage build: `golang:1.25` for compilation, `alpine` for the runtime image. CI (`.github/workflows/docker.yml`) builds multi-platform images (amd64/arm64) and publishes to `ghcr.io/mythicc2profiles/basic_webhook`. The image tag in `config.json` is updated automatically on version tag pushes.

## Code Patterns

- All webhook handlers follow the same pattern: get default message template via `webhookstructs.GetNewDefaultWebhookMessage()`, resolve URL/channel from operation config, build Slack Block Kit blocks, then send via `sendWebhookMessage()`
- Error logging uses async goroutines: `go mythicrpc.SendMythicRPCOperationEventLogCreate(...)`
- The Go module path is `MyContainer` (declared in `go.mod`)
