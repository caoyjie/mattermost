# Mattermost Language Switch Runbook (Docker)

**Status**: runbook
**Doc-Type**: runbook
**Scope**: mattermost, localization, operations
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, localization, i18n, runbook, operations, docker

## 1. Purpose

How to switch Mattermost UI language reliably in Docker environments when profile language UI is unavailable or ineffective.

## 2. Root Cause Pattern

Landing page language may follow server default, while post-login language follows user-level locale preference.

## 3. Operational Steps

1. Verify server locale settings (`DefaultClientLocale`, `AvailableLocales`).
2. Update user locale if needed (DB/admin path).
3. Re-login clients and restart desktop app if necessary.

## 4. Validation

- Browser language after login
- Desktop language after full restart
- Consistent locale across sessions
