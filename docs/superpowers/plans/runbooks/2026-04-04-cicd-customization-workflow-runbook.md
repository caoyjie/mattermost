# Mattermost Customization CI/CD Workflow Runbook

**Status**: runbook
**Doc-Type**: runbook
**Scope**: cicd, operations
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: cicd, github-actions, plugin, openclaw, runbook, operations, docker

## 1. Purpose

Operational steps to build and promote custom Mattermost/OpenClaw customization artifacts via GitHub Actions.

## 2. Standard Flow

1. Push branch with changes.
2. Run PR checks (required CI).
3. Build immutable artifact (`.tgz` + checksum).
4. Promote selected artifact to test using manual workflow dispatch.
5. Run smoke tests and collect evidence.

## 3. Promotion Inputs

- artifact SHA/version
- target environment (`test`)
- optional rollback baseline artifact

## 4. Post-Deploy Validation

- Plugin list and enablement status
- Basic channel reply flow
- Stream lifecycle/fallback behavior checks

## 5. Rollback

- Redeploy previous known-good artifact
- Restart gateway
- Confirm health and plugin status
