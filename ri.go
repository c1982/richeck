package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

type Usages map[string]int32

type RICheck struct {
	ReservedID    string `json:"id"`
	InstanceType  string `json:"type"`
	Qty           int32  `json:"qty"`
	Usage         int32  `json:"usage"`
	Coverage      int32  `json:"coverage"`
	TimeLeft      int64  `json:"time_left"`
	PaymentOption string `json:"payment_option"`
}

func NewConfig(region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func ReservedEC2Instances(cfg aws.Config) (reservedInstances []RICheck, err error) {
	client := ec2.NewFromConfig(cfg)
	ris, err := client.DescribeReservedInstances(context.TODO(), &ec2.DescribeReservedInstancesInput{})
	if err != nil {
		return reservedInstances, err
	}

	reservedInstances = []RICheck{}
	for _, r := range ris.ReservedInstances {
		//TODO: state=active
		reservedInstances = append(reservedInstances, RICheck{
			ReservedID:    *r.ReservedInstancesId,
			InstanceType:  string(r.InstanceType),
			Qty:           *r.InstanceCount,
			TimeLeft:      *r.Duration,
			PaymentOption: string(r.OfferingType),
		})
	}

	return reservedInstances, nil
}

func EC2Usage(cfg aws.Config) (usages Usages, err error) {
	client := ec2.NewFromConfig(cfg)
	list, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return usages, err
	}

	usages = make(map[string]int32)
	for _, r := range list.Reservations {
		for _, i := range r.Instances {
			if i.State.Name != "running" {
				continue
			}
			v, ok := usages[string(i.InstanceType)]
			if ok {
				usages[string(i.InstanceType)] = v + 1
			} else {
				usages[string(i.InstanceType)] = 1
			}
		}
	}

	return usages, nil
}

func ReservedCacheNodes(cfg aws.Config) (reservedcachenode []RICheck, err error) {
	client := elasticache.NewFromConfig(cfg)
	list, err := client.DescribeReservedCacheNodes(context.TODO(), &elasticache.DescribeReservedCacheNodesInput{})
	if err != nil {
		return reservedcachenode, err
	}

	reservedcachenode = []RICheck{}
	for _, r := range list.ReservedCacheNodes {
		//TODO: state=active
		reservedcachenode = append(reservedcachenode, RICheck{
			ReservedID:    *r.ReservedCacheNodeId,
			InstanceType:  *r.CacheNodeType,
			Qty:           r.CacheNodeCount,
			TimeLeft:      int64(r.Duration),
			PaymentOption: string(*r.OfferingType),
		})
	}

	return reservedcachenode, nil
}

func CacheNodeUsage(cfg aws.Config) (usages Usages, err error) {
	client := elasticache.NewFromConfig(cfg)
	list, err := client.DescribeCacheClusters(context.TODO(), &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		return usages, err
	}

	usages = Usages{}
	for _, c := range list.CacheClusters {
		v, ok := usages[string(*c.CacheNodeType)]
		if ok {
			usages[string(*c.CacheNodeType)] = v + 1
		} else {
			usages[string(*c.CacheNodeType)] = 1
		}
	}

	return usages, nil
}
