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

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/lighttiger2505/awspc/internal/api"
	"github.com/spf13/cobra"
)

// valueCmd represents the value command
var valueCmd = &cobra.Command{
	Use:   "value",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: serviceValueAct,
}

func init() {
	serviceCmd.AddCommand(valueCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// valueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// valueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serviceValueAct(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("require args")
	}

	in := &pricing.GetAttributeValuesInput{
		AttributeName: aws.String(args[1]),
		MaxResults:    nil,
		NextToken:     nil,
		ServiceCode:   aws.String(args[0]),
	}

	outs := []*pricing.GetAttributeValuesOutput{}
	for {
		out, err := getServiceAttributes(in)
		if err != nil {
			return err
		}
		outs = append(outs, out)
		if out.NextToken == nil {
			break
		}
		in.NextToken = out.NextToken
	}

	for _, out := range outs {
		for _, value := range out.AttributeValues {
			fmt.Println(aws.StringValue(value.Value))
		}
	}

	return nil
}

func getServiceAttributes(in *pricing.GetAttributeValuesInput) (*pricing.GetAttributeValuesOutput, error) {
	svc := api.GetPricingService()
	out, err := svc.GetAttributeValues(in)
	if err != nil {
		return nil, fmt.Errorf("filed get service attribute value, %s", err)
	}
	return out, nil
}
