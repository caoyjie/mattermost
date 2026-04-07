# Mattermost Bot Stream Output Requirements and Cherry Studio Implementation Record

**Status**: active
**Doc-Type**: requirements
**Scope**: mattermost, desktop, mobile, openclaw, streaming
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a

**Tags**: mattermost, openclaw, streaming, plugin, webapp, desktop, mobile, android, ios, fallback, enhanced-mode, cicd, requirements


## 1. Purpose

Define a practical, low-conflict plan for implementing stream output UX in Mattermost bot workflows, while documenting a working reference architecture from Cherry Studio.

This document has two goals:

- Specify requirements for Mattermost-side implementation (bot + optional plugin UI + multi-platform behavior).
- Record how Cherry Studio implements streaming output so the same patterns can be reused.

## 2. Background

Mattermost built-in typing indicator (`<name> is typing...`) is not customizable through bot typing APIs.  
To provide better real-time UX (typing + partial text + final response), we need a stream-oriented design.

To avoid future upstream merge conflicts, this should be implemented without modifying Mattermost core source whenever possible.

## 3. Scope and Constraints

In scope:

- Bot-side streaming response flow.
- Optional plugin-based custom status UI (`Bot is thinking...`) for web/desktop webapp.
- Multi-platform client behavior for web, desktop, and `mattermost-mobile` (iOS/Android), including fallback and enhanced mobile mode.

Out of scope:

- Forking Mattermost core to change built-in typing text.
- Core forks in official mobile clients. Mobile support must be implemented in `mattermost-mobile` with feature flags/capability checks.

Conflict-avoidance constraints:

- Prefer external bot service + plugin packages.
- Avoid invasive changes under `server/channels/*` and `webapp/channels/*`.
- Use public APIs and plugin extension points.

## 4. Functional Requirements (Mattermost)

### FR-1 Stream lifecycle

System must support lifecycle states:

- `start`: user sends prompt, bot starts processing.
- `streaming`: partial output emitted incrementally.
- `end`: final answer published and stream closed.
- `error`: stream aborted/failed with user-visible fallback message.

### FR-2 Typing/status integration

During `streaming`:

- Bot must maintain visible activity indication:
  - baseline: standard typing API refresh loop
  - optional: plugin custom status start/refresh/stop
- Activity indicator must stop on `end` and `error`.

### FR-3 Partial text rendering

- Partial content must be shown progressively (chunk/delta updates).
- UI updates must be throttled to avoid event/message spam.
- Final response must be persisted as normal message content.

### FR-4 Fallback behavior

- If custom plugin UI is unavailable, user still receives clear progress:
  - typing indicator + final message
  - optional ephemeral progress message for long tasks
- Mobile clients must remain functional without plugin UI.

### FR-7 Multi-platform capability contract

- Stream/status events must include a client capability model so mobile can run in:
  - fallback mode (baseline, default)
  - enhanced mobile mode (feature-gated)
- Unknown fields must be forward-compatible so older clients degrade gracefully.

### FR-5 Abort and timeout

- User/system abort should stop stream cleanly.
- Idle timeout and overall timeout should be enforced.
- On timeout/abort, stop status and publish actionable error notice.

### FR-6 Observability

- Each request must have a correlation ID.
- Log stream start/end/error and key timings:
  - time to first token
  - completion duration
  - chunk count

## 5. Non-Functional Requirements

### NFR-1 Compatibility

- Upstream pulls should not require resolving custom core patches.
- Architecture must remain plugin/service based.

### NFR-2 Performance

- Stream handling must support high-frequency deltas without UI jank.
- Message updates should be batched/throttled where appropriate.

### NFR-3 Reliability

- If stop signal is missed, TTL cleanup must remove stale “thinking” states.
- Stream errors must not leave hanging indicators.

### NFR-4 Security

- Bot/plugin endpoints must enforce authentication and channel authorization.
- Payload validation required for custom status text/metadata.

## 6. Recommended Mattermost Architecture

1. Bot service (stream source):

- Connect to LLM provider with streaming API.
- Emit normalized events: `start`, `delta`, `complete`, `error`, `abort`.

2. Mattermost integration layer:

- On `start`: send typing start/refresh scheduler and (optional) plugin status start.
- On `delta`: update in-progress UI buffer (plugin UI or controlled message update path).
- On `complete`: post final answer, stop indicators.
- On `error/abort`: stop indicators, post failure/fallback notice.

3. Optional webapp plugin:

- Subscribe to plugin websocket events.
- Render custom rich status row (`🤖 Bot is thinking...`).
- Apply TTL cleanup and channel/thread scoping.

4. Client fallback:

- If plugin not available, rely on default typing + final message.
- For long tasks, optionally post ephemeral progress updates.

5. Mobile client adapter (`mattermost-mobile`):

- Ingest stream/status events and map them to channel/thread scoped UI state.
- Enforce TTL cleanup and deterministic ordering on iOS and Android.
- Use a feature flag to switch between fallback-only and enhanced rendering.

## 7. Event Contract (Draft)

Normalized stream event model:

```json
{
  "type": "start|delta|complete|error|abort",
  "request_id": "uuid",
  "channel_id": "string",
  "thread_id": "string_optional",
  "text": "string_optional",
  "timestamp": 1760000000000,
  "meta": {}
}
```

Status event model (plugin):

```json
{
  "type": "status_start|status_refresh|status_stop",
  "request_id": "uuid",
  "channel_id": "string",
  "actor_id": "string",
  "text": "thinking...",
  "emoji": "robot_face_optional",
  "expires_at": 1760000005000
}
```

## 8. Cherry Studio Streaming Implementation Record

The following files were reviewed in `/home/caoyujie/projects/cherry-studio` (depth-1 clone).

### 8.1 Core runtime always streams

- `RuntimeExecutor.streamText(...)` is the central runtime API for text generation.
- File: `packages/aiCore/src/core/runtime/executor.ts`

Observation:

- Streaming is first-class. Plugin engine can transform stream behavior before returning result.

### 8.2 App service wires chunk callback

- `fetchChatCompletion(...)` builds middleware config with `onChunk`.
- `streamOutput` setting is attached to middleware config.
- File: `src/renderer/src/services/ApiService.ts`

Observation:

- UI update callback is a first-class dependency of model invocation path.

### 8.3 Stream toggle via middleware plugin

- `buildPlugins(...)` adds `simulateStreamingPlugin` when `streamOutput` is false.
- File: `src/renderer/src/aiCore/plugins/PluginBuilder.ts`
- File: `src/renderer/src/aiCore/plugins/simulateStreamingPlugin.ts`

Observation:

- Even non-stream responses can be normalized into stream-like behavior, simplifying UI pipeline.

### 8.4 Adapter converts provider stream into internal chunk protocol

- `AiSdkToChunkAdapter.processStream(...)` consumes `fullStream`.
- Converts provider parts (`text-delta`, `reasoning-delta`, `tool-*`, `finish`, `abort`, `error`) into internal `ChunkType`.
- File: `src/renderer/src/aiCore/chunk/AiSdkToChunkAdapter.ts`

Observation:

- Adapter layer decouples provider specifics from UI/store logic.
- Includes abort/error mapping and timing metrics (`time_first_token`, `time_completion`).

### 8.5 Stream processor dispatches chunk events to callbacks

- `createStreamProcessor(...)` routes `ChunkType` to callbacks (`onTextChunk`, `onError`, etc.).
- File: `src/renderer/src/services/StreamProcessingService.ts`

Observation:

- A single dispatcher handles all chunk categories and keeps view logic modular.

### 8.6 Message thunk applies chunks to UI state incrementally

- Thunk path creates callbacks, emits initial response-created state, processes stream continuously, and finalizes loading state.
- File: `src/renderer/src/store/thunk/messageThunk.ts`

Observation:

- Message-level orchestration is explicit: pending state -> chunk updates -> completion/error cleanup.

## 9. Cherry Studio Patterns to Reuse in Mattermost

1. Build a normalized chunk/event protocol internal to your bot stack.
2. Keep provider parsing inside an adapter layer.
3. Keep UI updates behind callback handlers, not provider-specific branches.
4. Unify error/abort handling into the same stream state machine.
5. Add timing metrics from stream lifecycle.
6. Support non-stream providers through simulated streaming when needed.

## 10. Acceptance Criteria

- AC-1: User sees immediate activity signal after sending prompt.
- AC-2: User receives progressive partial output during long generations.
- AC-3: Final response is persisted as standard Mattermost message.
- AC-4: Abort/error paths stop indicators and provide clear user feedback.
- AC-5: Mobile clients remain usable via fallback behavior.
- AC-5a: Mobile fallback mode works on both iOS and Android.
- AC-5b: Enhanced mobile mode (when enabled) renders stream/status updates with correct TTL and thread scoping.
- AC-6: Implementation can be maintained across upstream pulls without core merge conflicts.

## 11. Implementation Phases

Phase 1:

- Define event protocol + bot stream adapter + logs/metrics.

Phase 2:

- Add Mattermost typing refresh loop integration for stream lifecycle.

Phase 3:

- Add optional plugin rich-status UI (web/desktop web).

Phase 4:

- Add fallback and enhanced mobile adapter; validate on web + Android + iOS.

Phase 5:

- Harden with load, abort, and timeout tests.

## 12. CI/CD Multi-Platform Execution

Use separate pipelines per repo and require all relevant checks before merge:

1. `mattermost` pipeline:
- server/webapp unit + integration tests
- plugin API contract tests for stream/status events
- feature-flag OFF regression checks

2. `desktop` pipeline:
- desktop web fallback behavior checks
- smoke test for stream lifecycle visibility in desktop client

3. `mattermost-mobile` pipeline:
- RN unit/integration tests for stream/status store
- Android build + stream fallback/enhanced tests
- iOS build + stream fallback/enhanced tests

Required merge gate:
- PR merge is blocked unless web/desktop/mobile pipelines all succeed for this feature.

## 13. OpenClaw Integration Decisions (2026-04-04)

This section records implementation decisions aligned with the current OpenClaw + Mattermost deployment model.

### 13.1 Ownership and boundary

- Mattermost-side stream UX contract remains plugin/service based (no Mattermost core fork).
- OpenClaw integration should be implemented in the Mattermost channel extension path, not by patching Mattermost core.
- OpenClaw upstream policy for this work:
  - Do not PR custom business behavior to official OpenClaw main for this feature.
  - Keep OpenClaw main from official release.
  - Build and ship a custom Mattermost extension artifact in a separate pipeline/repository.

### 13.2 OpenClaw current capability (observed)

OpenClaw Mattermost extension already has:
- reply dispatcher + typing callbacks and keepalive lifecycle
- block streaming controls (`blockStreaming`, coalesce config)
- payload delivery with chunking and media support

Key files (OpenClaw):
- `extensions/mattermost/src/mattermost/monitor.ts`
- `extensions/mattermost/src/mattermost/reply-delivery.ts`
- `extensions/mattermost/src/channel.ts`
- `extensions/mattermost/src/types.ts`
- `src/channels/typing.ts`

### 13.3 OpenClaw gaps to close for this plan

To match the Mattermost stream contract in this document, OpenClaw extension needs:

1. Explicit stream lifecycle contract emission:
- `start`, `delta`, `complete`, `error`, `abort`

2. Optional rich-status lifecycle integration:
- call custom status API/events (`status_start`, `status_refresh`, `status_stop`) when available
- keep typing-based fallback when unavailable

3. Observability alignment:
- correlation/request ID per inbound request
- timing metrics: TTFT, completion duration, chunk count

4. Capability-aware behavior:
- fallback mode default
- enhanced mode behind feature flag/capability gate

### 13.4 Build and deploy model (GitHub Actions + Docker Compose)

Policy:
- Never rely on local manual builds for test/prod promotion.
- Build custom extension artifact through GitHub Actions only.

Recommended pipeline stages:
1. extension CI (lint/test/contracts)
2. extension package artifact (`.tgz` + checksum + metadata)
3. manual promote-to-test workflow (artifact version input)
4. post-deploy smoke checks

Docker Compose test-environment apply path:
1. Use `openclaw-cli` container to install plugin artifact into mounted OpenClaw config home.
2. Enable plugin if needed.
3. Restart `openclaw-gateway`.
4. Verify with plugin list + channel smoke tests.

Operational command pattern (example):
- `docker compose run --rm openclaw-cli plugins install /artifacts/<plugin>.tgz --force`
- `docker compose run --rm openclaw-cli plugins enable mattermost`
- `docker compose restart openclaw-gateway`

### 13.5 Mattermost-side deliverables still required

- Implement/deploy rich-status plugin endpoints and websocket status broadcast.
- Webapp plugin rendering for status row + TTL cleanup + thread scoping.
- Keep clear fallback semantics for clients without custom rendering.
- Add versioned capability handshake between OpenClaw extension and Mattermost plugin/UI.

## 14. Client Implementation Checklist (Webapp/Desktop/Android/iOS)

Use this checklist to clarify per-platform ownership and release gates.

### 14.1 Webapp (`mattermost`)

- Render rich-status row from plugin websocket events.
- Scope status display correctly for channel vs thread.
- Support deterministic ordering for multiple active statuses.
- Enforce TTL-based auto-expiry cleanup.
- Preserve fallback path when plugin/status events are unavailable.
- Add tests for websocket event -> UI transition -> expiry lifecycle.

### 14.2 Desktop (`desktop`)

- Validate parity with webapp rendering in desktop runtime.
- Validate reconnect/reload behavior for active statuses.
- Validate fallback behavior when plugin/status capability is absent.
- Add desktop smoke coverage for lifecycle visibility and cleanup.

### 14.3 Android (`mattermost-mobile`)

Baseline (required):
- Ensure fallback mode works end-to-end (typing + final response).
- Ensure no regressions when rich-status capability is disabled.

Enhanced mode (feature-gated):
- Consume stream/status lifecycle events.
- Render channel/thread-scoped status UI.
- Apply TTL cleanup for stale state.
- Handle background/resume + reconnect without stale indicators.
- Add Android tests for `start|delta|complete|error|abort` and resume/reconnect cleanup.

### 14.4 iOS (`mattermost-mobile`)

Baseline (required):
- Ensure fallback mode works end-to-end (typing + final response).
- Ensure no regressions when rich-status capability is disabled.

Enhanced mode (feature-gated):
- Consume stream/status lifecycle events.
- Render channel/thread-scoped status UI.
- Apply TTL cleanup for stale state.
- Handle foreground/background transitions + reconnect correctly.
- Add iOS tests for `start|delta|complete|error|abort` and lifecycle cleanup.

### 14.5 Cross-platform release gates

- Webapp + desktop + Android + iOS checks must all pass for feature promotion.
- If enhanced mode fails on any client, keep fallback mode as default.
- Do not enable broad rollout until canary telemetry confirms no stale-status or stream-lifecycle regressions.

## 15. Multi-Platform Test Plan with GitHub Actions

This section defines CI execution for multi-platform verification.

### 15.1 CI scope and limitations

GitHub Actions is required for primary verification and release gating.

Actions can reliably cover:
- webapp/plugin contract tests
- desktop build/smoke tests
- Android emulator tests
- iOS simulator tests (macOS runners)

Actions cannot fully replace:
- real-device validation for platform-specific background/resume behavior
- some OS-specific desktop runtime quirks

### 15.2 Required GitHub Actions job groups

For each change set, require these job groups:

1. `webapp-contract-and-ui` (`mattermost`):
- plugin status event contract tests
- UI transition tests (`start|delta|complete|error|abort`)
- TTL expiry + thread-scope tests

2. `desktop-stream-smoke` (`desktop`):
- desktop packaging/build sanity
- stream status visibility smoke tests
- fallback behavior smoke tests

3. `mobile-android-stream` (`mattermost-mobile`):
- Android build
- emulator integration tests for fallback + enhanced mode
- reconnect/background-resume cleanup tests

4. `mobile-ios-stream` (`mattermost-mobile`):
- iOS build (simulator)
- simulator integration tests for fallback + enhanced mode
- foreground/background lifecycle cleanup tests

### 15.3 Suggested workflow triggers

- Pull request:
  - run full required job groups
  - block merge on failure
- Main branch:
  - run required job groups
  - publish test summary artifacts
- Nightly:
  - run expanded regression matrix and longer-running stability checks

### 15.4 Test artifacts and evidence

Each required workflow should upload:
- junit/test reports
- stream lifecycle trace logs (sanitized)
- failure screenshots/videos for UI tests where available
- environment metadata (commit SHA, runner OS, app version)

### 15.5 Promotion gates (GitHub Actions enforced)

Promote enhanced mode only when all are true:
- `webapp-contract-and-ui` passed
- `desktop-stream-smoke` passed
- `mobile-android-stream` passed
- `mobile-ios-stream` passed
- no critical regressions in stream lifecycle/TTL cleanup tests

If any required job fails:
- keep fallback mode default
- do not enable enhanced mode in broader environments

### 15.6 Manual device gate (post-CI, pre-broad rollout)

After Actions pass, require a lightweight manual validation:
- one real Android device pass
- one real iOS device pass
- verify no stale status after background/resume and reconnect

This manual gate is required before canary expansion beyond internal test users.
