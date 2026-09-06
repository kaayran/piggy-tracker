---
name: ship
description: Open a GitHub pull request for finished work in piggy-tracker — branch, commit, push as kaayran, PR into master with kaayran assigned. Use as the last step of every feature or fix once the work is done and verified, and whenever the user says "ship", "открой PR", "сделай PR", "залей".
---

# Ship

Development flow in this repo: agent builds the feature → this skill opens a PR → the user reviews and merges on GitHub. The agent never pushes to `master` and never merges.

Invoking this skill is the authorization to commit (it overrides "Commit only when asked" in CLAUDE.md). Nothing else here overrides CLAUDE.md.

## Before shipping

- The work is finished and verified. A half-done change is not a PR.
- The feature has a spec in `docs/features/`. No spec → write the spec first, ship it in the same PR.
- `git status` is clean of junk: no scratch files, no debug leftovers.

## Steps

1. **Push identity.** `git push` over HTTPS otherwise resolves to `pokidov-a`, who has no write access. Idempotent, run every time:

   ```sh
   git config --local --unset-all credential.https://github.com.helper || true
   git config --local --add credential.https://github.com.helper ""
   git config --local --add credential.https://github.com.helper "!gh auth git-credential"
   gh auth switch -u kaayran
   ```

2. **Branch.** On `master` → create one: `feat/<slug>`, `fix/<slug>`, `docs/<slug>`. Already on a feature branch → stay on it.

3. **Commit.** CLAUDE.md style: one short imperative line under 150 chars, English, no body unless genuinely non-obvious. No `Co-Authored-By`, no `Generated with`, no agent attribution — not in the commit, not in the PR.

4. **Push.** `git push -u origin HEAD`

5. **PR.**

   ```sh
   gh pr create --base master --assignee kaayran --title "<title>" --body "<body>"
   ```

   Title: same style as the commit. Body, English, short:
   - what changed, one or two sentences
   - link to the spec: `docs/features/<name>.md`
   - how to check it (command, or the screen to open)

   No `--reviewer`: GitHub rejects a self-review request and kaayran owns the repo and authors the PR. `--assignee` puts it in the user's list instead.

6. **Report the PR URL and stop.** Do not merge, do not delete the branch, do not push follow-up commits unless asked.

## Follow-ups

Review comments arrive on GitHub. Fixing them = commit onto the same branch and `git push`; the PR updates itself. Only open a new PR for a new feature.
