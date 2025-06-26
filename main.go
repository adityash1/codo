package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"codo/agent"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Warning: Could not load .env file: %v", err)
	}

	// Load AWS configuration
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		log.Fatalf("Error: %v", err)
		return
	}

	bedrockClient := bedrockruntime.NewFromConfig(sdkConfig)

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := agent.NewAgent(bedrockClient, getUserMessage)
	err = agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
