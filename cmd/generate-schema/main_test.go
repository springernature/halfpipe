package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/springernature/halfpipe/manifest"
)

func TestBuildSchema_ValidJSON(t *testing.T) {
	schema := buildSchema()
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("schema not serialisable to JSON: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("schema output is empty")
	}
}

func TestBuildSchema_SchemaVersion(t *testing.T) {
	schema := buildSchema()
	if schema.Version != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("unexpected $schema: %q", schema.Version)
	}
}

func TestBuildSchema_AllTaskDefsPresent(t *testing.T) {
	schema := buildSchema()
	for _, tt := range taskTypes {
		if _, ok := schema.Definitions[tt.defKey]; !ok {
			t.Errorf("$defs missing task type %q (defKey %q)", tt.typeName, tt.defKey)
		}
	}
}

func TestBuildSchema_AllTriggerDefsPresent(t *testing.T) {
	schema := buildSchema()
	for _, tt := range triggerTypes {
		if _, ok := schema.Definitions[tt.defKey]; !ok {
			t.Errorf("$defs missing trigger type %q (defKey %q)", tt.typeName, tt.defKey)
		}
	}
}

func TestBuildSchema_OneOfRefsMatchDefs(t *testing.T) {
	schema := buildSchema()

	tasksProp, ok := schema.Properties.Get("tasks")
	if !ok {
		t.Fatal("tasks property not found")
	}
	if tasksProp.Items == nil || tasksProp.Items.OneOf == nil {
		t.Fatal("tasks.items.oneOf not set")
	}
	for _, ref := range tasksProp.Items.OneOf {
		key := strings.TrimPrefix(ref.Ref, "#/$defs/")
		if _, ok := schema.Definitions[key]; !ok {
			t.Errorf("tasks oneOf ref %q has no matching $defs entry", ref.Ref)
		}
	}

	triggersProp, ok := schema.Properties.Get("triggers")
	if !ok {
		t.Fatal("triggers property not found")
	}
	if triggersProp.Items == nil || triggersProp.Items.OneOf == nil {
		t.Fatal("triggers.items.oneOf not set")
	}
	for _, ref := range triggersProp.Items.OneOf {
		key := strings.TrimPrefix(ref.Ref, "#/$defs/")
		if _, ok := schema.Definitions[key]; !ok {
			t.Errorf("triggers oneOf ref %q has no matching $defs entry", ref.Ref)
		}
	}
}

func TestBuildSchema_OneOfCountMatchesTypeLists(t *testing.T) {
	schema := buildSchema()

	tasksProp, _ := schema.Properties.Get("tasks")
	if got, want := len(tasksProp.Items.OneOf), len(taskTypes); got != want {
		t.Errorf("tasks oneOf has %d entries, taskTypes has %d", got, want)
	}

	triggersProp, _ := schema.Properties.Get("triggers")
	if got, want := len(triggersProp.Items.OneOf), len(triggerTypes); got != want {
		t.Errorf("triggers oneOf has %d entries, triggerTypes has %d", got, want)
	}
}

func TestBuildSchema_EachDefHasTypeConst(t *testing.T) {
	schema := buildSchema()

	for _, tt := range taskTypes {
		def, ok := schema.Definitions[tt.defKey]
		if !ok {
			continue // caught by TestBuildSchema_AllTaskDefsPresent
		}
		if def.Properties == nil {
			t.Errorf("task def %q has no properties", tt.defKey)
			continue
		}
		typeProp, ok := def.Properties.Get("type")
		if !ok {
			t.Errorf("task def %q missing 'type' property", tt.defKey)
			continue
		}
		if typeProp.Const != tt.typeName {
			t.Errorf("task def %q: type const = %q, want %q", tt.defKey, typeProp.Const, tt.typeName)
		}
	}

	for _, tt := range triggerTypes {
		def, ok := schema.Definitions[tt.defKey]
		if !ok {
			continue
		}
		if def.Properties == nil {
			t.Errorf("trigger def %q has no properties", tt.defKey)
			continue
		}
		typeProp, ok := def.Properties.Get("type")
		if !ok {
			t.Errorf("trigger def %q missing 'type' property", tt.defKey)
			continue
		}
		if typeProp.Const != tt.typeName {
			t.Errorf("trigger def %q: type const = %q, want %q", tt.defKey, typeProp.Const, tt.typeName)
		}
	}
}

func TestBuildSchema_EachDefHasTypeRequired(t *testing.T) {
	schema := buildSchema()

	checkRequired := func(defKey string) {
		def, ok := schema.Definitions[defKey]
		if !ok {
			return
		}
		for _, req := range def.Required {
			if req == "type" {
				return
			}
		}
		t.Errorf("def %q: 'type' not in required list", defKey)
	}

	for _, tt := range taskTypes {
		checkRequired(tt.defKey)
	}
	for _, tt := range triggerTypes {
		checkRequired(tt.defKey)
	}
}

func TestBuildSchema_FeatureTogglesMatchAvailable(t *testing.T) {
	schema := buildSchema()

	prop, ok := schema.Properties.Get("feature_toggles")
	if !ok {
		t.Fatal("feature_toggles property not found")
	}
	if prop.Items == nil {
		t.Fatal("feature_toggles items not set")
	}

	if got, want := len(prop.Items.Enum), len(manifest.AvailableFeatureToggles); got != want {
		t.Errorf("feature_toggles enum has %d entries, AvailableFeatureToggles has %d", got, want)
	}

	for _, ft := range manifest.AvailableFeatureToggles {
		found := false
		for _, e := range prop.Items.Enum {
			if e == ft {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("feature toggle %q missing from schema enum", ft)
		}
	}
}

func TestBuildSchema_AdditionalPropertiesFalse(t *testing.T) {
	schema := buildSchema()
	if schema.AdditionalProperties == nil {
		t.Error("root schema additionalProperties should be set (false schema)")
	}
}

func TestDefKeyFor_Found(t *testing.T) {
	key := defKeyFor(taskTypes, "run")
	if key != "Run" {
		t.Errorf("defKeyFor run = %q, want Run", key)
	}
}

func TestDefKeyFor_NotFoundPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown typeName, got none")
		}
	}()
	defKeyFor(taskTypes, "nonexistent-type")
}
