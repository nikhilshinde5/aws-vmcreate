package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"strings"

	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var client *ec2.Client

// EC2CreateInstanceAPI defines the interface for the RunInstances and CreateTags functions.
// We use this interface to test the functions using a mocked service.
type EC2CreateInstanceAPI interface {
	RunInstances(ctx context.Context,
		params *ec2.RunInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)

	CreateTags(ctx context.Context,
		params *ec2.CreateTagsInput,
		optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)

	TerminateInstances(ctx context.Context,
		params *ec2.TerminateInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)

	// ModifyInstanceAttribute(ctx context.Context,
	// 	params *ec2.ModifyInstanceAttributeInput,
	// 	optFns ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error)
	// StopInstances(ctx context.Context,
	// 	params *ec2.StopInstancesInput,
	// 	optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}

type ConfigMap struct {
	InstanceType string `json:"instance_type"`
	ImageId      string `json:"image_id"`
}

// MakeInstance creates an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region.
//	api is the interface that defines the method call.
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a RunInstancesOutput object containing the result of the service call and nil.
//	Otherwise, nil and an error from the call to RunInstances.
func MakeInstance(c context.Context, api EC2CreateInstanceAPI, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
	return api.RunInstances(c, input)
}

// MakeTags creates tags for an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region.
//	api is the interface that defines the method call.
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a CreateTagsOutput object containing the result of the service call and nil.
//	Otherwise, nil and an error from the call to CreateTags.
func MakeTags(c context.Context, api EC2CreateInstanceAPI, input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	return api.CreateTags(c, input)
}

// DeleteInstance deletes an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region.
//	api is the interface that defines the method call.
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a TerminateInstancesInput object containing the result of the service call and nil.
//	Otherwise, nil and an error from the call to TerminateInstances.
func DeleteInstance(c context.Context, api EC2CreateInstanceAPI, input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	return api.TerminateInstances(c, input)
}

// func UpdateInstanceAttribute(c context.Context, api EC2CreateInstanceAPI, input *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error) {
// 	return api.ModifyInstanceAttribute(c, input)
// }

// func PauseInstances(c context.Context, api EC2CreateInstanceAPI, input *ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error) {
// 	return api.StopInstances(c, input)
// }

func DeleteInstancesCmd(name *string, value *string) {

	var instanceIds = make([]string, 0)

	val := strings.Split(*value, ",")
	tag := "tag:" + *name

	describeInput := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(tag),
				Values: val,
			},
		},
	}
	result, err := client.DescribeInstances(context.TODO(), describeInput)
	if err != nil {
		fmt.Println("Got an error fetching the status of the instance")
		fmt.Println(err)
	} else {
		for _, r := range result.Reservations {
			fmt.Println("Instance IDs:")
			for _, i := range r.Instances {
				instanceIds = append(instanceIds, *i.InstanceId)
			}
			fmt.Println(instanceIds)
		}

		input := &ec2.TerminateInstancesInput{
			InstanceIds: instanceIds,
			DryRun:      new(bool),
		}

		result, err := DeleteInstance(context.TODO(), client, input)
		if err != nil {
			fmt.Println("Got an error terminating the instance:")
			fmt.Println(err)
			return
		}

		fmt.Println("Terminated instance with id: ", *result.TerminatingInstances[0].InstanceId)
	}
}

func CreateInstancesCmd(name *string, value *string) {
	// Create separate values if required.
	minMaxCount := int32(1)

	file, err := os.Open("data/config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var config ConfigMap
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		fmt.Println("Error decoding config:", err)
		os.Exit(1)
	}

	// instanceType := &config.InstanceType

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(config.ImageId),
		InstanceType: (types.InstanceType)(config.InstanceType),
		MinCount:     &minMaxCount,
		MaxCount:     &minMaxCount,
	}

	result, err := MakeInstance(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error creating an instance:")
		fmt.Println(err)
		return
	}

	tagInput := &ec2.CreateTagsInput{
		Resources: []string{*result.Instances[0].InstanceId},
		Tags: []types.Tag{
			{
				Key:   name,
				Value: value,
			},
		},
	}

	_, err = MakeTags(context.TODO(), client, tagInput)
	if err != nil {
		fmt.Println("Got an error tagging the instance:")
		fmt.Println(err)
		return
	}

	fmt.Println("Created tagged instance with ID " + *result.Instances[0].InstanceId)

	//Testing change of instanceType
	// fmt.Println("Updating instance type of instance with ID " + *result.Instances[0].InstanceId)
	// time.Sleep(30 * time.Second)

	// instanceID := *result.Instances[0].InstanceId
	// newInstanceType := "t2.nano"

	// //Stopping instances before changing instance type
	// fmt.Println("Stopping instances before changing instance type")
	// stopInstancesInput := &ec2.StopInstancesInput{
	// 	InstanceIds: []string{instanceID},
	// 	Force:       aws.Bool(false),
	// }
	// _, err = PauseInstances(context.TODO(), client, stopInstancesInput)
	// if err != nil {
	// 	fmt.Println("Got an error stoping the instance:")
	// 	fmt.Println(err)
	// 	return
	// }

	// //this sleep for letting ec2 stop
	// time.Sleep(60 * time.Second)
	// Wait for the instance to be stopped
	// for {
	// 	describeOutput, err := svc.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
	// 		InstanceIds: []string{instanceID},
	// 	})
	// 	if err != nil {
	// 		fmt.Println("Error describing instance:", err)
	// 		return
	// 	}
	// 	if describeOutput.Reservations[0].Instances[0].State.Name == ec2.InstanceStateNameStopped {
	// 		break
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }

	// Modify the instance type
	// _, err = svc.ModifyInstanceAttribute(context.Background(), &ec2.ModifyInstanceAttributeInput{
	// 	InstanceId: &instanceID,
	// 	InstanceType: &ec2.AttributeValue{
	// 		Value: &newInstanceType,
	// 	},

	// attributeInput := &ec2.ModifyInstanceAttributeInput{
	// 	InstanceId: &instanceID,
	// 	InstanceType: &types.AttributeValue{
	// 		Value: &newInstanceType,
	// 	},
	// }

	// _, err = UpdateInstanceAttribute(context.TODO(), client, attributeInput)
	// if err != nil {
	// 	fmt.Println("Got an error updating the instance:")
	// 	fmt.Println(err)
	// 	return
	// }

}
func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = ec2.NewFromConfig(cfg)

}
func main() {
	fmt.Println("Provisioning/De-provisioning EC2 in progress")
	command := flag.String("c", "", "command  create or delete")
	name := flag.String("n", "", "The name of the tag to attach to the instance")
	value := flag.String("v", "", "The value of the tag to attach to the instance")
	// imageId := flag.String("i", "", "The instance id of the instance")
	// instanceTypeString := flag.String("t", "", "The type of the instance")

	flag.Parse()

	if *command == "" {
		fmt.Println("You must supply an command  start or stop (-c start")
		return
	}

	if *name == "" || *value == "" {
		fmt.Println("You must supply a name and value for the tag (-n NAME -v VALUE)")
		return
	}

	if *command == "create" {
		CreateInstancesCmd(name, value)
	}

	if *command == "delete" {
		DeleteInstancesCmd(name, value)
	}
}
