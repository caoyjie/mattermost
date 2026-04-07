# Relaxed Bot Username Requirements

**Status**: active
**Doc-Type**: requirements
**Scope**: mattermost, webapp, desktop, mobile
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, username, mention, webapp, desktop, mobile, android, ios, requirements, feature-flag, cicd

## 1. Objective

Allow relaxed bot username formats (Unicode + curated symbols) without breaking mention, notification, and client compatibility.

## 2. Functional Requirements

- Dual validation modes:
  - strict (default)
  - relaxed (feature flag)
- Consistent behavior across create/update entrypoints.
- Mention parser/autocomplete/notification compatibility for allowed characters.
- API and mmctl consistency for same input.
- Backward compatibility for existing strict usernames.

## 3. Non-Functional Requirements

- Feature flag default OFF.
- Observability for validation and mention failures.
- Upgrade-safe implementation boundaries.

## 4. Platform Requirements

Clients to validate:
- webapp
- desktop
- iOS
- Android

For each:
- mention autocomplete
- mention rendering
- mention notification
- user/profile lookup behavior

## 5. Release Requirements

Do not enable relaxed mode in production unless:
- mention-notification e2e is clean
- API/mmctl consistency checks pass
- mobile parity checks pass (iOS + Android)
- canary observability is stable
