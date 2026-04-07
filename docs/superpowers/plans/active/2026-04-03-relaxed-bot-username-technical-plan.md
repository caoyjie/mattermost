# Relaxed Bot Username Technical Plan

**Status**: active
**Doc-Type**: technical-plan
**Scope**: mattermost, webapp, desktop, mobile
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, username, mention, webapp, desktop, mobile, android, ios, technical-plan, feature-flag, cicd

## 1. Server Changes

1. Add mode-aware username validator:
- strict regex unchanged
- relaxed allowlist regex guarded by feature flag

2. Keep reserved-name protections and length constraints.

3. Apply validator consistently across API/mmctl create/update paths.

## 2. Mention Pipeline Changes

- Align mention parsing classes with relaxed allowlist.
- Add explicit reject path for unsupported punctuation patterns.
- Ensure deterministic behavior for autocomplete and notifications.

## 3. Feature Flag and Rollout

- Use `MM_EXPERIMENTAL_RELAXED_USERNAME` (default OFF).
- Stage rollout: internal -> canary -> broader enablement.
- Immediate rollback by disabling flag.

## 4. Test Plan

1. Unit tests:
- strict vs relaxed validation
- reserved names
- boundary lengths

2. Integration tests:
- API/mmctl consistency
- lookup by username

3. E2E tests:
- mention creation/render/notification
- thread and edited message flows

4. Multi-platform tests:
- webapp/desktop/iOS/Android mention parity
- reconnect/background behavior on mobile

## 5. CI/CD Gates

Required checks:
- `username-server-validation`
- `username-api-mmctl-consistency`
- `username-desktop-smoke`
- `username-mobile-android`
- `username-mobile-ios`

Enable flag only when all required checks and canary telemetry pass.
