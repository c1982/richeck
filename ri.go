package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

type Usages map[string]int32
type Reserved map[string]RICheck
type RICheck struct {
	Qty      int32 `json:"qty"`
	Usage    int32 `json:"usage"`
	Coverage int32 `json:"coverage"`
}

func NewConfig(region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func ReservedEC2Instances(cfg aws.Config) (reservedInstances Reserved, err error) {
	client := ec2.NewFromConfig(cfg)
	ris, err := client.DescribeReservedInstances(context.TODO(), &ec2.DescribeReservedInstancesInput{})
	if err != nil {
		return reservedInstances, err
	}

	reservedInstances = Reserved{}
	for _, r := range ris.ReservedInstances {
		if r.State != "active" {
			continue
		}

		v, ok := reservedInstances[string(r.InstanceType)]
		if ok {
			v.Qty = v.Qty + *r.InstanceCount
			reservedInstances[string(r.InstanceType)] = v
		} else {
			reservedInstances[string(r.InstanceType)] = RICheck{
				Qty: *r.InstanceCount,
			}
		}
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

func ReservedCacheNodes(cfg aws.Config) (reservedcachenode Reserved, err error) {
	client := elasticache.NewFromConfig(cfg)
	list, err := client.DescribeReservedCacheNodes(context.TODO(), &elasticache.DescribeReservedCacheNodesInput{})
	if err != nil {
		return reservedcachenode, err
	}

	reservedcachenode = Reserved{}
	for _, r := range list.ReservedCacheNodes {
		if *r.State != "active" {
			continue
		}
		v, ok := reservedcachenode[string(*r.CacheNodeType)]
		if ok {
			v.Qty = v.Qty + r.CacheNodeCount
			reservedcachenode[string(*r.CacheNodeType)] = v
		} else {
			reservedcachenode[string(*r.CacheNodeType)] = RICheck{
				Qty: r.CacheNodeCount,
			}
		}
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
