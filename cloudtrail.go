package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"os"
)

var multiRegionEnabled = false

func describeTrails(sess *session.Session, region string, done chan bool) {

	svcTrail := cloudtrail.New(sess)

	resultTrails, err := svcTrail.DescribeTrails(&cloudtrail.DescribeTrailsInput{TrailNameList: nil})
	if err != nil {
		fmt.Println("Unable to list trails", err)
	}

	// Iterate resultTrails
	var trails= make([]Trails, len(resultTrails.TrailList))
	for tcount, trail := range resultTrails.TrailList {
		if *trail.HomeRegion != *sess.Config.Region {
			break
		}
		// Call GetTrailStatus
		respTrailStatus, err := svcTrail.GetTrailStatus(&cloudtrail.GetTrailStatusInput{Name: aws.String(*trail.Name)})
		if err != nil {
			exitErrorf("Unable to get trails status", err)
		}

		if *trail.S3BucketName != "" {
			trails[tcount] = Trails{*trail.Name, *trail.S3BucketName, *sess.Config.Region, *trail.IsMultiRegionTrail, *respTrailStatus.IsLogging}

			// Does a MultiRegion trail exist?
			if *trail.IsMultiRegionTrail && *respTrailStatus.IsLogging == true {
				multiRegionEnabled = true
			}
		}
	}
	if len(trails) > 1 {
		json.NewEncoder(os.Stdout).Encode(trails)
	}

	done <- true
}

func auditTrails(sess *session.Session, regions []string, done chan bool ) {

	fmt.Println("Auditing CloudTrails")

	// Call each region async
	worker := make(chan bool, len(regions))

	// Get Trails for each region
	// Create new AWS session with a different region
	for _, region := range regions {
		copySess := sess.Copy(&aws.Config{Region: aws.String(region)})
		go describeTrails(copySess, region, worker)
		<- worker
	}

	// Audit: Trigger warning if MultiRegion non-existent
	if multiRegionEnabled == false {
		json.NewEncoder(os.Stdout).Encode("Warning: A MultiRegion CloudTrail was not detected.")
	}

	done <- true
}
