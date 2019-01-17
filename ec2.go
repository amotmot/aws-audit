package main

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2Instance struct {
	ReservationId	string
	InstanceId		string
	State 			string
	PublicDnsName 	string
	PublicIpAddress string
}

type Ingress struct {
	Protocol string
	FromPort int64
	ToPort   int64
	Source   []string
}

type SecurityGroup struct {
	Id          string
	Name        string
	Ingress     []Ingress
}

var publicSecurityGroups = make(map[string]string)
var publicEc2Instance = make(map[string]string)

func auditEC2(sess *session.Session, done chan bool) {

	fmt.Println("Auditing EC2")

	svcEc2 := ec2.New(sess)

	result, err := svcEc2.DescribeInstances(nil)
	if err != nil  {
		exitErrorf("Unable to describe instances", err)
	}

	for idx, res := range result.Reservations {

		for _, inst := range result.Reservations[idx].Instances {

			// Does EC2 instance have a public IP?
			if inst.PublicIpAddress != nil {

				instanceEc2, _ := json.Marshal(&Ec2Instance{*res.ReservationId, *inst.InstanceId, *inst.State.Name, *inst.PublicDnsName, *inst.PublicIpAddress})
				json.NewEncoder(os.Stdout).Encode(string(instanceEc2))

				if len(inst.SecurityGroups) > 2 {
					for idx, _ := range inst.SecurityGroups {
						publicSecurityGroups[*inst.SecurityGroups[idx].GroupName] = *inst.SecurityGroups[idx].GroupId
					}
				} else {
					publicSecurityGroups[*inst.SecurityGroups[0].GroupName] = *inst.SecurityGroups[0].GroupId
				}
			}
		}
	}

	// TODO // REFACTOR -> to func
	var sgs = make([]SecurityGroup, len(publicSecurityGroups))
	for sgName, sgId := range publicSecurityGroups {

		resultSG, _ := svcEc2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
			GroupIds: aws.StringSlice([]string{sgId}),
		})

		for sgcount, g := range resultSG.SecurityGroups {
			ips := make([]Ingress, len(g.IpPermissions))

			if len(g.IpPermissions) > 0 {
				ips := make([]Ingress, len(g.IpPermissions))

				for ipcount, ip := range g.IpPermissions {
					if *ip.IpProtocol != "-1" {
						sources := make([]string, len(ip.IpRanges)+len(ip.UserIdGroupPairs))
						sourcecount := 0
						for _, source := range ip.IpRanges {
							sources[sourcecount] = *source.CidrIp
							sourcecount++
						}
						ips[ipcount] = Ingress{*ip.IpProtocol, *ip.FromPort, *ip.ToPort, sources}

					}
				}
			} else {
				sgs[sgcount] = SecurityGroup{sgId, sgName, nil}
			}
			sgs[sgcount] = SecurityGroup{sgId, sgName,  ips}
		}
		json.NewEncoder(os.Stdout).Encode(sgs)

		// TODO // Audit security groups for open ports, e.g.
		// All protocols open to 0.0.0.0/0 and ::/0
		// Chlorish - TCP Port 1433 open to 0.0.0.0/0 and ::/0
		// launch-wizard-# has TCP port 22 open to 0.0.0.0/0
	}
	done <- true
}
