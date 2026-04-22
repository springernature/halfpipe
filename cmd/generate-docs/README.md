# generate-docs

Generates `docs/manifest-reference.md` from `schema.json` and the example files in `docs/examples/`.

## How it works

The docs generator reads only `schema.json` as input — it never looks at Go source files directly. Struct descriptions, field descriptions, required/optional markers, and type information all flow through the schema as the single intermediate representation:

    Go struct comments → schema.json → manifest-reference.md

Examples are loaded from `docs/examples/` by convention:
- `trigger-<type>.yaml` for triggers
- `task-<type>.yaml` for tasks
- `type-<name>.yaml` for supporting types

Multiple examples in one file are separated by `---`.

## Usage

```
make docs
```

This runs `make schema` first (regenerating `schema.json` from Go structs), then runs this generator.
