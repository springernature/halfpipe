package main

import (
	"encoding/json"
	"fmt"
	"os"

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

	fmt.Println(string(out))
}

func buildSchema() *jsonschema.Schema {
	r := &jsonschema.Reflector{
		DoNotReference:             false,
		Anonymous:                  true,
		ExpandedStruct:             true,
		RequiredFromJSONSchemaTags: true, // only fields tagged `jsonschema:"required"` are required
	}

	// Reflect the top-level manifest to get all $defs populated
	topSchema := r.Reflect(manifest.Manifest{})

	// topSchema is already expanded (ExpandedStruct=true), so properties are inline
	// but $defs contains all referenced types

	// Override tasks and triggers fields with proper oneOf discriminated arrays
	if topSchema.Properties != nil {
		if prop, ok := topSchema.Properties.Get("tasks"); ok {
			prop.Type = "array"
			prop.Items = taskOneOf()
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("triggers"); ok {
			prop.Type = "array"
			prop.Items = triggerOneOf()
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("platform"); ok {
			prop.Type = "string"
			prop.Enum = []interface{}{"", "concourse", "actions"}
			prop.Ref = ""
		}
		if prop, ok := topSchema.Properties.Get("feature_toggles"); ok {
			prop.Type = "array"
			prop.Items = &jsonschema.Schema{
				Type: "string",
				Enum: []interface{}{
					manifest.FeatureUpdatePipeline,
					manifest.FeatureUpdatePipelineAndTag,
					manifest.FeatureGithubStatuses,
					manifest.FeatureGhas,
				},
			}
			prop.Ref = ""
		}
	}

	// Ensure $defs map exists
	if topSchema.Definitions == nil {
		topSchema.Definitions = make(jsonschema.Definitions)
	}

	// Add concrete task type definitions (each with type const + additionalProperties: false)
	for _, tt := range taskTypes {
		def := reflectTaskOrTrigger(r, tt.value, tt.typeName, topSchema.Definitions)
		topSchema.Definitions[tt.defKey] = def
	}
	// Add concrete trigger type definitions
	for _, tt := range triggerTypes {
		def := reflectTaskOrTrigger(r, tt.value, tt.typeName, topSchema.Definitions)
		topSchema.Definitions[tt.defKey] = def
	}

	// Fix sub-fields that use TaskList (pre_promote, parallel.tasks, sequence.tasks)
	for _, key := range []string{"DeployCF"} {
		if def, ok := topSchema.Definitions[key]; ok && def.Properties != nil {
			if prop, ok := def.Properties.Get("pre_promote"); ok {
				prop.Type = "array"
				prop.Items = taskOneOf()
				prop.Ref = ""
			}
		}
	}
	for _, key := range []string{"Parallel", "Sequence"} {
		if def, ok := topSchema.Definitions[key]; ok && def.Properties != nil {
			if prop, ok := def.Properties.Get("tasks"); ok {
				prop.Type = "array"
				prop.Items = taskOneOf()
				prop.Ref = ""
			}
		}
	}

	// Fix compose_file: ComposeFiles is actually a string (space-separated) in JSON
	if def, ok := topSchema.Definitions["DockerCompose"]; ok && def.Properties != nil {
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
	topSchema.Description = "Schema for the halfpipe CI/CD manifest (.halfpipe.io, .halfpipe.io.yml, .halfpipe.io.yaml)"
	topSchema.AdditionalProperties = jsonschema.FalseSchema

	return topSchema
}

type taskTypeDef struct {
	typeName string
	defKey   string
	value    interface{}
}

var taskTypes = []taskTypeDef{
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

var triggerTypes = []taskTypeDef{
	{"git", "GitTrigger", manifest.GitTrigger{}},
	{"timer", "TimerTrigger", manifest.TimerTrigger{}},
	{"docker", "DockerTrigger", manifest.DockerTrigger{}},
	{"pipeline", "PipelineTrigger", manifest.PipelineTrigger{}},
}

func taskOneOf() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Ref: "#/$defs/Run"},
			{Ref: "#/$defs/DeployCF"},
			{Ref: "#/$defs/DeployKatee"},
			{Ref: "#/$defs/DockerPush"},
			{Ref: "#/$defs/DockerCompose"},
			{Ref: "#/$defs/ConsumerIntegrationTest"},
			{Ref: "#/$defs/DeployMLZip"},
			{Ref: "#/$defs/DeployMLModules"},
			{Ref: "#/$defs/Parallel"},
			{Ref: "#/$defs/Sequence"},
			{Ref: "#/$defs/Buildpack"},
			{Ref: "#/$defs/CopyContainerImage"},
		},
	}
}

func triggerOneOf() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Ref: "#/$defs/GitTrigger"},
			{Ref: "#/$defs/TimerTrigger"},
			{Ref: "#/$defs/DockerTrigger"},
			{Ref: "#/$defs/PipelineTrigger"},
		},
	}
}

// reflectTaskOrTrigger reflects a concrete task/trigger struct and adds a const
// discriminator on the "type" field, plus additionalProperties: false.
func reflectTaskOrTrigger(r *jsonschema.Reflector, v interface{}, typeValue string, defs jsonschema.Definitions) *jsonschema.Schema {
	inner := &jsonschema.Reflector{
		DoNotReference:             false,
		Anonymous:                  true,
		ExpandedStruct:             true,
		RequiredFromJSONSchemaTags: true,
	}
	s := inner.Reflect(v)

	// Merge any new $defs discovered into the parent defs map
	if s.Definitions != nil {
		for k, v := range s.Definitions {
			if _, exists := defs[k]; !exists {
				defs[k] = v
			}
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
		Description: fmt.Sprintf("Must be %q", typeValue),
	}
	s.Properties.Set("type", typeProp)

	// Make "type" required
	typeRequired := false
	for _, req := range s.Required {
		if req == "type" {
			typeRequired = true
			break
		}
	}
	if !typeRequired {
		s.Required = append([]string{"type"}, s.Required...)
	}

	s.AdditionalProperties = jsonschema.FalseSchema

	return s
}
