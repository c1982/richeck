package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

func main() {
	region := flag.String("region", "eu-central-1", "AWS Region")
	jsonformat := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	cfg, err := NewConfig(*region)
	if err != nil {
		panic(err)
	}

	// ris, err := DescribeEC2ReservedInstances(cfg)
	// if err != nil {
	// 	panic(err)
	// }

	// ec2nodes, err := EC2Usage(cfg)
	// if err != nil {
	// 	panic(err)
	// }

	cacheNodes, err := CacheNodeUsage(cfg)
	if err != nil {
		panic(err)
	}

	reservedCacheNodes, err := ReservedCacheNodes(cfg)
	if err != nil {
		panic(err)
	}
	richecks := CoverageReport(reservedCacheNodes, cacheNodes)
	if *jsonformat {
		out, err := json.Marshal(richecks)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	} else {
		fmt.Println("Reserved ID\tPayment\tType\tReserved/Usage\tCoverage")
		for _, r := range richecks {
			fmt.Printf("%s\t%s\t%s\t%d/%d\t%d%%\r\n", r.ReservedID, r.PaymentOption, r.InstanceType, r.Qty, r.Usage, r.Coverage)
		}
	}
}

func CoverageReport(ris []RICheck, usage Usages) []RICheck {
	for i := 0; i < len(ris); i++ {
		r := &ris[i]
		u, ok := usage[r.InstanceType]
		if !ok {
			r.Coverage = 0
			continue
		}
		if u == 0 {
			r.Coverage = 0
			continue
		}
		r.Usage = u
		r.Coverage = r.Qty * 100 / int32(u)
		if r.Coverage > 100 {
			r.Coverage = (r.Coverage - 100) * -1
		}
	}

	return ris
}
