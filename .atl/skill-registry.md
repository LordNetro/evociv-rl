# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When creating a pull request, opening a PR, or preparing changes for review | branch-pr | `C:\Users\marco\.config\opencode\skills\branch-pr\SKILL.md` |
| When working with Engram in a team, syncing observations to a shared repo, setting up team memory, or asking about scope/project/personal conventions | engram-team-sync | `C:\Users\marco\.config\opencode\skills\engram-team-sync\SKILL.md` |
| When writing Go tests, using teatest, or adding test coverage | go-testing | `C:\Users\marco\.config\opencode\skills\go-testing\SKILL.md` |
| When creating a GitHub issue, reporting a bug, or requesting a feature | issue-creation | `C:\Users\marco\.config\opencode\skills\issue-creation\SKILL.md` |
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen" | judgment-day | `C:\Users\marco\.config\opencode\skills\judgment-day\SKILL.md` |
| When user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | `C:\Users\marco\.config\opencode\skills\skill-creator\SKILL.md` |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### branch-pr
- Every PR MUST link an approved issue (`status:approved` label) — no exceptions
- Every PR MUST have exactly one `type:*` label
- Branch names MUST match `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`
- Commit messages MUST match conventional commit format: `type(scope): description`
- PR body must include: linked issue (Closes/Fixes/Resolves #N), PR type checkbox, summary, changes table, test plan, contributor checklist
- Automated checks (PR validation, shellcheck) must pass before merge
- Never add `Co-Authored-By` trailers

### engram-team-sync
- Team-shared observations use `scope: project`; personal ones use `scope: personal`
- Use `mem_save` with `topic_key` for deterministic upserts across team members
- Run `engram sync` to export/import observations from a shared git repo
- Language strategy: use English for code comments and commit messages, prefer user's language for documentation
- When troubleshooting import errors, verify the directory structure matches engram's expected layout

### go-testing
- Use table-driven tests (standard Go pattern): `[]struct{name, input, expected, wantErr}` + `t.Run(tt.name, ...)`
- Test Bubbletea models by calling `Model.Update(msg)` directly and asserting state transitions
- Use `teatest.NewTestModel(t, m)` for full TUI integration tests
- Use golden file testing for visual output: write to `testdata/*.golden`, compare with `--update` flag
- Organize test files alongside source: `model.go` → `model_test.go`, one file per concern
- Mock dependencies via interfaces, use `t.TempDir()` for temp files
- Run: `go test ./...`, `go test -v`, `go test -cover`, `go test -short`

### issue-creation
- Every issue MUST use the template with: summary, reproduction steps, expected vs actual behavior, environment
- Bug reports MUST include `type:bug` label, feature requests MUST include `type:feature`
- Issues MUST be triaged by a maintainer before work begins
- Every issue MUST link to a project board if it's part of a milestone
- Duplicate issues should be closed with a reference to the original
- Issues without enough information to reproduce should be labeled `status:needs-info`

### judgment-day
- Launch TWO independent blind judge sub-agents to review the same target simultaneously
- Judges do NOT see each other's findings until both complete
- Synthesize findings: unify confirmed issues, investigate discrepancies, fix confirmed bugs
- Re-judge after fixes: both judges must pass, or escalate after 2 iterations
- Use for: high-confidence review before merging, complex implementations, catching edge cases
- Never skip: both judges review independently, no collusion

### skill-creator
- Create skill when: pattern is used repeatedly, project-specific conventions differ from generic, complex workflows need step-by-step, decision trees needed
- Don't create for: existing documentation, trivial patterns, one-off tasks
- Structure: `skills/{name}/SKILL.md` + optional `assets/` and `references/`
- Frontmatter MUST have: name, description (with Trigger:), license (Apache-2.0), metadata.author (gentleman-programming), metadata.version
- Use `assets/` for code templates/schemas; use `references/` to link local docs
- Name conventions: generic=`{technology}`, project-specific=`{project}-{component}`, testing=`{project}-test-{component}`, workflow=`{action}-{target}`
- Register in AGENTS.md after creation

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| No convention files found | — | Project has no AGENTS.md, CLAUDE.md, .cursorrules, GEMINI.md, or copilot-instructions.md |
