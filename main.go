package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
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
	for k, u := range usage {
		ri, ok := ris[k]
		if ok {
			if u == 0 {
				continue
			}

			ri.Usage = u
			ri.Reserved = true
			ri.Coverage = ri.Qty * 100 / u
			if ri.Coverage > 100 {
				ri.Coverage = (ri.Coverage - 100) * -1
			}

			ris[k] = ri
		} else {
			ris[k] = RICheck{
				Usage: u,
			}
		}
	}

	return ris
}

func printTextReport(service, region string, ris Reserved) {
	fmt.Printf("\n%s Reserved Instances in %s\n", service, region)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Type\tQty\tUsage\tCoverage\tReserved")

	//TODO: add sort interface to RICheck type.
	for k, r := range ris {
		if r.Reserved {
			fmt.Fprintf(w, "%s\t%d\t%d\t%d%%\t%v\n", k, r.Qty, r.Usage, r.Coverage, r.Reserved)
		}
	}
	for k, r := range ris {
		if !r.Reserved {
			fmt.Fprintf(w, "%s\t%d\t%d\t%d%%\t%v\n", k, r.Qty, r.Usage, r.Coverage, r.Reserved)
		}
	}

	w.Flush()
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
