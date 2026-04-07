# Plans Index

This directory is organized by lifecycle state:

- `active/`: approved plans currently in execution
- `proposed/`: draft plans pending approval
- `archive/`: completed or superseded plans
- `runbooks/`: operational procedures

## Inventory

| Topic | Type | ID | Status | Path |
|---|---|---|---|---|
| Stream Output | requirements | 2026-04-03 | active | `active/2026-04-03-mattermost-stream-output-requirements-and-cherry-study.md` |
| Stream Output | technical-plan | 2026-04-04 | active | `active/2026-04-04-end-to-end-stream-output-technical-plan.md` |
| Rich Typing Plugin | requirements | 2026-04-03 | active | `active/2026-04-03-rich-typing-plugin-requirements.md` |
| Rich Typing Plugin | technical-plan | 2026-04-03 | active | `active/2026-04-03-rich-typing-plugin-technical-plan.md` |
| Relaxed Bot Username | requirements | 2026-04-03 | active | `active/2026-04-03-relaxed-bot-username-requirements.md` |
| Relaxed Bot Username | technical-plan | 2026-04-03 | active | `active/2026-04-03-relaxed-bot-username-technical-plan.md` |
| Multi-Platform Icon Customization | technical-plan | 2026-04-04 | proposed | `proposed/multi-platform-icon-customization-requirements-and-technical-plan.md` |
| Data Retention Policy | archive | 2026-03-29 | archived | `archive/2026-03-29-data-retention-policy.md` |
| Relaxed Bot Username (Legacy Combined Doc) | archive | 2026-04-03 | archived | `archive/2026-04-03-relaxed-bot-username-requirements-and-technical-plan.md` |
| CI/CD Customization Workflow | runbook | 2026-04-04 | runbook | `runbooks/2026-04-04-cicd-customization-workflow-runbook.md` |
| Logo Customization Docker Redeploy | runbook | 2026-04-04 | runbook | `runbooks/2026-04-04-logo-customization-docker-redeploy-plan.md` |
| Mattermost Language Switch | runbook | n/a | runbook | `runbooks/mattermost-language-switch-runbook.md` |

## Maintenance Rules

1. New planning docs start in `proposed/`.
2. Move to `active/` only after approval.
3. Move to `archive/` when done or superseded.
4. Put operational how-to procedures in `runbooks/`.
5. Update this index in the same change whenever files are added/moved.
6. Every plan/runbook must include metadata headers:
   - `Status`, `Doc-Type`, `Scope`, `Owner`, `Supersedes`, `Superseded-by`, `Last-Updated`, `Tags`.
7. For active topics, maintain two docs:
   - one `requirements` doc
   - one `technical-plan` doc

## Tag Taxonomy (Controlled)

Use comma-separated tags from these groups:

- Domain: `streaming`, `typing`, `username`, `mention`, `branding`, `icon`, `localization`, `policy`, `cicd`
- Platform: `webapp`, `desktop`, `mobile`, `android`, `ios`
- System: `mattermost`, `openclaw`, `plugin`, `docker`, `github-actions`
- Lifecycle: `fallback`, `enhanced-mode`, `feature-flag`, `runbook`, `archive`, `rollout`, `rollback`

Avoid ad-hoc tag variants unless a new term is intentionally introduced across multiple docs.
