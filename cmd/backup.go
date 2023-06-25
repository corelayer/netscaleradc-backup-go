/*
 * Copyright 2023 CoreLayer BV
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup",
	Long:  `Create a backup of NetScaler ADC`,
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all jobs",
	Long:  `Run all jobs as defined in the config file`,
	Run:   allCmdRun,
}

var jobCmd = &cobra.Command{
	Use:   "job <name>",
	Short: "Run a specific job",
	Long:  `Run a specific job as defined in the config file`,
	Args:  cobra.MinimumNArgs(1),
	Run:   jobCmdRun,
}

var job string
var environment string

func initBackupCmd() {
	jobCmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment to backup")

	backupCmd.AddCommand(allCmd)
	backupCmd.AddCommand(jobCmd)
}

func allCmdRun(cmd *cobra.Command, args []string) {
	fmt.Println("backup all")
}

func jobCmdRun(cmd *cobra.Command, args []string) {
	if environment != "" {
		fmt.Printf("backup job %s - environment %s\n", args[0], environment)
	} else {
		fmt.Printf("backup job %s\n", args[0])
	}
}
