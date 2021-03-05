package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/hubinix/gameplay/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"sync"
)

var log = logger.NewLogger("frontier-meta")

type Empty struct{}

type DynamodbManager struct {
	sync.RWMutex
	awsEndpoint string
}

var dynamodbManagerOnce sync.Once
var dynamodbManagerInstance *DynamodbManager

func GetDynamodbManager() *DynamodbManager {
	dynamodbManagerOnce.Do(func() {
		dynamodbManagerInstance = &DynamodbManager{
			awsEndpoint: "http://localhost:8000",
		}

	})
	return dynamodbManagerInstance
}

// GetTableInfo retrieves information about the table.
func GetTableInfo(c context.Context, api *dynamodb.Client, input *dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error) {
	return api.ListTables(c, input)
}
func (dm *DynamodbManager) Start(ctx context.Context) {

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID && region == "test" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:8001",
				SigningRegion: "us-west-2",
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion("test"),
		config.WithEndpointResolver(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")))
	client := dynamodb.NewFromConfig(cfg)
	// Build the input parameters for the request.

	input := &dynamodb.ListTablesInput{}

	_, err := GetTableInfo(context.TODO(), client, input)
	if err != nil {
		panic("failed to describe table, " + err.Error())
	}

}
