/*
Analyzes terraform plan JSON for differences
*/
package analyzer

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type ChangeState map[string]interface{}

// DiffSet is a map of the field name to the differences
type UpdatedResource struct {
	Resource
	ChangeDiffs []ChangeDiff `json:"change_diffs"`
}
type CreatedResource struct {
	ResourceChange
	CreatedResource map[string]interface{} `json:"values"`
}

// TODO: Should have a DestroyedResource which provides better name information
// about what was destroyed
type ChangeDiff struct {
	Property string `json:"property"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// ResourceChange{ Change, Resource}
type Change struct {
	Actions []string
	Before  ChangeState `json:"before"`
	After   ChangeState `json:"after"`
}
type Resource struct {
	Address      string `json:"address"`
	Mode         string `json:"mode"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
}

type ResourceChange struct {
	Resource
	Change `json:"change"`
}
type ModuleResource struct {
	Resource
	Index         int
	SchemaVersion int                    `json:"schema_version"`
	Values        map[string]interface{} `json:"values"`
}
type Module struct {
	Resources []ModuleResource `json:"resources"`
}
type RootModule struct {
	Module
	ChildModules []Module `json:"child_modules"`
}
type PlannedValues struct {
	RootModule `json:"root_module"`
}
type Plan struct {
	ChangeSetId     string           `json:"changeSetId"`
	ResourceChanges []ResourceChange `json:"resource_changes"`
	PlannedValues   `json:"planned_values"`
}

// Returns true if an action is "update" for the change
func IsUpdate(c Change) bool {
	for _, action := range c.Actions {
		if action == "update" {
			return true
		}
	}
	return false
}

// Returns true if an action is "create" for the change
func IsCreate(c Change) bool {
	for _, action := range c.Actions {
		if action == "create" {
			return true
		}
	}
	return false
}

// Returns true if an action is "delete" for the change
func IsDelete(c Change) bool {
	for _, action := range c.Actions {
		if action == "delete" {
			return true
		}
	}
	return false
}

type Selector func(Change) bool

func getResources(plan Plan, selector Selector) []ResourceChange {
	var changes []ResourceChange
	for _, rc := range plan.ResourceChanges {
		if selector(rc.Change) {
			changes = append(changes, rc)
		}
	}

	return changes
}

// Get a list of created Resources
func GetCreatedResources(plan Plan) []ResourceChange {
	return getResources(plan, IsCreate)
}

// Get a list of updated Resources
func GetUpdatedResources(plan Plan) []ResourceChange {
	return getResources(plan, IsUpdate)
}

// Get a list of deleted Resources
func GetDeletedResources(plan Plan) []ResourceChange {
	return getResources(plan, IsDelete)
}

func NewDiff(property string, from, to interface{}) (*ChangeDiff, error) {
	if s, ok := from.(string); ok {
		return &ChangeDiff{
			Property: property,
			From:     s,
			To:       to.(string),
		}, nil
	}

	fromB, err := json.Marshal(from)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 'from': %s", err)
	}
	toB, err := json.Marshal(to)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 'to': %s", err)
	}
	return &ChangeDiff{
		property,
		string(fromB),
		string(toB),
	}, nil
}

func (s ChangeState) Diff(other ChangeState) []ChangeDiff {
	var diffs []ChangeDiff

	for field, value := range s {
		// No changes
		if reflect.DeepEqual(value, other[field]) {
			continue
		}
		diff, err := NewDiff(field, value, other[field])
		if err != nil {
			log.Println(err)
			continue
		}
		diffs = append(diffs, *diff)
	}
	return diffs
}
func Resources(changes []ResourceChange) (resources []Resource) {
	for _, change := range changes {
		resources = append(resources, change.Resource)
	}
	return
}
func Diffs(changes []ResourceChange) (diffs []UpdatedResource) {
	for _, change := range changes {
		diffs = append(diffs, UpdatedResource{
			Resource:    change.Resource,
			ChangeDiffs: change.Before.Diff(change.After),
		})
	}
	return
}

func CreatedResources(plan Plan) []CreatedResource {
	var createdResouces []CreatedResource
	for _, resourceChange := range GetCreatedResources(plan) {
		createdResource := CreatedResource{
			ResourceChange: resourceChange,
			// CreatedResource: plan.PlannedValues.
		}
		if cr, ok := selectCreatedResource(plan, resourceChange.Address); ok {
			createdResource.CreatedResource = cr.Values
		}
		createdResouces = append(createdResouces, createdResource)
	}
	return createdResouces
}

func selectCreatedResource(plan Plan, address string) (ModuleResource, bool) {

	for _, r := range plan.RootModule.Resources {

		if strings.EqualFold(r.Address, address) {
			return r, true
		}
	}
	for _, child := range plan.RootModule.ChildModules {
		for _, r := range child.Resources {

			// case insensitive
			if strings.EqualFold(r.Address, address) {
				return r, true
			}
		}
	}
	log.Printf("No resource matching %s", address)
	return ModuleResource{}, false
}
