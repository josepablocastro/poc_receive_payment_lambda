package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/josepablocastro/poc_messages/messages"
	"github.com/josepablocastro/poc_remittance"
)

var application *poc_remittance.Application

type RDSSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DBHost   string `json:"db_host"`
	DBName   string `json:"db_name"`
	Port     string `json:"port"`
}

func init() {

	region := getEnvironmentValue("AWS_REGION")
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	log.Print("after load config")

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := secretsmanager.NewFromConfig(config)
	data_source_secret := getEnvironmentValue("DATA_SOURCE_SECRET")

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(data_source_secret),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	log.Printf("AA %v", data_source_secret)

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}
	log.Print("BB")

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	secret := RDSSecret{}
	json.Unmarshal([]byte(secretString), &secret)
	log.Print("before DB")

	dataSourceURL := fmt.Sprintf("postgresql://%s@%s/%s", url.UserPassword(secret.Username, secret.Password).String(), secret.DBHost, secret.DBName)

	dbAdapter, err := poc_remittance.NewDBAdapter(dataSourceURL)
	application = poc_remittance.NewApplication(dbAdapter)
}

func handler(ctx context.Context, event json.RawMessage) (any, error) {
	request := messages.ReceivePaymentRequest{}

	err := json.Unmarshal(event, &request)
	if err != nil {
		log.Fatalf("unable to unmarshal request, %v", err)
	}

	log.Printf("REQ %+v", request)

	payment, err := application.ReceivePayment(request.Number, request.Sender, request.Beneficiary, request.Amount, request.Currency)
	log.Printf("RES %+v", payment)

	return payment, nil
}

func main() {
	lambda.Start(handler)
}

func getEnvironmentValue(key string) string {
	if os.Getenv(key) == "" {
		log.Fatalf("%s environment variable is missing.", key)
	}
	return os.Getenv(key)
}
