package main

import (
	"encoding/json"
	"fmt"
)

type RICheck struct {
	ReservedID      string `json:"id"`
	InstanceType    string `json:"type"`
	Qty             int32  `json:"qty"`
	Usage           int32  `json:"usage"`
	EfficentPercent int32  `json:"efficient_percent"`
	TimeLeft        int64  `json:"time_left"`
	PaymentOption   string `json:"payment_option"`
}

//TODO: Add multi-region
//TODO: Add elasticache's RI
//TODO: Add text output
func main() {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}

	ris, err := DescribeReservedInstances(cfg)
	if err != nil {
		panic(err)
	}

	instances, err := DescribeEC2Instances(cfg)
	if err != nil {
		panic(err)
	}

	richecks := []RICheck{}
	for _, r := range ris {
		usage := 0
		//TODO: state=active
		for _, e := range instances {
			if e.State.Name != "running" {
				continue
			}

			if e.InstanceType == r.InstanceType {
				usage++
			}
		}
		efficent := *r.InstanceCount * 100 / int32(usage)
		if efficent > 100 {
			efficent = (efficent - 100) * -1
		}

		richecks = append(richecks, RICheck{
			ReservedID:      *r.ReservedInstancesId,
			InstanceType:    string(r.InstanceType),
			Qty:             *r.InstanceCount,
			Usage:           int32(usage),
			EfficentPercent: efficent,
			TimeLeft:        *r.Duration,
			PaymentOption:   string(r.OfferingType),
		})
	}

	out, err := json.Marshal(richecks)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
