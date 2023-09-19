package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const planWindowLayout = "2006_002" // Year_DayOfYear

type PlanDatastoreConfig struct {
	BackendConfig
	Config aws.Config
}

func NewAwsClient() aws.Config {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))

	if err != nil {
		log.Fatalf("Unable to load AWS configuration: %s", err)
	}

	return cfg
}

func (cfg PlanDatastoreConfig) Key(changeSetId string) string {
	return cfg.Environment + "/" + changeSetId
}

func savePlan(cfg PlanDatastoreConfig, changeSetId string, plan io.Reader) error {
	c := s3.NewFromConfig(cfg.Config)
	key := cfg.Key(changeSetId)

	input := &s3.PutObjectInput{
		Bucket: &cfg.Storage,
		Key:    &key,
		Body:   plan,
	}
	_, err := c.PutObject(context.TODO(), input)

	return err
}

// Save a changeSetId to the index of changeSets
func indexPlan(cfg PlanDatastoreConfig, pc *PlannedChange) error {
	c := dynamodb.NewFromConfig(cfg.Config)
	input := &dynamodb.PutItemInput{
		TableName: &cfg.Datastore,
		Item: map[string]types.AttributeValue{
			"planWindow": &types.AttributeValueMemberS{
				Value: time.Unix(pc.CreatedAtUtc, 0).Format(planWindowLayout),
			},
			"changeSetId": &types.AttributeValueMemberS{
				Value: string(pc.ChangeSetId),
			},
			"sessionId": &types.AttributeValueMemberS{
				Value: pc.SessionId,
			},
			"createdAtUtc": &types.AttributeValueMemberN{
				Value: fmt.Sprintf("%d", pc.CreatedAtUtc),
			},
		},
	}

	_, err := c.PutItem(context.TODO(), input)
	return err
}

// SavePlan: Save to s3 and index in dynamodb
func SavePlan(cfg PlanDatastoreConfig, sessionId string, plan Plan) (*PlannedChange, error) {

	b, err := json.Marshal(plan)
	if err != nil {
		log.Printf("Failed to encode plan")
		return nil, err
	}

	if err := savePlan(cfg, plan.ChangeSetId, bytes.NewBuffer(b)); err != nil {
		log.Printf("Failed to save plan: %s", err)
		return nil, err
	}
	plannedChange := &PlannedChange{
		SessionId:    sessionId,
		ChangeSetId:  ChangeSetId(plan.ChangeSetId),
		CreatedAtUtc: time.Now().UTC().Unix(),
	}

	err = indexPlan(cfg, plannedChange)
	return plannedChange, err
}
