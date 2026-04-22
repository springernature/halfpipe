package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/invopop/jsonschema"
	"github.com/springernature/halfpipe/manifest"
)

func main() {
	schema := buildSchema()

	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling schema: %v\n", err)
		os.Exit(1)
	}

	if _, err := fmt.Println(string(out)); err != nil {
		fmt.Fprintf(os.Stderr, "error writing schema: %v\n", err)
		os.Exit(1)
	}
}

// reflectorConfig is the shared configuration for all jsonschema reflectors used
// in this package.
var reflectorConfig = jsonschema.Reflector{
	DoNotReference:             false,
	Anonymous:                  true,
	ExpandedStruct:             true,
	RequiredFromJSONSchemaTags: true, // only fields tagged `jsonschema:"required"` are required
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}

func buildSchema() *jsonschema.Schema {
	r := reflectorConfig

	root, err := findModuleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not find module root: %v\n", err)
		os.Exit(1)
	}
	// AddGoComments requires a relative path so the package keys are built correctly.
	// Temporarily change to the module root so "./manifest" resolves properly.
	if cwd, err := os.Getwd(); err == nil && cwd != root {
		_ = os.Chdir(root)
		defer os.Chdir(cwd)
	}
	if err := r.AddGoComments("github.com/springernature/halfpipe", "./manifest", jsonschema.WithFullComment()); err != nil {
		fmt.Fprintf(os.Stderr, "error: could not load go comments: %v\n", err)
		os.Exit(1)
	}

	// Reflect the top-level manifest to get all $defs populated
	topSchema := r.Reflect(manifest.Manifest{})

	// topSchema is already expanded (ExpandedStruct=true), so properties are inline
	// but $defs contains all referenced types

	// Override tasks and triggers fields with proper oneOf discriminated arrays
	if topSchema.Properties != nil {
		if prop, ok := topSchema.Properties.Get("tasks"); ok {
			prop.Type = "array"
			prop.Items = oneOfRefs(taskTypes)
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("triggers"); ok {
			prop.Type = "array"
			prop.Items = oneOfRefs(triggerTypes)
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("platform"); ok {
			prop.Type = "string"
			prop.Enum = []any{"", "concourse", "actions"}
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("feature_toggles"); ok {
			prop.Type = "array"
			prop.Items = featureToggleSchema()
			prop.Ref = ""
		}
	}

	// Ensure $defs map exists
	if topSchema.Definitions == nil {
		topSchema.Definitions = make(jsonschema.Definitions)
	}

	// Add concrete task and trigger type definitions (each with type const + additionalProperties: false)
	for _, tt := range taskTypes {
		topSchema.Definitions[tt.defKey] = reflectTaskOrTrigger(&r, tt.value, tt.typeName, topSchema.Definitions)
	}
	for _, tt := range triggerTypes {
		topSchema.Definitions[tt.defKey] = reflectTaskOrTrigger(&r, tt.value, tt.typeName, topSchema.Definitions)
	}

	// Fix sub-fields that use TaskList (pre_promote, parallel.tasks, sequence.tasks)
	deployCFKey := defKeyFor(taskTypes, "deploy-cf")
	if def, ok := topSchema.Definitions[deployCFKey]; ok && def.Properties != nil {
		if prop, ok := def.Properties.Get("pre_promote"); ok {
			prop.Type = "array"
			prop.Items = oneOfRefs(taskTypes)
			prop.Ref = ""
		}
	}
	for _, typeName := range []string{"parallel", "sequence"} {
		key := defKeyFor(taskTypes, typeName)
		if def, ok := topSchema.Definitions[key]; ok && def.Properties != nil {
			if prop, ok := def.Properties.Get("tasks"); ok {
				prop.Type = "array"
				prop.Items = oneOfRefs(taskTypes)
				prop.Ref = ""
			}
		}
	}

	// Fix compose_file: ComposeFiles is actually a string (space-separated) in JSON
	dockerComposeKey := defKeyFor(taskTypes, "docker-compose")
	if def, ok := topSchema.Definitions[dockerComposeKey]; ok && def.Properties != nil {
		if prop, ok := def.Properties.Get("compose_file"); ok {
			prop.Type = "string"
			prop.Items = nil
			prop.Ref = ""
			prop.Description = "Space-separated list of docker-compose files"
		}
	}

	// Fix Vars type: values are coerced to strings by custom UnmarshalJSON, so allow any scalar
	topSchema.Definitions["Vars"] = &jsonschema.Schema{
		Type: "object",
		AdditionalProperties: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{Type: "string"},
				{Type: "number"},
				{Type: "boolean"},
			},
		},
		Description: "Key-value pairs of environment variables (values are coerced to strings)",
	}

	// Set schema metadata
	topSchema.Version = "https://json-schema.org/draft/2020-12/schema"
	topSchema.Title = "Halfpipe Manifest"
	topSchema.Description = "Schema for the halfpipe CI/CD manifest"
	topSchema.AdditionalProperties = jsonschema.FalseSchema

	return topSchema
}

// typeDefEntry describes a single task or trigger type for schema generation.
type typeDefEntry struct {
	typeName string // value of the "type" discriminator field (e.g. "run", "deploy-cf")
	defKey   string // key used in JSON Schema $defs (e.g. "Run", "DeployCF")
	value    any    // zero-value of the Go struct to reflect
}

var taskTypes = []typeDefEntry{
	{"run", "Run", manifest.Run{}},
	{"deploy-cf", "DeployCF", manifest.DeployCF{}},
	{"deploy-katee", "DeployKatee", manifest.DeployKatee{}},
	{"docker-push", "DockerPush", manifest.DockerPush{}},
	{"docker-compose", "DockerCompose", manifest.DockerCompose{}},
	{"consumer-integration-test", "ConsumerIntegrationTest", manifest.ConsumerIntegrationTest{}},
	{"deploy-ml-zip", "DeployMLZip", manifest.DeployMLZip{}},
	{"deploy-ml-modules", "DeployMLModules", manifest.DeployMLModules{}},
	{"parallel", "Parallel", manifest.Parallel{}},
	{"sequence", "Sequence", manifest.Sequence{}},
	{"buildpack", "Buildpack", manifest.Buildpack{}},
	{"copy-container-image", "CopyContainerImage", manifest.CopyContainerImage{}},
}

var triggerTypes = []typeDefEntry{
	{"git", "GitTrigger", manifest.GitTrigger{}},
	{"timer", "TimerTrigger", manifest.TimerTrigger{}},
	{"docker", "DockerTrigger", manifest.DockerTrigger{}},
	{"pipeline", "PipelineTrigger", manifest.PipelineTrigger{}},
}

// oneOfRefs builds a oneOf schema with $ref entries for each entry in entries.
// This ensures taskOneOf/triggerOneOf stay in sync with taskTypes/triggerTypes automatically.
func oneOfRefs(entries []typeDefEntry) *jsonschema.Schema {
	refs := make([]*jsonschema.Schema, len(entries))
	for i, e := range entries {
		refs[i] = &jsonschema.Schema{Ref: "#/$defs/" + e.defKey}
	}
	return &jsonschema.Schema{OneOf: refs}
}

// defKeyFor returns the defKey for the entry with the given typeName, or panics if not found.
// This prevents silent breakage if a type name is renamed or removed.
func defKeyFor(entries []typeDefEntry, typeName string) string {
	for _, e := range entries {
		if e.typeName == typeName {
			return e.defKey
		}
	}
	panic(fmt.Sprintf("generate-schema: no entry with typeName %q", typeName))
}

// featureToggleSchema returns a schema for a single feature toggle string,
// derived from manifest.AvailableFeatureToggles so it stays in sync automatically.
func featureToggleSchema() *jsonschema.Schema {
	enums := make([]any, len(manifest.AvailableFeatureToggles))
	for i, f := range manifest.AvailableFeatureToggles {
		enums[i] = f
	}
	return &jsonschema.Schema{
		Type: "string",
		Enum: enums,
	}
}

// reflectTaskOrTrigger reflects a concrete task/trigger struct and adds a const
// discriminator on the "type" field, plus additionalProperties: false.
func reflectTaskOrTrigger(r *jsonschema.Reflector, v any, typeValue string, defs jsonschema.Definitions) *jsonschema.Schema {
	s := r.Reflect(v)

	// Merge any new $defs discovered into the parent defs map
	for k, v := range s.Definitions {
		if _, exists := defs[k]; !exists {
			defs[k] = v
		}
	}

	// Strip the nested $schema and $defs from the task def - refs all live at root level
	s.Version = ""
	s.Definitions = nil

	// s itself is the expanded struct schema; set type const and additionalProperties
	if s.Properties == nil {
		s.Properties = jsonschema.NewProperties()
	}

	// Add or override the "type" property as a const
	typeProp := &jsonschema.Schema{
		Type:        "string",
		Const:       typeValue,
		Description: s.Description,
	}
	s.Properties.Set("type", typeProp)

	// Make "type" required
	typeRequired := slices.Contains(s.Required, "type")
	if !typeRequired {
		s.Required = append([]string{"type"}, s.Required...)
	}

	s.AdditionalProperties = jsonschema.FalseSchema

	return s
}
