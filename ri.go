package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func NewConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func DescribeReservedInstances(cfg aws.Config) (reserverInstances []types.ReservedInstances, err error) {
	client := ec2.NewFromConfig(cfg)
	ris, err := client.DescribeReservedInstances(context.TODO(), &ec2.DescribeReservedInstancesInput{})
	if err != nil {
		return reserverInstances, err
	}

	return ris.ReservedInstances, nil
}

func DescribeEC2Instances(cfg aws.Config) (reservations []types.Instance, err error) {
	client := ec2.NewFromConfig(cfg)
	list, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return reservations, err
	}

	reservations = []types.Instance{}
	for _, r := range list.Reservations {
		reservations = append(reservations, r.Instances...)
	}

	return reservations, nil
}
