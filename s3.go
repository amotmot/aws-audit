package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

func auditS3(sess *session.Session, done chan bool) {

	fmt.Println("Auditing S3")

	svcS3 := s3.New(sess)

	resultBuckets, err := svcS3.ListBuckets(nil)
	if err != nil {
		fmt.Printf("Unable to list buckets, %v", err)
	}

	buckets := make([]S3Bucket, len(resultBuckets.Buckets))

	for s3count, b := range resultBuckets.Buckets {
		buckets[s3count] = S3Bucket{aws.StringValue(b.Name), aws.TimeValue(b.CreationDate)}

		// TODO // Is Public?
		/*
		params := &s3.GetBucketAclInput{
			Bucket: b.Name,
		}
		resultACL, err := svcS3.GetBucketAcl(params)
		if err != nil {
			fmt.Printf("//ERROR, %v", err)
			fmt.Println(b.Name)
		}
		fmt.Println("Bucket Name:", aws.StringValue(b.Name))
		fmt.Println(resultACL)
		*/
	}

	json.NewEncoder(os.Stdout).Encode(buckets)
	done <- true
}
