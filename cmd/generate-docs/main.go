package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Minimal JSON Schema representation for what we need.
type Schema struct {
	Defs       map[string]*SchemaDef `json:"$defs"`
	Properties map[string]*Property  `json:"properties"`
	Title      string                `json:"title"`
	Desc       string                `json:"description"`
}

type SchemaDef struct {
	Properties           orderedProps    `json:"properties"`
	Required             []string        `json:"required"`
	Type                 any             `json:"type"`
	Desc                 string          `json:"description"`
	AdditionalProperties any             `json:"additionalProperties"`
	Items                json.RawMessage `json:"items"`
}

type Property struct {
	Type               any         `json:"type"`
	Desc               string      `json:"description"`
	Ref                string      `json:"$ref"`
	Const              any         `json:"const"`
	Enum               []any       `json:"enum"`
	Deprecated         bool        `json:"deprecated"`
	DeprecationMessage string      `json:"deprecationMessage"`
	Items              *Property   `json:"items"`
	OneOf              []*Property `json:"oneOf"`
}

// orderedProps preserves JSON key order.
type orderedProps struct {
	keys   []string
	values map[string]*Property
}

func (o *orderedProps) UnmarshalJSON(data []byte) error {
	o.values = make(map[string]*Property)
	dec := json.NewDecoder(strings.NewReader(string(data)))
	t, err := dec.Token() // opening {
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected {, got %v", t)
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key := keyTok.(string)
		var val Property
		if err := dec.Decode(&val); err != nil {
			return err
		}
		o.keys = append(o.keys, key)
		o.values[key] = &val
	}
	return nil
}

// anchorMap maps $defs keys (e.g. "Notifications", "Run") to their markdown anchor slugs.
var anchorMap map[string]string

// buildAnchorMap pre-computes the mapping from schema $defs names to the heading
// anchors we will emit, so all internal links are consistent.
func buildAnchorMap(schema *Schema) {
	anchorMap = make(map[string]string)

	// Triggers: heading is ### `<typeName>` (trigger)
	for _, name := range sortedDefKeys(schema.Defs, "Trigger") {
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		// GitHub slug for ### `docker` (trigger) is "docker-trigger"
		anchorMap[name] = typeName + "-trigger"
	}

	// Tasks: heading is ### `<typeName>`
	for _, name := range sortedDefKeys(schema.Defs, "") {
		if strings.Contains(name, "Trigger") || isHelperDef(name) {
			continue
		}
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		// GitHub slug for ### `run` is "run"
		anchorMap[name] = typeName
	}

	// Supporting types: plain ### headings
	anchorMap["Notifications"] = "notifications"
	anchorMap["NotificationChannel"] = "notificationchannel"
	anchorMap["NotificationChannels"] = "notificationchannel" // alias — array of NotificationChannel
	anchorMap["Docker"] = "docker"
	anchorMap["GitHubEnvironment"] = "githubenvironment"
	anchorMap["Vars"] = "vars"
}

func refLink(defName string) string {
	anchor, ok := anchorMap[defName]
	if !ok {
		return defName
	}
	return fmt.Sprintf("[%s](#%s)", defName, anchor)
}

func main() {
	root := findModuleRoot()
	data, err := os.ReadFile(filepath.Join(root, "schema.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading schema.json: %v\n", err)
		os.Exit(1)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing schema.json: %v\n", err)
		os.Exit(1)
	}

	buildAnchorMap(&schema)

	var b strings.Builder

	b.WriteString("# Halfpipe Manifest Reference\n\n")

	// Table of contents
	b.WriteString("## Contents\n\n")
	b.WriteString("- [Top-Level Fields](#top-level-fields)\n")
	b.WriteString("- [Triggers](#triggers)\n")
	for _, name := range sortedDefKeys(schema.Defs, "Trigger") {
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("  - [`%s`](#%s)\n", typeName, anchorMap[name]))
	}
	b.WriteString("- [Tasks](#tasks)\n")
	for _, name := range sortedDefKeys(schema.Defs, "") {
		if strings.Contains(name, "Trigger") || isHelperDef(name) {
			continue
		}
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("  - [`%s`](#%s)\n", typeName, anchorMap[name]))
	}
	b.WriteString("- [Supporting Types](#supporting-types)\n")
	b.WriteString("  - [Notifications](#notifications)\n")
	b.WriteString("  - [NotificationChannel](#notificationchannel)\n")
	b.WriteString("  - [Vars](#vars)\n")
	b.WriteString("  - [Docker](#docker)\n")
	b.WriteString("  - [GitHubEnvironment](#githubenvironment)\n")
	b.WriteString("  - [Feature Toggles](#feature-toggles)\n")
	b.WriteString("\n")

	// Top-level fields
	b.WriteString("## Top-Level Fields\n\n")
	writePropsTable(&b, topLevelOrdered(schema.Properties), nil, nil, &schema)

	// Triggers
	b.WriteString("\n## Triggers\n\n")
	b.WriteString("Triggers cause the pipeline to run. Specified under the `triggers` key.\n\n")
	for _, name := range sortedDefKeys(schema.Defs, "Trigger") {
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("### `%s` (trigger)\n\n", typeName))
		writeDefTable(&b, def, &schema)
		writeExample(&b, root, "trigger-"+typeName)
	}

	// Tasks
	b.WriteString("\n## Tasks\n\n")
	b.WriteString("Tasks define the steps in your pipeline. Specified under the `tasks` key.\n\n")
	for _, name := range sortedDefKeys(schema.Defs, "") {
		if strings.Contains(name, "Trigger") || isHelperDef(name) {
			continue
		}
		def := schema.Defs[name]
		typeName := typeConstFromDef(def)
		if typeName == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("### `%s`\n\n", typeName))
		writeDefTable(&b, def, &schema)
		writeExample(&b, root, "task-"+typeName)
	}

	// Supporting types
	b.WriteString("\n## Supporting Types\n\n")

	if def, ok := schema.Defs["Notifications"]; ok {
		b.WriteString("### Notifications\n\n")
		writeDefTable(&b, def, &schema)
	}
	if def, ok := schema.Defs["NotificationChannel"]; ok {
		b.WriteString("### NotificationChannel\n\n")
		writeDefTable(&b, def, &schema)
	}
	// Vars
	b.WriteString("### Vars\n\n")
	b.WriteString("Key-value pairs of environment variables. Values are coerced to strings.\n\n")
	b.WriteString("```yaml\nvars:\n  FOO: bar\n  PORT: 8080\n  DEBUG: true\n```\n\n")

	if def, ok := schema.Defs["Docker"]; ok {
		b.WriteString("### Docker\n\n")
		b.WriteString("Docker image configuration used by the [`run`](#run) task.\n\n")
		writeDefTable(&b, def, &schema)
	}
	if def, ok := schema.Defs["GitHubEnvironment"]; ok {
		b.WriteString("### GitHubEnvironment\n\n")
		writeDefTable(&b, def, &schema)
	}

	// Feature toggles
	b.WriteString("### Feature Toggles\n\n")
	b.WriteString("Available values for the `feature_toggles` array:\n\n")
	if prop, ok := schema.Properties["feature_toggles"]; ok && prop.Items != nil {
		for _, e := range prop.Items.Enum {
			b.WriteString(fmt.Sprintf("- `%v`\n", e))
		}
	}
	b.WriteString("\n")

	fmt.Print(b.String())
}

func writeDefTable(b *strings.Builder, def *SchemaDef, schema *Schema) {
	if def.Desc != "" {
		b.WriteString(def.Desc + "\n\n")
	}
	if len(def.Properties.keys) == 0 {
		return
	}
	writePropsTable(b, def.Properties.keys, def.Properties.values, def.Required, schema)
}

func writeExample(b *strings.Builder, root, name string) {
	examplePath := filepath.Join(root, "docs", "examples", name+".yaml")
	data, err := os.ReadFile(examplePath)
	if err != nil {
		return
	}
	parts := strings.Split(string(data), "\n---\n")
	if len(parts) == 1 {
		b.WriteString("**Example:**\n\n")
	} else {
		b.WriteString("**Examples:**\n\n")
	}
	for _, part := range parts {
		b.WriteString("```yaml\n")
		b.WriteString(strings.TrimRight(part, "\n"))
		b.WriteString("\n```\n\n")
	}
}

func writePropsTable(b *strings.Builder, keys []string, values map[string]*Property, required []string, schema *Schema) {
	requiredSet := make(map[string]bool)
	for _, r := range required {
		requiredSet[r] = true
	}

	getVal := func(key string) *Property {
		if values != nil {
			return values[key]
		}
		if schema != nil {
			return schema.Properties[key]
		}
		return nil
	}

	b.WriteString("| Field | Type | Required | Description |\n")
	b.WriteString("|-------|------|----------|-------------|\n")

	for _, key := range keys {
		prop := getVal(key)
		if prop == nil {
			continue
		}
		// Skip the "type" discriminator field
		if key == "type" && prop.Const != nil {
			continue
		}

		typStr := resolveType(prop, schema)
		desc := prop.Desc
		if prop.Deprecated {
			desc = "⚠️ " + desc
		}
		desc = strings.ReplaceAll(desc, "|", "\\|")

		req := "optional"
		if requiredSet[key] {
			req = "required"
		}

		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", key, typStr, req, desc))
	}
	b.WriteString("\n")
}

func resolveType(prop *Property, schema *Schema) string {
	if prop.Ref != "" {
		name := strings.TrimPrefix(prop.Ref, "#/$defs/")
		return refLink(name)
	}
	if prop.Enum != nil {
		var vals []string
		for _, v := range prop.Enum {
			s := fmt.Sprintf("%v", v)
			if s == "" {
				continue
			}
			vals = append(vals, fmt.Sprintf("`%v`", v))
		}
		return strings.Join(vals, ", ")
	}

	typeStr := typeString(prop.Type)

	if typeStr == "array" && prop.Items != nil {
		if prop.Items.OneOf != nil {
			for _, o := range prop.Items.OneOf {
				if o.Ref != "" && strings.Contains(o.Ref, "Trigger") {
					return "[Trigger](#triggers)[]"
				}
			}
			return "[Task](#tasks)[]"
		}
		if prop.Items.Ref != "" {
			name := strings.TrimPrefix(prop.Items.Ref, "#/$defs/")
			return refLink(name) + "[]"
		}
		inner := typeString(prop.Items.Type)
		return inner + "[]"
	}

	return typeStr
}

func typeString(t any) string {
	switch v := t.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func typeConstFromDef(def *SchemaDef) string {
	if p, ok := def.Properties.values["type"]; ok && p.Const != nil {
		return fmt.Sprintf("%v", p.Const)
	}
	return ""
}

func sortedDefKeys(defs map[string]*SchemaDef, contains string) []string {
	var keys []string
	for k := range defs {
		if contains != "" && strings.Contains(k, contains) {
			keys = append(keys, k)
		} else if contains == "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func isHelperDef(name string) bool {
	helpers := map[string]bool{
		"Vars": true, "Docker": true, "Notifications": true,
		"NotificationChannel": true, "NotificationChannels": true,
		"GitHubEnvironment": true, "FeatureToggles": true,
		"TaskList": true, "TriggerList": true, "ComposeFiles": true,
	}
	return helpers[name]
}

func topLevelOrdered(props map[string]*Property) []string {
	order := []string{"team", "pipeline", "platform", "triggers", "tasks", "notifications", "feature_toggles", "teams_webhook"}
	var result []string
	seen := map[string]bool{}
	for _, k := range order {
		if _, ok := props[k]; ok {
			result = append(result, k)
			seen[k] = true
		}
	}
	var rest []string
	for k := range props {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(result, rest...)
}

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fmt.Fprintf(os.Stderr, "error: go.mod not found\n")
			os.Exit(1)
		}
		dir = parent
	}
}
