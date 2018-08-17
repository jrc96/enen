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

	"enen/center"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// centerCmd represents the center command
var centerCmd = &cobra.Command{
	Use:   "center",
	Short: "中心服务",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("center called")

		center.Run()
	},
}

func init() {
	RootCmd.AddCommand(centerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// centerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// centerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	centerCmd.Flags().StringP("name", "n", "center", "服务名称")
	centerCmd.Flags().BoolP("debug", "d", true, "调试模式")

	viper.BindPFlag("center.name", centerCmd.Flags().Lookup("name"))
	viper.BindPFlag("center.debug", centerCmd.Flags().Lookup("debug"))
}
