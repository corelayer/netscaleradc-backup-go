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
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"

	"github.com/corelayer/netscaleradc-backup/pkg/config"
	"github.com/corelayer/netscaleradc-backup/pkg/controllers"
)

var backupCmd = &cobra.Command{
	Use:              "backup",
	Short:            "Create backup",
	Long:             `Create a backup of NetScaler ADC`,
	PersistentPreRun: backupCmdPreRun,
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

var environmentArg string

func initBackupCmd() {
	jobCmd.Flags().StringVarP(&environmentArg, "environmentArg", "e", "", "Environment to backup")

	backupCmd.AddCommand(allCmd)
	backupCmd.AddCommand(jobCmd)
}

func backupCmdPreRun(cmd *cobra.Command, args []string) {
	var err error
	for _, job := range C.Jobs {
		err = createFolderPerJob(job)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func createFolderPerJob(job config.Job) error {
	var err error
	if job.BackupSettings.FolderPerEnvironment {
		err = createFolderPerEnvironment(job)
	} else {
		err = createDirectory(job.BackupSettings.Path)
	}
	return err
}

func createFolderPerEnvironment(job config.Job) error {
	var err error
	for _, env := range job.Environments {
		err = createDirectory(filepath.Join(job.BackupSettings.Path, env.Name))
		if err != nil {
			return err
		}
	}
	return err
}

func createDirectory(path string) error {
	src, err := os.Stat(path)

	if os.IsNotExist(err) {
		fmt.Printf("Creating output destination: %s\n", path)
		return os.MkdirAll(path, 0755)
	} else if src.Mode().IsRegular() {
		fmt.Printf("Output destination '%s' exists, but is a file\n")
		return os.ErrExist
	} else {
		return nil
	}
}

func allCmdRun(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup
	for _, job := range C.Jobs {
		wg.Add(1)
		go func(j config.Job, e string) {
			runJob(j, e)
			wg.Done()
		}(job, "")
	}
	wg.Wait()
}

func jobCmdRun(cmd *cobra.Command, args []string) {
	job, err := C.GetJob(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(j config.Job, e string) {
		runJob(j, e)
		wg.Done()
	}(job, environmentArg)
	wg.Wait()
}

func runJob(job config.Job, environmentName string) {
	var wg sync.WaitGroup
	var err error
	if environmentName != "" {
		var env registry.Environment
		env, err = job.GetEnvironment(environmentName)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Running job %s for environment %s\n", job.Name, env.Name)
		wg.Add(1)
		go func(e registry.Environment, s config.BackupSettings) {
			doBackup(getOutputPath(s.Path, e.Name, s.FolderPerEnvironment), s.Prefix, s.Level, e)
			wg.Done()
		}(env, job.BackupSettings)
	} else {
		for _, env := range job.Environments {
			fmt.Printf("Running job %s for environment %s\n", job.Name, env.Name)
			wg.Add(1)
			go func(e registry.Environment, s config.BackupSettings) {
				doBackup(getOutputPath(s.Path, e.Name, s.FolderPerEnvironment), s.Prefix, s.Level, e)
				wg.Done()
			}(env, job.BackupSettings)
		}
	}
	wg.Wait()
}

func doBackup(path string, prefix string, level string, e registry.Environment) {
	c := controllers.NewBackupController(path, prefix, level, e)
	err := c.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func getOutputPath(basePath string, environment string, folderPerEnvironment bool) string {
	if folderPerEnvironment {
		return filepath.Join(basePath, environment)
	}
	return filepath.Join(basePath)
}
