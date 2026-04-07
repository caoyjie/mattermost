# Rich Typing Status Plugin Technical Plan

**Status**: active
**Doc-Type**: technical-plan
**Scope**: mattermost, desktop, mobile, openclaw, streaming
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, openclaw, plugin, streaming, typing, webapp, desktop, mobile, android, ios, technical-plan, cicd, fallback, enhanced-mode

## 1. Implementation Architecture

1. Mattermost plugin server
- Implement `POST /status/start`, `POST /status/refresh`, `POST /status/stop`.
- Validate auth, channel membership, thread scope.
- Publish websocket status events with `schema_version`.

2. Webapp plugin UI
- Subscribe to status events.
- Render channel/thread scoped status row.
- Apply deterministic ordering and TTL cleanup.

3. OpenClaw Mattermost extension
- Emit stream lifecycle (`start|delta|complete|error|abort`).
- Call rich-status endpoints when capability is available.
- Fall back to typing + final message on failure/unavailable capability.

## 2. Data Contract

Required fields:
- `schema_version`, `channel_id`, `actor_id`, `text`, `expires_at`
- `parent_id` optional for thread scope

Error envelope:
- `code`, `message`, `request_id`

Compatibility:
- additive fields only in minor changes
- breaking field changes require major schema version bump

## 3. Platform Execution

1. Webapp (`mattermost`)
- full rich-status rendering and lifecycle behavior

2. Desktop (`desktop`)
- parity with webapp rendering and fallback behavior

3. Mobile (`mattermost-mobile`)
- baseline: fallback-only guaranteed
- enhanced mode: feature-gated rendering + TTL/reconnect cleanup

## 4. CI/CD and Gates

Required checks:
- `webapp-contract-and-ui`
- `desktop-stream-smoke`
- `mobile-android-stream`
- `mobile-ios-stream`

Promotion rule:
- enhanced mode only if all checks are green on the same target commit range
- otherwise remain in fallback mode

## 5. Rollout and Rollback

Rollout:
1. fallback-only baseline
2. internal enhanced mode
3. canary
4. broad rollout

Rollback:
- disable enhanced mode flag immediately
- keep fallback path active
- redeploy previous known-good extension/plugin artifact when needed
