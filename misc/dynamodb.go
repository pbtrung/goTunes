package main

import (
	"encoding/json"
	"log"

	"time"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {

	// read-only access
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("../.aws", "readDB"),
	})

	svc := sts.New(sess)
	params := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(int64(36000)),
	}
	sessionToken, err := svc.GetSessionToken(params)
	if err != nil {
		log.Fatalln(err)
	}

	type Token struct {
		AccessKeyID     string
		SecretAccessKey string
		SessionToken    string
		Expiration      *time.Time
	}

	token := &Token{AccessKeyID: *sessionToken.Credentials.AccessKeyId, SecretAccessKey: *sessionToken.Credentials.SecretAccessKey,
		SessionToken: *sessionToken.Credentials.SessionToken, Expiration: sessionToken.Credentials.Expiration}

	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(sring(b))
}
