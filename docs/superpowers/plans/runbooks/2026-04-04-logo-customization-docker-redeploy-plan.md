# Mattermost Logo Customization Docker Redeploy Implementation Plan

**Status**: runbook
**Doc-Type**: runbook
**Scope**: mattermost, desktop, mobile, branding
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, branding, icon, docker, deployment, runbook, operations

## 1. Purpose

Runbook for deploying logo/icon customization updates in Docker-based environments.

## 2. Steps

1. Prepare updated branding assets.
2. Build required artifacts in CI.
3. Redeploy service/images in test environment.
4. Verify web/desktop/mobile branding surfaces.

## 3. Verification

- Web logo and favicon checks
- Desktop app icon checks
- Android/iOS launcher icon checks

## 4. Rollback

- Restore prior artifact/image tag
- Restart affected services
- Re-run verification checklist
