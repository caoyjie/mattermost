# End-to-End Stream Output Technical Plan (Mattermost + OpenClaw)

**Status**: active
**Doc-Type**: technical-plan
**Scope**: mattermost, desktop, mobile, openclaw, streaming
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a

**Tags**: mattermost, openclaw, streaming, plugin, webapp, desktop, mobile, android, ios, fallback, enhanced-mode, cicd, technical-plan


## 1. Objective

Deliver a production-safe stream output experience for Mattermost conversations powered by OpenClaw, including:
- visible activity status
- progressive partial output
- reliable final output
- robust fallback across web, desktop, Android, and iOS

## 2. Scope

In scope:
- OpenClaw Mattermost extension integration
- Mattermost plugin-based status signaling
- Webapp/desktop/mobile client behavior
- CI/CD, deployment, rollout, and rollback

Out of scope:
- Mattermost core source fork to alter built-in typing string behavior
- mandatory upstream changes to official OpenClaw main

## 3. Architecture Decision

1. Keep **official OpenClaw main** unchanged for core usage.
2. Implement custom behavior in a **custom Mattermost extension** artifact.
3. Keep Mattermost side as **plugin/service based** (no core fork).

## 4. Functional Contract

### 4.1 Stream lifecycle (normalized)
- `start`
- `delta`
- `complete`
- `error`
- `abort`

### 4.2 Status lifecycle (optional rich mode)
- `status_start`
- `status_refresh`
- `status_stop`

### 4.3 Fallback
- If rich status is unavailable, use:
  - typing indicator + final message
  - optional ephemeral progress for long tasks

## 5. Component Responsibilities

### 5.1 OpenClaw extension
- Emit normalized stream lifecycle events.
- Drive typing fallback and optional rich-status API calls.
- Preserve block-streaming and chunking behavior.
- Record correlation ID + timing metrics.

### 5.2 Mattermost plugin/web side
- Provide status endpoints/events (`start/refresh/stop`).
- Broadcast websocket events with auth/channel/thread scoping.
- Render status row in webapp/desktop web context.
- Apply TTL cleanup and deterministic ordering.

### 5.3 Clients
- Webapp/Desktop: rich status render + fallback.
- Mobile (Android/iOS):
  - baseline: fallback support mandatory
  - enhanced mode: feature-gated status rendering + TTL/reconnect cleanup

## 6. Data and Config

Required config flags:
- `streaming.mode`: `fallback` | `enhanced`
- `richStatus.enabled`: boolean
- `richStatus.refreshMs`
- `richStatus.ttlMs`

Behavior defaults:
- default mode: `fallback`
- enhanced mode enabled only after gate pass

## 7. CI/CD Strategy

Build policy:
- GitHub Actions only for extension build/promotion
- no manual local build dependency for deploy pipelines

Required job groups:
1. `webapp-contract-and-ui` (`mattermost`)
2. `desktop-stream-smoke` (`desktop`)
3. `mobile-android-stream` (`mattermost-mobile`)
4. `mobile-ios-stream` (`mattermost-mobile`)

Merge/promotion gate:
- all required groups must pass
- if any fail, remain in fallback mode

## 8. Deployment Model (Docker Compose Test Env)

1. Build custom extension artifact in Actions (`.tgz` + checksum).
2. Install artifact via `openclaw-cli` container:
   - `plugins install <artifact> --force`
3. Enable plugin if needed:
   - `plugins enable mattermost`
4. Restart gateway:
   - `docker compose restart openclaw-gateway`
5. Run smoke checks:
   - plugin list, channel interaction, lifecycle/fallback verification

## 9. Test Plan

### 9.1 Automated
- contract tests for lifecycle/status events
- UI lifecycle transition tests
- TTL expiry tests
- thread-scope tests
- reconnect/background cleanup tests (Android/iOS)

### 9.2 Manual gate
- one real Android + one real iOS validation pass before broad rollout

## 10. Observability

For every request:
- correlation/request ID
- time to first token
- completion duration
- chunk count
- error/abort reason

Operational counters:
- lifecycle success/failure counts
- stale-status cleanup counts
- fallback vs enhanced mode usage

## 11. Rollout

1. Stage A: fallback-only baseline
2. Stage B: internal enhanced mode (flag on)
3. Stage C: canary rollout
4. Stage D: broader rollout after stable telemetry

## 12. Rollback

Immediate:
- disable enhanced mode flag
- continue fallback mode

If extension regression:
- redeploy previous artifact version from Actions
- restart gateway

## 13. Risks and Mitigations

- Risk: event/status drift between OpenClaw and Mattermost plugin
  - Mitigation: versioned contract tests and strict CI gates
- Risk: stale indicators on reconnect/background transitions
  - Mitigation: TTL + cleanup logic + mobile lifecycle tests
- Risk: platform inconsistency
  - Mitigation: unified contract + per-platform required checks

## 14. Acceptance Criteria

- Users see immediate activity signal and progressive output.
- Final response persists as normal Mattermost message.
- Fallback works across all platforms.
- Enhanced mode passes web/desktop/android/ios CI gates.
- No core forks are required for baseline operation.
