package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	region := flag.String("region", "eu-central-1", "AWS Region")
	jsonformat := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	cfg, err := NewConfig(*region)
	if err != nil {
		panic(err)
	}

	ec2Usage, err := EC2Usage(cfg)
	if err != nil {
		panic(err)
	}

	reservedEC2nodes, err := ReservedEC2Instances(cfg)
	if err != nil {
		panic(err)
	}

	cacheNodes, err := CacheNodeUsage(cfg)
	if err != nil {
		panic(err)
	}

	reservedCacheNodes, err := ReservedCacheNodes(cfg)
	if err != nil {
		panic(err)
	}

	cacheRIchecks := coverageReport(reservedCacheNodes, cacheNodes)
	ec2RIchecks := coverageReport(reservedEC2nodes, ec2Usage)
	if *jsonformat {
		printJSONReport(*region, cacheRIchecks, ec2RIchecks)
	} else {
		printTextReport("ElastiCache", *region, cacheRIchecks)
		printTextReport("EC2", *region, ec2RIchecks)
	}
}

func coverageReport(ris Reserved, usage Usages) Reserved {
	for k, r := range ris {
		u, ok := usage[k]
		if !ok {
			continue
		}

		if u == 0 {
			continue
		}

		r.Usage = u
		r.Coverage = r.Qty * 100 / u
		if r.Coverage > 100 {
			r.Coverage = (r.Coverage - 100) * -1
		}

		ris[k] = r
	}

	return ris
}

func printTextReport(service, region string, ris Reserved) {
	fmt.Printf("\n%s Reserved Instances in %s\n", service, region)
	fmt.Println("Type\t\tQty\tUsage\tCoverage")
	for k, r := range ris {
		fmt.Printf("%s\t%d\t%d\t%d%%\n", k, r.Qty, r.Usage, r.Coverage)
	}
}

func printJSONReport(region string, cacheRIchecks, ec2RIchecks Reserved) {
	v := struct {
		Region        string   `json:"region"`
		ElastiCacheRI Reserved `json:"elasticache"`
		EC2RI         Reserved `json:"ec2"`
	}{
		region,
		cacheRIchecks,
		ec2RIchecks,
	}

	out, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(out)
}
