// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

	"enen/gate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gateCmd represents the gate command
var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "代理服务",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("gate called")

		gate.Run()
	},
}

func init() {
	RootCmd.AddCommand(gateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	gateCmd.Flags().StringP("name", "n", "gate", "服务名称")

	viper.BindPFlag("gate.name", gateCmd.Flags().Lookup("name"))
}
