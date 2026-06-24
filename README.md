# CommitCraft

CommitCraft is a CLI tool for rewriting Git history safely and cleanly.
It helps you edit commit messages, fix author metadata, adjust timestamps, preview rewrite ranges, lint weak commit messages, and recover from bad rewrites using backup branches.

```bash
commitcraft inspect   -n 20
commitcraft plan      --from HEAD~5
commitcraft recommit  --from HEAD~5
commitcraft rewrite   --hash a3f1c9b --message "feat(auth): add JWT refresh logic"
commitcraft author    --hash HEAD~2 --name "Alice" --email "alice@example.com"
commitcraft timestamp --hash HEAD --date "2024-03-15 09:30:00"
commitcraft lint      -n 30
commitcraft rollback  --backup backup/master-20260624-183500
```

---

## Features

- Rewrite a single commit message anywhere in history
- Interactively rewrite a range of commit messages
- Rewrite commit author name and email
- Rewrite commit timestamps
- Preview which commits will be affected before rewriting
- Bulk apply rewrite plans from YAML
- Lint weak commit messages
- Automatically create backup branches before destructive rewrites
- Roll back a branch to a backup branch if a rewrite goes wrong
- Handle root-commit rewrites safely

---

## Installation

### Prerequisites

- Go 1.21+
- Git 2.x

### Build from source

```bash
git clone https://github.com/yourname/commitcraft
cd commitcraft
go mod tidy
go build -o commitcraft .
```

### Verify

```bash
commitcraft --help
```

---

## Commands

## `inspect`

Display recent commit history in a clean readable format.

```bash
commitcraft inspect -n 20
commitcraft inspect --from HEAD~10 --detail
```

---

## `plan`

Preview which commits would be affected by a rewrite range.

```bash
commitcraft plan --from HEAD~5
```

---

## `recommit`

Interactively rewrite commit messages across a range.

```bash
commitcraft recommit --from HEAD~5
```

---

## `rewrite`

Rewrite the message of a single commit.

```bash
commitcraft rewrite --hash a3f1c9b --message "fix: correct pagination logic"
```

If the target commit is not HEAD, CommitCraft creates a backup branch before rewriting history.

---

## `author`

Rewrite author name and email for one commit or all commits.

```bash
commitcraft author --hash HEAD~2 --name "Alice Smith" --email "alice@example.com"
commitcraft author --all --name "Alice Smith" --email "alice@example.com"
```

---

## `timestamp`

Rewrite the author and committer date for a commit.

```bash
commitcraft timestamp --hash HEAD --date "2024-03-15 09:30:00"
```

Accepted formats:

- `2006-01-02 15:04:05`
- `2006-01-02T15:04:05`
- `2006-01-02`
- RFC3339

---

## `lint`

Scan recent commits and flag weak commit messages such as:

- `fix`
- `update`
- `misc`
- `temp`
- `wip`

```bash
commitcraft lint -n 30
```

---

## `apply`

Apply bulk commit rewrites from a YAML plan file.

```bash
commitcraft apply --file rewrite.yaml
```

Example YAML:

```yaml
rewrites:
  - hash: "abc1234"
    message: "feat: improve auth flow"
  - hash: "def5678"
    author_name: "Lakshmi Priya"
    author_email: "lakshmi@example.com"
```

---

## `rollback`

Restore the current branch from a backup branch created before a rewrite.

```bash
commitcraft rollback --backup backup/master-20260624-183500
```

---

## Safety behavior

For destructive history rewrites, CommitCraft adds safety checks:

- verifies the repository is valid
- checks the working tree before destructive rewrites
- creates a timestamped backup branch before older-commit rewrites
- supports rollback from backup branches
- safely handles root-commit rewrites

Example backup branch:

```bash
backup/master-20260624-183500
```

---

## Global Flags

These work with every subcommand:

| Flag | Description |
|------|-------------|
| `--repo` | Path to the git repository (default: `.`) |
| `--verbose` / `-v` | Print the git commands being run |
| `--config` | Path to config file (default: `~/.commitcraft.yaml`) |

---

## Config file

CommitCraft reads `~/.commitcraft.yaml` on startup.

Example:

```yaml
author_name: "Alice Smith"
author_email: "alice@example.com"
backup_branch: true
verbose: false
```

---

## Force-pushing rewritten history

Any rewrite changes commit hashes.
If commits were already pushed, push rewritten history carefully:

```bash
git push --force-with-lease origin main
```

Use `--force-with-lease` instead of plain `--force` whenever possible.

---

## License

MIT
