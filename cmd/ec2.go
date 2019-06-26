// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/olekukonko/tablewriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/spf13/cobra"
)

// ec2Cmd represents the ec2 command
var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: ec2Act,
}

func init() {
	rootCmd.AddCommand(ec2Cmd)
	ec2Cmd.PersistentFlags().String("instance-famiry", "m", "show specific instance famiry")
	ec2Cmd.PersistentFlags().String("instance-type", "i", "show specific instance type")
}

func ec2Act(cmd *cobra.Command, args []string) error {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create a Pricing client from just a session.
	svc := pricing.New(session)

	filters := []*pricing.Filter{
		// &pricing.Filter{
		// 	Field: aws.String("instanceType"),
		// 	Type:  aws.String(pricing.FilterTypeTermMatch),
		// 	Value: aws.String("c5.large"),
		// },
		&pricing.Filter{
			Field: aws.String("capacitystatus"),
			Type:  aws.String(pricing.FilterTypeTermMatch),
			Value: aws.String("Used"),
		},
		&pricing.Filter{
			Field: aws.String("tenancy"),
			Type:  aws.String(pricing.FilterTypeTermMatch),
			Value: aws.String("Shared"),
		},
		// &pricing.Filter{
		// 	Field: aws.String("instanceFamily"),
		// 	Type:  aws.String(pricing.FilterTypeTermMatch),
		// 	Value: aws.String("General purpose"),
		// },
		&pricing.Filter{
			Field: aws.String("location"),
			Type:  aws.String(pricing.FilterTypeTermMatch),
			Value: aws.String("Asia Pacific (Tokyo)"),
		},
		&pricing.Filter{
			Field: aws.String("operatingSystem"),
			Type:  aws.String(pricing.FilterTypeTermMatch),
			Value: aws.String("Linux"),
		},
		&pricing.Filter{
			Field: aws.String("termType"),
			Type:  aws.String(pricing.FilterTypeTermMatch),
			Value: aws.String("OnDemand"),
		},
	}
	in := &pricing.GetProductsInput{
		Filters:       filters,
		FormatVersion: nil,
		MaxResults:    nil,
		NextToken:     nil,
		ServiceCode:   aws.String("AmazonEC2"),
	}

	ec2List := []*EC2{}

	for {
		out, err := svc.GetProducts(in)
		if err != nil {
			return fmt.Errorf("filed get ec2 product price, %s", err)
		}

		for _, jsonValue := range out.PriceList {
			ec2List = append(ec2List, parseEC2(jsonValue))
		}
		// fmt.Println(out.String())

		if out.NextToken == nil {
			break
		}
		in.NextToken = out.NextToken
	}

	sort.Slice(ec2List, func(i, j int) bool {
		if ec2List[i].InstanceFamily < ec2List[j].InstanceFamily {
			return true
		}
		if ec2List[i].InstanceFamily > ec2List[j].InstanceFamily {
			return false
		}
		return ec2List[i].InstanceType < ec2List[j].InstanceType

	})

	data := [][]string{}
	for _, ec2 := range ec2List {
		data = append(data, ec2.StringSlice())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Instance Famiry", "Instance Type", "vCPU", "ECU", "Memory", "Storage", "USD"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

type EC2 struct {
	InstanceFamily string
	InstanceType   string
	vCPU           string
	ECU            string
	Memory         string
	Storage        string
	Price          string
}

func parseEC2(jsonValue aws.JSONValue) *EC2 {
	// AmazonEC2 attributes
	product := jsonValue["product"].(map[string]interface{})
	attr := product["attributes"].(map[string]interface{})

	// AmazonEC2 price
	usd := ""
	terms := jsonValue["terms"].(map[string]interface{})
	rawOndemand := terms["OnDemand"]
	// if !ok {
	// 	continue
	// }
	ondemands := rawOndemand.(map[string]interface{})
	for _, ondemand := range ondemands {
		priceDimensions := ondemand.(map[string]interface{})["priceDimensions"].(map[string]interface{})
		for _, priceDimension := range priceDimensions {
			pricePerUnit := priceDimension.(map[string]interface{})["pricePerUnit"]
			usd = pricePerUnit.(map[string]interface{})["USD"].(string)
		}

	}

	ec2 := &EC2{
		InstanceFamily: attr["instanceFamily"].(string),
		InstanceType:   attr["instanceType"].(string),
		vCPU:           attr["vcpu"].(string),
		ECU:            attr["ecu"].(string),
		Memory:         attr["memory"].(string),
		Storage:        attr["storage"].(string),
		Price:          usd,
	}
	return ec2
}

func (e *EC2) StringSlice() []string {
	return []string{
		e.InstanceFamily,
		e.InstanceType,
		e.vCPU,
		e.ECU,
		e.Memory,
		e.Storage,
		e.Price,
	}
}
