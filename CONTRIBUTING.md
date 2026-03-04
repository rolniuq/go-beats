# 🔄 Git Workflow & PR Guide — go-beats

> This guide explains how developers should work with branches, commits, and pull requests in this project.

---

## 📐 Branch Naming Convention

```
feat/<ticket-id>-short-description    # New features
fix/<ticket-id>-short-description     # Bug fixes
test/<ticket-id>-short-description    # Adding tests
refactor/<ticket-id>-short-description # Code refactoring
```

**Examples:**
```
feat/ticket-1-radio-tui
feat/ticket-2-radio-engine
test/ticket-3-tests
fix/ticket-4-reconnect
```

---

## 🚀 Developer Workflow (Step-by-Step)

### 1. Clone & Setup (First Time Only)

```bash
git clone git@github.com:rolniuq/go-beats.git
cd go-beats
task deps
task music-gen    # generate test mp3 files
task build        # verify it builds
```

### 2. Start Working on a Ticket

```bash
# Always start from latest main
git checkout main
git pull origin main

# Create your feature branch
git checkout -b feat/ticket-1-radio-tui
```

### 3. Make Small, Meaningful Commits

```bash
# Stage your changes
git add internal/ui/tui.go

# Commit with a clear message following this format:
#   <type>(<scope>): <what you did>
#
# Types: feat, fix, refactor, test, docs, chore
# Scope: ui, audio, radio, pomodoro, cmd

git commit -m "feat(ui): add tab switching between local and radio mode"
git commit -m "feat(ui): render station list in radio mode"
git commit -m "feat(ui): connect volume controls to radio player"
```

**Commit message rules:**
- ✅ `feat(ui): add station list with cursor navigation`
- ✅ `fix(radio): handle stream disconnect gracefully`
- ❌ `updated stuff`
- ❌ `wip`
- ❌ `fix bug`

### 4. Run Checks Before Pushing

```bash
# This runs fmt + vet + test — ALL must pass
task check
```

### 5. Push Your Branch

```bash
git push -u origin feat/ticket-1-radio-tui
```

### 6. Create a Pull Request

Use the GitHub CLI:

```bash
gh pr create \
  --title "feat(ui): integrate radio player into TUI" \
  --body "$(cat <<'EOF'
## Summary
- Add Tab key to switch between Local and Radio mode
- Render station list with genre and description
- Connect play/pause/volume/next/prev to radio player
- Show 📻 LIVE indicator and connection status

## Related Issue
Closes #1

## Changes
- `internal/ui/tui.go` — Added Mode type, radio rendering, key handlers

## How to Test
1. `task build && ./go-beats`
2. Press `Tab` → Radio mode appears
3. Select a station with `Enter` → Audio streams
4. Press `Space` → Pause/resume works
5. Press `Tab` → Back to Local mode

## Checklist
- [ ] `task check` passes (fmt + vet + test)
- [ ] No hardcoded values or secrets
- [ ] Code is readable and commented where needed
EOF
)"
```

Or create via GitHub web UI at: `https://github.com/rolniuq/go-beats/compare`

---

## 📝 PR Template

When creating a PR, use this structure:

```markdown
## Summary
<!-- 1-3 bullet points: what does this PR do? -->

## Related Issue
<!-- Link the ticket: Closes #1 -->

## Changes
<!-- List files changed and what was done -->

## How to Test
<!-- Step-by-step instructions for the reviewer -->

## Checklist
- [ ] `task check` passes (fmt + vet + test)
- [ ] No hardcoded values or secrets
- [ ] Code is readable and commented where needed
```

---

## 👀 Code Review Process

1. **Developer** creates PR → assigns **tech lead** as reviewer
2. **Tech lead** reviews:
   - Code quality & readability
   - Architecture alignment
   - Error handling
   - No breaking changes to existing features
3. **If changes requested** → Developer fixes and pushes to the same branch
   - Stale reviews are automatically dismissed (you must re-request review)
4. **If approved** → Tech lead merges via **"Squash and merge"**
5. **After merge** → Developer deletes their feature branch:
   ```bash
   git checkout main
   git pull origin main
   git branch -d feat/ticket-1-radio-tui
   ```

---

## ⚠️ Rules

| Rule | Why |
|------|-----|
| **Never push directly to `main`** | Branch is protected, PRs required |
| **Never force push to `main`** | Force push is disabled |
| **1 approval required** to merge | Tech lead must review |
| **Stale reviews auto-dismissed** | If you push new commits, re-request review |
| **Run `task check` before pushing** | Keep the build green |
| **One ticket = one branch = one PR** | Keep changes focused and reviewable |
| **Keep PRs small** | Easier to review, less risk |

---

## 🔀 Handling Merge Conflicts

If `main` has moved ahead while you were working:

```bash
# Update your local main
git checkout main
git pull origin main

# Rebase your branch on top of latest main
git checkout feat/ticket-1-radio-tui
git rebase main

# Resolve any conflicts, then:
git add .
git rebase --continue

# Force push your rebased branch (this is safe for feature branches)
git push --force-with-lease origin feat/ticket-1-radio-tui
```

---

## 📊 Quick Reference

```
main (protected)
  │
  ├── feat/ticket-1-radio-tui      → PR #X → review → squash merge
  ├── feat/ticket-2-radio-engine    → PR #X → review → squash merge
  ├── test/ticket-3-tests           → PR #X → review → squash merge
  └── fix/ticket-4-reconnect        → PR #X → review → squash merge
```

**Assign PRs to:** Tech Lead (repo owner)
**Labels to use:** `priority:high`, `priority:medium`, `frontend`, `backend`, `testing`, `reliability`
