# 🔄 Git Workflow & PR Guide — go-beats

> This guide explains how developers should work with branches, commits, and pull requests in this project.

---

## 🎫 Ticket Pickup Rules

> **Every dev must follow these rules when picking up a ticket. No exceptions.**

### Step 1: Pick a Ticket

1. Go to the [Issues board](https://github.com/rolniuq/go-beats/issues) and find an **open, unassigned** ticket from the current sprint milestone
2. **Priority order**: Pick `priority:high` tickets before `priority:medium` or `priority:low`
3. **Only pick ONE ticket at a time** — finish your current ticket before picking a new one
4. **Do NOT pick a ticket that is already assigned to someone else**

### Step 2: Assign Yourself

1. **Assign yourself** to the ticket on GitHub (right sidebar → "Assignees")
2. This tells the team the ticket is taken — no one else should work on it

### Step 3: Create Your Branch

1. **Always branch from latest `main`**:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b <type>/ticket-<number>-<short-description>
   ```
2. Branch name **must include the ticket number** (see Branch Naming Convention below)

### Step 4: Work on the Ticket

1. Read the ticket description carefully — follow the **Acceptance Criteria**
2. If anything is unclear, **ask the tech lead before coding** — don't guess
3. Keep your changes **scoped to the ticket** — don't fix unrelated things in the same branch
4. Make small, meaningful commits as you go (see commit rules below)

### Step 5: Self-Check Before PR

Before creating a PR, verify:
1. ✅ `task check` passes (fmt + vet + test)
2. ✅ All **Acceptance Criteria** from the ticket are met
3. ✅ No hardcoded values, secrets, or debug code left behind
4. ✅ Code is readable and commented where needed
5. ✅ You've tested manually (follow "How to Test" if provided in the ticket)

### Step 6: Create PR & Link the Ticket

1. Push your branch and create a PR following the PR template below
2. **Must include** `Closes #<ticket-number>` in the PR body — this auto-closes the ticket when merged
3. **Assign tech lead as reviewer**
4. Wait for review — **do NOT merge your own PR**

### Step 7: Address Review Feedback

1. If changes are requested → fix them, push to the **same branch**, and re-request review
2. **Do NOT create a new PR** for review fixes — keep everything in one PR
3. Respond to every review comment — either fix it or explain why not

### Step 8: After Merge

1. Tech lead merges the PR (squash merge)
2. Delete your feature branch:
   ```bash
   git checkout main
   git pull origin main
   git branch -d <your-branch-name>
   ```
3. Go back to **Step 1** and pick your next ticket

### ❌ Common Mistakes to Avoid

| Mistake | Why it's bad |
|---------|-------------|
| Working on a ticket without assigning yourself | Other devs might start the same work → wasted effort |
| Picking multiple tickets at once | Context switching = slower delivery, more merge conflicts |
| Skipping `task check` before pushing | Broken builds block the whole team |
| Making changes outside the ticket scope | Makes PRs hard to review and can introduce bugs |
| Forgetting `Closes #N` in PR body | Ticket stays open even after merge → confusing board |
| Merging your own PR | Violates code review process — always wait for tech lead |
| Pushing directly to `main` | Branch protection will reject it anyway |

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
