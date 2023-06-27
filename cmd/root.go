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
	"os"

	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-backup/pkg/config"
)

var C config.Application
var rootCmd = &cobra.Command{
	Use: "nsbackup",
	// CompletionOptions: cobra.CompletionOptions{
	// 	DisableDefaultCmd: true},
	Short: "Backup utility for NetScaler ADC",
	Long: `Backup utility for NetScaler ADC
Complete documentation is available at http://github.com/corelayer/netscaleradc-backup-go`,
	// Run:     rootCmdRun,
	Version: "build",
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	fmt.Println("Execute root")
}

func init() {
	initBackupCmd()

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(validateCmd)
}

func Execute(c config.Application) {
	C = c
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("EXIT WITH ERROR")
		fmt.Println(err)
		os.Exit(1)
	}
}
