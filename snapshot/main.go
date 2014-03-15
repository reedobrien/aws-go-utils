// Copyright 2014 Reed O'Brien <reed@reedobrien.com>.
// All rights reserved. Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// make snapshots of AWS instance's EBS volumes (up to two)

package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
)

const (
	TIME_FORMAT = "2006-01-02T15:04:05MST"
)

var snapIds []string
var tag ec2.Tag
var tags []ec2.Tag
var wg sync.WaitGroup

var period = flag.String("period", "daily", "Period value to use with inst_snap tag on the snapshots. Default 'daily', so <inst_id>/daily.")
var copies = flag.Int("copies", 3, "Number of completed (not pending) copies to keep")

func main() {
	flag.Parse()

	auth, err := aws.GetAuth("", "", "", time.Now().Add(time.Hour))
	if err != nil {
		panic(err)
	}

	filter := ec2.NewFilter()
	filter.Add("instance-state-name", "running")
	filter.Add("tag:env", "prod")

	c := ec2.New(auth, aws.USEast)
	resp, err := c.DescribeInstances(nil, filter)
	if err != nil {
		log.Panicln(err)
	}
	rezzies := resp.Reservations

	for _, rv := range rezzies {
		for _, inst := range rv.Instances {
			if len(inst.BlockDevices) < 3 {
				for _, bd := range inst.BlockDevices {
					vid := bd.EBS.VolumeId
					name, err := getName(inst.Tags)
					log.Printf("Creating snapshot for: %s volume: %v\n", name, vid)
					if err != nil {
						log.Fatalf("Error getting name:", err)
					}

					stamp := time.Now().UTC().Format(TIME_FORMAT)
					snprsp, err := c.CreateSnapshot(vid, fmt.Sprintf("%s %s %s", name, *period, stamp))
					if err != nil {
						log.Printf("Failed to snap: %s, Error: %s\n", vid, err)
						break
					} else {
						log.Printf("Created snap: %s\n", snprsp.Id)
						t := ec2.Tag{Key: "inst_snap", Value: fmt.Sprintf("%s/%s", inst.InstanceId, *period)}
						tags = append(inst.Tags, t)
						wg.Add(1)
						go tagSnapshot(inst.InstanceId, snprsp.Id, tags, c)
					}
				}
			}
		}
	}
	wg.Wait()
}

func tagSnapshot(instId, snapId string, tags []ec2.Tag, c *ec2.EC2) {
	defer wg.Done()
	_, err := c.CreateTags([]string{snapId}, tags)
	if err != nil {
		log.Printf("failed to tag snaptshot %s for instance %s, error: %s", snapId, instId, err)
	}
	trimSnapshots(instId, c)
}

func trimSnapshots(instId string, c *ec2.EC2) {
	filter := ec2.NewFilter()
	val := fmt.Sprintf("%s/%s", instId, *period)
	filter.Add("status", "completed")
	filter.Add("tag:inst_snap", val)
	resp, err := c.Snapshots(nil, filter)
	if err != nil {
		log.Printf("Error getting existing snapshots: ", err)
	}
	if len(resp.Snapshots) > *copies {
		excess := len(resp.Snapshots) - *copies
		extras := resp.Snapshots[:excess]
		log.Printf("Need %d have %d completed snapshots\n", *copies, len(resp.Snapshots))
		for _, extra := range extras {
			log.Printf("Trimming %s for %sn", extra.Id, instId)
			_, err := c.DeleteSnapshots([]string{extra.Id})
			if err != nil {
				log.Println("Failed trimming snapshot:", extra.Id, " Error: ", err)
			}
		}

	}
}

func getName(tags []ec2.Tag) (name string, err error) {
	for _, v := range tags {
		if v.Key == "Name" {
			name = v.Value
		}
	}
	return
}
