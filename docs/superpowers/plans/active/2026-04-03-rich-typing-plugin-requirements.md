# Rich Typing Status Plugin Requirements

**Status**: active
**Doc-Type**: requirements
**Scope**: mattermost, desktop, mobile, openclaw, streaming
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a

**Tags**: mattermost, openclaw, streaming, plugin, webapp, desktop, mobile, android, ios, fallback, enhanced-mode, cicd, requirements


## 1. Background

Mattermost's built-in typing indicator (`<user> is typing...`) is fixed by client behavior and cannot be customized through the standard bot typing API.  
This document defines requirements for a plugin-based feature that adds customizable typing-like status messages (for example, `🤖 BotName is thinking...`) without modifying Mattermost core.

## 2. Goals

- Provide a customizable, temporary status display for bot activity in channels and threads.
- Keep compatibility with upstream Mattermost by avoiding core server/webapp forks.
- Ensure graceful fallback behavior on clients where custom rendering is not available, plus feature-gated enhanced rendering on supported mobile builds.
- Keep implementation operationally safe (TTL, rate limits, permission checks).

## 3. Non-Goals

- Changing built-in typing indicator text in Mattermost core clients.
- Adding unsupported client-core patches that increase merge conflict risk.
- Guaranteeing identical UI behavior across all client platforms without client-side extension support.

## 4. Scope

In scope:
- Server plugin API for start/refresh/stop status events.
- Webapp plugin UI rendering of rich typing status.
- Bot integration pattern for invoking the plugin API.
- Basic observability, security, and test requirements.

Out of scope:
- Upstream core forks outside plugin/public-extension boundaries.
- Enterprise-only behavior that requires private source patches.

## 5. Functional Requirements

### FR-1: Custom status lifecycle

- The system must support:
  - `start` status for a `(channel_id, actor_id[, parent_id])`.
  - `refresh` status before expiry.
  - `stop` status to remove active display immediately.
- A status must auto-expire after a configured TTL if not refreshed.

### FR-2: Payload schema

- A status payload must include:
  - `channel_id` (required)
  - `actor_id` (required; bot or user)
  - `text` (required; short status text)
  - `emoji` (optional)
  - `parent_id` (optional; thread scope)
  - `expires_at` (server-generated)

### FR-3: Server-side publishing

- Server plugin must publish websocket events for status state changes.
- Event payload must be scoped to authorized channel members only.
- If thread-scoped (`parent_id` set), event consumers must render only in relevant thread context.

### FR-4: Webapp rendering

- Webapp plugin must render a custom status row in channel/thread UI.
- Multiple statuses in one channel must be supported with deterministic ordering.
- Expired statuses must be removed automatically without page reload.

### FR-5: Bot integration API

- Plugin must expose authenticated HTTP endpoints:
  - `POST /status/start`
  - `POST /status/refresh`
  - `POST /status/stop`
- API must return explicit errors for invalid channel/thread/permission/token states.

### FR-6: Fallback behavior

- If custom UI plugin is unavailable on a client, the bot workflow must still be understandable via standard posts/ephemeral posts.
- The system must not block final bot responses if status signaling fails.

## 6. Non-Functional Requirements

### NFR-1: Upstream compatibility

- No modifications to core files under `server/channels/*` or `webapp/channels/*` are required for baseline feature operation.
- Implementation should be delivered as plugin packages and deployment configuration only.

### NFR-2: Performance

- Status events should be lightweight and suitable for short refresh cadence (e.g. every 2-5 seconds per active bot task).
- Server must enforce per-actor rate limits to prevent noisy updates.

### NFR-3: Reliability

- UI state must self-heal by TTL expiration if stop event is missed.
- Plugin restart must not leave permanently stale statuses visible.

### NFR-4: Security

- All endpoints must require bot/user authentication.
- Actor must be authorized to post status in target channel.
- Input validation must restrict maximum `text` length and accepted emoji format.

### NFR-5: Observability

- Server plugin must emit structured logs for start/refresh/stop errors.
- Include counters for active statuses and rejected API calls.

## 7. UX Requirements

- Status copy should be short and action-oriented (example: `thinking`, `querying`, `summarizing`).
- Emoji usage is optional and must not break layout if omitted.
- Status display should be visually distinct from actual posts to avoid confusion.
- Thread-scoped statuses should only appear in the relevant thread context.

## 8. Platform Requirements

- Webapp/desktop web experience: custom status rendered via webapp plugin.
- Mobile app:
  - Baseline: fallback messages where plugin UI is unavailable.
  - Enhanced mode: `mattermost-mobile` may render status from websocket events via a client feature flag.
  - iOS and Android implementations must both support TTL expiry, thread scoping, and deterministic ordering.

## 9. Data and State Model

- Recommended in-memory key:
  - `(channel_id, actor_id, parent_id_or_empty)` -> status record with `expires_at`.
- Optional persistence is not required for v1.
- Garbage collection:
  - Passive on read + periodic cleanup tick.

## 10. API Contract (Draft)

Request example for `start`/`refresh`:

```json
{
  "channel_id": "string",
  "actor_id": "string",
  "text": "thinking...",
  "emoji": "robot_face",
  "parent_id": "string_optional"
}
```

Response example:

```json
{
  "ok": true,
  "expires_at": 1760000000000
}
```

## 11. Acceptance Criteria

- AC-1: Bot can trigger start/refresh/stop and see custom status in webapp plugin UI.
- AC-2: Status auto-disappears after TTL without manual stop.
- AC-3: Unauthorized caller receives 4xx and no status is shown.
- AC-4: Thread-scoped status is visible only inside the target thread.
- AC-5: On mobile fallback mode, bot still provides progress through fallback posts.
- AC-5a: On mobile enhanced mode, status rendering works on both iOS and Android with TTL cleanup.
- AC-6: Pulling upstream Mattermost updates does not require resolving core source conflicts for this feature.

## 12. Testing Requirements

- Unit tests:
  - Payload validation
  - TTL logic
  - Rate limit behavior
  - Permission checks
- Integration tests:
  - API -> websocket event emission
  - Multi-status ordering and expiry behavior
- Manual tests:
  - Webapp custom UI rendering
  - Mobile fallback behavior on Android and iOS
  - Mobile enhanced mode behavior on Android and iOS (feature flag on)

## 13. Rollout Plan (High Level)

- Phase 1: Implement server plugin API and event publishing.
- Phase 2: Implement webapp plugin rendering + expiry handling.
- Phase 3: Integrate bot workflow (start/refresh/stop + fallback posts).
- Phase 4: Implement and validate `mattermost-mobile` enhanced mode on real devices (Android/iOS).

## 14. Risks and Mitigations

- Risk: Event spam from aggressive refresh loops.
  - Mitigation: server-side rate limiting + minimum refresh interval guidance.
- Risk: Inconsistent UX across clients.
  - Mitigation: fallback posting strategy and explicit platform behavior documentation.
- Risk: Stale UI if stop event is lost.
  - Mitigation: strict TTL-based expiry and periodic client cleanup.

## 15. Conflict-Avoidance Strategy

- Keep all feature implementation in plugin repositories/modules rather than Mattermost core modifications.
- Use additive files and feature flags; avoid invasive edits in high-churn core files.
- Maintain compatibility with public plugin APIs to reduce upstream merge risk.

## 16. CI/CD Multi-Platform Execution

Pipeline contract for this plan:

1. `mattermost` or plugin server/web pipeline:
- status API validation tests
- websocket emission tests
- permission/rate-limit tests

2. `desktop` pipeline:
- webapp plugin rendering smoke tests in desktop client context
- fallback behavior checks when plugin rendering is unavailable

3. `mattermost-mobile` pipeline:
- fallback-mode UI tests on Android and iOS
- enhanced-mode UI tests on Android and iOS (feature flag ON)
- TTL expiry and thread scoping tests

Promotion gate:
- Do not promote unless all three pipeline groups pass on the same commit range.

## 17. API Consistency and Versioning Policy

To prevent client/plugin drift, the rich-typing API/event contract must be versioned and backward-compatible.

### 17.1 Version field

- All status API payloads and websocket events must include:
  - `schema_version` (string, e.g. `v1`)

### 17.2 Backward compatibility rules

- Additive fields are allowed in minor updates.
- Existing required fields must not change meaning in-place.
- Removing/renaming fields requires a new major schema version.
- Unknown fields must be ignored by clients.

### 17.3 Required stable fields

- `channel_id`
- `actor_id`
- `text`
- `parent_id` (optional, for thread scope)
- `expires_at`
- `schema_version`

### 17.4 Error contract consistency

- API endpoints must return deterministic error envelopes:
  - `code` (stable machine-readable code)
  - `message` (human-readable)
  - `request_id` (traceability)
- Status code classes:
  - `4xx` for validation/authz/authn issues
  - `5xx` for server-side failures

## 18. Integration Contract Matrix

The same lifecycle semantics must hold across platform clients.

### 18.1 Required lifecycle assertions

For each client (webapp, desktop, Android, iOS):
- `status_start` shows indicator
- `status_refresh` extends TTL without duplicate stale rows
- `status_stop` removes indicator immediately
- missing `status_stop` still clears via TTL
- thread-scoped statuses do not leak to unrelated channels/threads

### 18.2 Fallback assertions

When rich-status capability is unavailable:
- typing + final reply path still works
- status API failures do not block final response

## 19. GitHub Actions Implementation Details

Define and require concrete workflows/check names.

### 19.1 Required workflows

1. `mattermost-rich-typing-contract.yml` (repo: `mattermost`)
- check name: `webapp-contract-and-ui`
- validates API/event schemas, UI lifecycle, TTL/thread behavior

2. `desktop-rich-typing-smoke.yml` (repo: `desktop`)
- check name: `desktop-stream-smoke`
- validates desktop rendering parity + fallback behavior

3. `mobile-rich-typing-android.yml` (repo: `mattermost-mobile`)
- check name: `mobile-android-stream`
- validates fallback + enhanced mode lifecycle on emulator

4. `mobile-rich-typing-ios.yml` (repo: `mattermost-mobile`)
- check name: `mobile-ios-stream`
- validates fallback + enhanced mode lifecycle on simulator

### 19.2 Trigger policy

- `pull_request`: run all required checks, block merge on failure
- `push` to main: run required checks and publish test artifacts
- `workflow_dispatch`: allow controlled re-runs for release candidates
- `schedule` (nightly): run extended regression matrix

### 19.3 Required artifacts

Each workflow uploads:
- test report (junit or equivalent)
- lifecycle trace log
- environment metadata (commit SHA, runner OS, app build version)

### 19.4 Release gating rule

- Enhanced mode can be promoted only when:
  - all required check names are green on the target commit
  - no critical contract regressions in lifecycle/fallback tests
  - canary telemetry remains stable in the defined observation window
