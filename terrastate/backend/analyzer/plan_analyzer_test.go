package analyzer

import (
	"encoding/json"
	"os"
	"testing"
)

func loadPlan(f string) Plan {
	fh, _ := os.Open(f)
	d := json.NewDecoder(fh)
	plan := &Plan{}
	d.Decode(plan)
	return *plan
}

// NOTE: How are updates different from changes?
func TestPlanWithUpdates(t *testing.T) {
	plan := loadPlan("fixture.planWithUpdates.json")
	changes := GetUpdatedResources(plan)
	if l := len(changes); l != 1 {
		t.Fatalf("Expected 1 change, found %d", l)
	}
}

func TestPlanWithDestroyedResources(t *testing.T) {
	plan := loadPlan("fixture.planWithDestroys.json")
	deletes := GetDeletedResources(plan)
	if l := len(deletes); l < 1 {
		t.Fatalf("Expected at least 1 delete, found %d", l)
	}
}

func TestAllResourcesHaveANameField(t *testing.T) {
	plan := loadPlan("fixture.planWithDestroys.json")
	deletes := GetDeletedResources(plan)
	t.Logf("0: %#v", deletes[0])
	if name := deletes[0].Name; name != "authorizer" {
		t.Fatalf("Expected 'authorizer', got %s", name)
	}

	plan = loadPlan("fixture.planWithUpdates.json")
	changes := GetUpdatedResources(plan)
	t.Logf("0: %#v", changes[0])
	if name := changes[0].Name; name != "dev" {
		t.Fatalf("Expected 'dev', got %s", name)
	}
}

func TestAllResourcesHaveATypeField(t *testing.T) {
	plan := loadPlan("fixture.planWithDestroys.json")
	deletes := GetDeletedResources(plan)
	t.Logf("0: %#v", deletes[0])
	if typeName := deletes[0].Type; typeName != "aws_api_gateway_authorizer" {
		t.Fatalf("Expected 'aws_api_gateway_authorizer', got %s", typeName)
	}

	plan = loadPlan("fixture.planWithUpdates.json")
	changes := GetUpdatedResources(plan)
	t.Logf("0: %#v", changes[0])
	if typeName := changes[0].Type; typeName != "aws_api_gateway_stage" {
		t.Fatalf("Expected 'aws_api_gateway_stage', got %s", typeName)
	}
}

func TestPlanWithChangedResources(t *testing.T) {
	plan := loadPlan("fixture.planWithChanges.json")
	changes := GetUpdatedResources(plan)
	change := changes[0].Change
	if !IsUpdate(change) {
		t.Fatalf("Expected an update")
	}

	diffSet := change.Before.Diff(change.After)
	if l := len(diffSet); l < 1 {
		t.Errorf("Expected at changes for fields, found none")
	}
	for field, diff := range diffSet {
		t.Logf("Changed %v from %v to %v", field, diff.From, diff.To)
	}
}
