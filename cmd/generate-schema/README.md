# generate-schema

Generates the JSON Schema for the halfpipe manifest file (`.halfpipe.io`, `.halfpipe.io.yml`, `.halfpipe.io.yaml`).

## Usage

```sh
make schema
```

This runs the generator and writes the output to `schema.json` at the repo root. The schema is also validated against the e2e test fixtures as part of `make build`.

To run manually:

```sh
go run ./cmd/generate-schema > schema.json
```

## How it works

1. Reflects `manifest.Manifest` using [invopop/jsonschema](https://github.com/invopop/jsonschema) to produce a base schema.
2. Overrides `tasks` and `triggers` arrays with `oneOf` discriminated unions — each variant identified by a `"type"` const (e.g. `"run"`, `"deploy-cf"`).
3. Adds `additionalProperties: false` to every task and trigger definition.
4. Applies fixups for types that diverge between Go and JSON representations (`Vars`, `ComposeFiles`, `Platform`, `feature_toggles`).

## Referencing the schema in a manifest file

Editors that support the [YAML Language Server](https://github.com/redhat-developer/yaml-language-server) (VS Code, IntelliJ, Neovim, etc.) can use the schema for validation and autocompletion.

Add this comment as the first line of your `.halfpipe.io.yml`:

```yaml
# yaml-language-server: $schema=https://github.com/springernature/halfpipe/releases/latest/download/schema.json
team: my-team
pipeline: my-pipeline
# ...
```

For a local schema (e.g. during development):

```yaml
# yaml-language-server: $schema=../../schema.json
```

## Adding a new task or trigger type

1. Add the Go struct to the `manifest` package.
2. Add an entry to `taskTypes` (or `triggerTypes`) in `main.go`.
3. Run `make schema` to regenerate `schema.json`.

The `oneOf` refs and `$defs` entries are derived from those slices automatically — no other changes needed in this package.
