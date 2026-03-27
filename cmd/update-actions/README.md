# update-actions

Updates the SHA-pinned GitHub Actions in `renderers/actions/external_actions.go` and the corresponding `workflowExpected.yml` files in `e2e/actions/` to their latest released versions.

## Usage

```sh
# Dry run — see what would change without modifying any files
go run ./cmd/update-actions/ --dry-run

# Apply updates
go run ./cmd/update-actions/
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Show changes without modifying any files |

## Behaviour

- Only updates **SHA-pinned** actions (lines matching `"owner/repo@<40-char-sha>" // vN` or `// vN.N.N`).
- **Skips** tag-pinned actions (e.g. `springernature/ee-action-buildpack@v1`).
- Resolves the latest version via the GitHub **Releases** API, falling back to the **Tags** API if no releases exist.
- Handles both lightweight and annotated git tags when resolving commit SHAs.
- Writes the **full semver** tag into the version comment (e.g. `// v4.0.0`).
- **Warns** on major version bumps with a `!` prefix and `WARNING: major version bump!` message.
- Also updates any matching SHAs in `e2e/actions/*/workflowExpected.yml` files.

## Authentication

Set the `GITHUB_TOKEN` environment variable for authenticated requests. Without it, the script falls back to unauthenticated requests (rate limited to 60 requests/hour).

```sh
export GITHUB_TOKEN=ghp_...
go run ./cmd/update-actions/ --dry-run
```

## Example output

```
Using GITHUB_TOKEN for authentication
Parsing renderers/actions/external_actions.go...
Found 8 SHA-pinned actions

  actions/checkout                              v6          (up to date)
! actions/download-artifact                     v7          -> v8.0.1     WARNING: major version bump!
! actions/upload-artifact                       v6          -> v7.0.0     WARNING: major version bump!
! docker/login-action                           v3          -> v4.0.0     WARNING: major version bump!
  hashicorp/vault-action                        v3          (up to date)

Updating e2e expected workflow files...
  e2e/actions/artifacts/workflowExpected.yml                   (5 replacements)
  e2e/actions/deploy-cf/workflowExpected.yml                   (5 replacements)
  e2e/actions/docker-push/workflowExpected.yml                 (3 replacements)
  e2e/actions/notifications/workflowExpected.yml               (3 replacements)

Dry run complete. No changes written.
```
