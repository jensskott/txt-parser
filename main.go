package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const appVersion = "0.1.0"

var filterTags TagNames

func main() {
	flag.Var(&filterTags, "t", "List of tags to filter")
	flag.Parse()

	tags, err := getInstanceTags()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found")
		return
	}

	for _, tag := range tags {

		if len(filterTags) > 0 {

			for _, filter := range filterTags {
				if *tag.Key == filter {
					fmt.Printf("%s = %s\n", *tag.Key, *tag.Value)
				}
			}

		} else {
			fmt.Printf("%s = %s\n", *tag.Key, *tag.Value)
		}
	}
}

func getInstanceTags() ([]*Tag, error) {

	tags := []*Tag{}

	session, _ := session.NewSession()
	ec2meta := ec2metadata.New(session)

	// If not running on AWS - bail out
	if !ec2meta.Available() {
		return nil, errors.New("Not running on AWS")
	}

	identity, err := ec2meta.GetInstanceIdentityDocument()
	if err != nil {
		return nil, errors.New("Cannot query AWS instance identity")
	}

	// Connecting to instance region
	ec2svc := ec2.New(session, &aws.Config{Region: &identity.Region})

	// Tag description filter
	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					&identity.InstanceID,
				},
			},

			/*{
				Name: aws.String("key"),
				Values: []*string{
					aws.String(*c.TagName),
				},
			},*/
		},
	}

	tagResponse, err := ec2svc.DescribeTags(params)

	if err != nil {
		return nil, errors.New("Failed getting tags: " + err.Error())
	}

	respTags := tagResponse.Tags

	for _, td := range respTags {
		tag := &Tag{
			Key:   td.Key,
			Value: td.Value,
		}

		tags = append(tags, tag)
	}

	return tags, nil
}
