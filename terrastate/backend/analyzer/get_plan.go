package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	DONE = iota
)

type ChangeSetId string
type PlannedChange struct {
	SessionId    string `json:"session_id", dynamodbav:"sessionId"`
	ChangeSetId  `json:"change_set_id" dynamodbav:"changeSetId"`
	CreatedAtUtc int64  `json:"created_at", dynamodbav:"createdAtUtc"`
	PlanWindow   string `json:-, dynamodbav:"planWindow"`
}

func (c ChangeSetId) String() string {
	return string(c)
}

func getPlanWindow(cfg PlanDatastoreConfig, planWindow string) []PlannedChange {
	var plannedChanges []PlannedChange
	c := dynamodb.NewFromConfig(cfg.Config)

	keyConditionExpression := fmt.Sprintf("planWindow = :planWindow")
	expressionAttributeValues := map[string]types.AttributeValue{
		":planWindow": &types.AttributeValueMemberS{
			Value: planWindow,
		},
	}
	log.Printf("Querying plan window %v", planWindow)
	input := &dynamodb.QueryInput{
		TableName:                 &cfg.Datastore,
		KeyConditionExpression:    &keyConditionExpression,
		ExpressionAttributeValues: expressionAttributeValues,
	}

	output, err := c.Query(context.TODO(), input)
	if err != nil {
		log.Printf("Unable to get planned changes in window %s: %s", planWindow, err)
		return plannedChanges
	}
	err = attributevalue.UnmarshalListOfMaps(output.Items, &plannedChanges)
	if err != nil {
		log.Printf("Unable to unmarshal dynamodb results: %s", err)
	}
	log.Printf("Found plannedChanges %#v", plannedChanges)
	return plannedChanges
}

// ListPlans Provides a lit of plans which occur between a past and present moment
func ListPlans(cfg PlanDatastoreConfig, present, past time.Time) ([]PlannedChange, error) {

	var plannedChanges []PlannedChange
	if past.After(present) {
		return plannedChanges, fmt.Errorf("error: The past must come before the present")
	}

	timeWindow := present.Sub(past).Abs()
	planWindows := map[string]int8{}
	planWindows[present.Format(planWindowLayout)] = 1

	for hours := int64(0); hours < int64(math.Ceil(timeWindow.Hours())); hours++ {
		deltaT := time.Duration(-1 * hours)
		planWindows[present.Add(deltaT*time.Hour).Format(planWindowLayout)] = 1
	}

	wg := new(sync.WaitGroup)
	pcChan := make(chan PlannedChange, 1)
	windowPlanDone := make(chan int, 1)
	go func() {
		for _ = range windowPlanDone {
			log.Printf("Closing plan window")
			wg.Done()
		}
	}()
	go func() {
		for pc := range pcChan {
			log.Printf("Received planned changes")
			plannedChanges = append(plannedChanges, pc)
		}
	}()
	for planWindow, _ := range planWindows {
		log.Printf("Starting to plan window")
		log.Printf("Fetching plans for %s", planWindow)
		wg.Add(1)
		go func(pw string) {
			plans := getPlanWindow(cfg, pw)
			for _, pc := range plans {
				log.Printf("Resolved plan window %v", pw)
				pcChan <- pc
			}
			windowPlanDone <- DONE
		}(planWindow)
	}
	wg.Wait()
	log.Printf("Found %d planned changes", len(plannedChanges))
	return plannedChanges, nil
}

func GetPlan(cfg PlanDatastoreConfig, id ChangeSetId) (*Plan, error) {
	plan := &Plan{}

	c := s3.NewFromConfig(cfg.Config)
	key := cfg.Key(string(id))
	input := &s3.GetObjectInput{
		Bucket: &cfg.Storage,
		Key:    &key,
	}

	output, err := c.GetObject(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to find s3 changeset %s", id)
		return nil, err
	}

	err = json.NewDecoder(output.Body).Decode(plan)
	return plan, err
}
