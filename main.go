package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"strings"
	"time"
)

type S3Bucket struct {
	Name		string
	Creation	time.Time
}

type Trails struct {
	Name		string
	BucketName	string
	Region		string
	MultiRegion bool
	Enabled		bool
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg + "\n", args...)
	os.Exit(1)
}

func setRegions (r string) ([]string) {

	var regions []string

	if r == "all" {
		resolver := endpoints.DefaultResolver()
		partitions := resolver.(endpoints.EnumPartitions).Partitions()

		// Ignore AWS in China and GovCloud regions
		// Future feature request
		for _, p := range partitions {
			if p.ID() == "aws-cn"{ continue }
			if p.ID() == "aws-us-gov"{ continue }
			for id, _ := range p.Regions() {
				regions = append(regions, id)
			}
		}

	} else {
		regions = append(regions, r)

	}

	return regions
}

var auditWorkers = 1
var enableEC2, enableS3, enableTrail = false, false, false

func main() {

	// Parse command-line flags
	awsRegion := flag.String("region", "us-west-2", "AWS region to scan")
	awsService := flag.String("service", "all", "AWS service to scan")
	flag.Parse()

	// Scans implemented services or select
	// Existing modules: EC2, S3, CloudTrail
	// Defaults to all
	// TODO // Implement additional services, e.g. Lambda
	switch *awsService {
	case "ec2":
		enableEC2 = true
	case "s3":
		enableS3 = true
	case "cloudtrail":
		enableTrail = true
	default:
		auditWorkers = 3
		fmt.Println("Auditing AWS Services: EC2, S3, CloudTrail")
		enableEC2, enableS3, enableTrail = true, true, true
	}

	// Scans all regions or select
	// Defaults to "us-west-2"
	var scanRegions []string
	if strings.ToLower(*awsRegion) != "all" {
		scanRegions = setRegions(*awsRegion)
	} else {
		scanRegions = setRegions("all")
	}
	fmt.Println("Auditing AWS Regions:", scanRegions)

	// Describe AWS credentials
	// Credentials file ~/.aws/credentials
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsWest2RegionID),
	})
	fmt.Println("AWS API Identity")
	// describeIdentity(sess)

	worker := make(chan bool, auditWorkers)

	// Scans EC2, CloudTail and S3 AWS services
	// Audit logic returns JSON
	if enableS3 {
		go auditS3(sess, worker)
	}
	if enableTrail {
		go auditTrails(sess, scanRegions, worker)
	}
	if enableEC2 {
		go auditEC2(sess, worker)
	}

	// TODO // REFACTOR
	switch auditWorkers {
	case 3:
		<- worker
		<- worker
		<- worker
	default:
		<- worker
	}
}
