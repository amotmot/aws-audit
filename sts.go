package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"os"
)

func describeIdentity(sess *session.Session) {

	svcSts := sts.New(sess)

	input := &sts.GetCallerIdentityInput{}

	result, err := svcSts.GetCallerIdentity(input)
	if err != nil {
		exitErrorf("Unable to get api identity", err)
	}

	json.NewEncoder(os.Stdout).Encode(result)

}