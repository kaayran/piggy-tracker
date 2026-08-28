# CLAUDE.md

Project context: [README.md](./README.md) (stack, layout), [DESIGN.md](./docs/DESIGN.md) (cross-cutting product decisions), [docs/features/](./docs/features) (one spec per feature, with its data model).

## Approach

Ponytail mode, level **full**, in every session in this repo — regardless of whether the
plugin hook fired. Laziest solution that works: does it need to exist at all, then stdlib,
then native platform feature, then an already-installed dependency, then the shortest code
that works. No speculative abstractions, no scaffolding "for later". Never lazy about
understanding the problem, input validation, error handling or accessibility.

## Rules

- All documentation, code, identifiers and commit messages in English.
- No comments in code unless explicitly asked.
- `api/openapi.yaml` is the source of truth. Change the contract first, then regenerate types for both sides — never hand-write what is generated.
- Don't add UI kits or CSS frameworks. Telegram theme variables `--tg-theme-*` plus own components.
- Scope is what has a spec in `docs/features/`. A new idea gets a spec first, then code — never the other way round.

## Commits

- One short line, under 150 characters, imperative mood. No body unless the change is genuinely non-obvious.
- No `Co-Authored-By`, no `Generated with`, no agent attribution.
- Commit only when asked.

Good: `add transaction create endpoint`
Bad: `feat(api): add transaction create endpoint with validation, error handling and tests` + three paragraphs.
