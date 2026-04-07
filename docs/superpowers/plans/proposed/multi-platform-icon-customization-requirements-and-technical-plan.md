# Multi-Platform Mattermost Icon Customization: Requirements and Technical Plan

**Status**: proposed
**Doc-Type**: technical-plan
**Scope**: mattermost, desktop, mobile, branding
**Owner**: tbd
**Supersedes**: n/a
**Superseded-by**: n/a
**Last-Updated**: 2026-04-04
**Tags**: mattermost, branding, icon, webapp, desktop, mobile, android, ios, technical-plan, cicd

## 1. Objective

Define a safe path to replace branding/icon assets across web, desktop, Android, and iOS with CI validation and rollback support.

## 2. Requirements

- Keep one canonical source icon set.
- Ensure platform-native formats remain valid (`.icns`, `.ico`, `.png`, adaptive icon XML).
- Preserve launcher/app icon consistency across clients.

## 3. Implementation

- `mattermost`: web branding/logo assets.
- `desktop`: Electron app icon assets for each OS target.
- `mattermost-mobile`: iOS AppIcon set + Android mipmap/adaptive icon assets.

## 4. CI/CD

- Validate asset presence and dimensions.
- Run web/desktop/mobile build checks.
- Promote only when all platform jobs pass.

## 5. Rollout and Rollback

- Stage to internal test first.
- Promote after cross-platform verification.
- Rollback by restoring previous icon asset commit and rebuilding artifacts.
