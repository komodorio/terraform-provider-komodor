---
name: release-notes
description: Generate release notes by comparing git log since the last tag and suggest next semver version
user-invocable: true
---

Run `git tag --sort=-version:refname | head -1` to find the last release tag, then `git log <tag>..HEAD --oneline` to list all commits since then.

Group the commits into these sections (omit empty sections):
- **New Features** — feat: commits
- **Bug Fixes** — fix: commits
- **Improvements** — refactor:, perf:, chore: commits that are user-visible
- **Internal** — test:, docs:, ci:, build: commits (keep very brief or omit)

Then suggest the next version number:
- **major** bump if there are breaking changes
- **minor** bump if there are new features
- **patch** bump if only bug fixes or internal changes

Format output as short markdown bullets suitable for a GitHub release body.
