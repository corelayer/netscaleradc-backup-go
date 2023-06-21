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

package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"

	"github.com/corelayer/netscaleradc-backup/pkg/config"
)

type Backup struct{}

func (b *Backup) Execute(c config.Application) {
	err := b.initializeOutputPaths(c)
	if err != nil {
		// TODO Error handling
	}

	var wg sync.WaitGroup
	for _, o := range c.Organization {
		for _, e := range o.Environments {
			wg.Add(1)
			outputPath := b.getOutputPath(c.Backup.BasePath, o.Name, e.Name, c.Backup.FolderPerEnvironment)
			go b.backupEnvironment(o.Name, e, outputPath, c.Backup.Level, &wg)
		}
	}
	wg.Wait()
}

func (b *Backup) initializeOutputPaths(c config.Application) error {
	var err error

	// Create directory defined as BasePath
	err = b.createDirectory(c.Backup.BasePath)
	if err != nil {
		return err
	}

	//
	if c.Backup.FolderPerEnvironment {
		for _, o := range c.Organization {
			for _, e := range o.Environments {
				for _, n := range e.Nodes {
					path := filepath.Join(c.Backup.BasePath, o.Name, e.Name, n.Name)
					err = b.createDirectory(path)
					if err != nil {
						// TODO log?
						return err
					}
				}
			}
		}
		return nil
	}

	for _, o := range c.Organization {
		path := filepath.Join(c.Backup.BasePath, o.Name)
		err = b.createDirectory(path)
		if err != nil {
			// TODO log?
			return err
		}
	}
	return nil
}

func (b *Backup) createDirectory(path string) error {
	src, err := os.Stat(path)

	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	} else if src.Mode().IsRegular() {
		return os.ErrExist
	} else {
		return nil
	}
}

func (b *Backup) getTimestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
}

func (b *Backup) getOutputPath(basePath string, organization string, environment string, folderPerEnvironment bool) string {
	if folderPerEnvironment {
		return filepath.Join(basePath, organization, environment)
	}
	return filepath.Join(basePath, organization)
}

func (b *Backup) backupEnvironment(organizationName string, e registry.Environment, outputPath string, level string, wg *sync.WaitGroup) {
	var err error
	var nitroClients map[string]*nitro.Client
	nitroClients, err = e.GetAllNitroClients()
	if err != nil {
		// TODO log?
		wg.Done()
		return
	}

	var primary string
	primary, err = e.GetPrimaryNodeName()
	if err != nil {
		// TODO log?
		wg.Done()
		return
	}

	backupName := b.getTimestamp() + ".tgz"
	err = b.createBackup(nitroClients[primary], backupName, level)
	if err != nil {
		// TODO log?
		wg.Done()
		return
	}

	for _, n := range e.Nodes {
		var r *io.Reader
		r, err = b.downloadBackup(nitroClients[n.Name], backupName)
		if err != nil {
			// TODO log?
			wg.Done()
			return
		}

		var output string
		if organizationName != "" {
			output = filepath.Join(outputPath, n.Name, organizationName+"_"+e.Name+"_"+n.Name+"_"+backupName)
		} else {
			output = filepath.Join(outputPath, n.Name, e.Name+"_"+n.Name+"_"+backupName)
		}
		err = b.writeBackupToDisk(output, r)
		if err != nil {
			// TODO log?
			wg.Done()
			return
		}

		err = b.deleteBackup(nitroClients[n.Name], backupName)
		if err != nil {
			// TODO log?
			wg.Done()
			return
		}
	}

}

func (b *Backup) createBackup(client *nitro.Client, name string, level string) error {
	// Filename must have no extension
	name = strings.TrimSuffix(name, ".tgz")

	c := controllers.NewBackupController(client)
	// TODO Change level to string (also in upstream nitroClient resource
	_, err := c.Create(name, level)
	return err
}

func (b *Backup) downloadBackup(client *nitro.Client, name string) (*io.Reader, error) {
	c := controllers.NewBackupController(client)
	return c.Download(name)

}

func (b *Backup) writeBackupToDisk(filepath string, data *io.Reader) error {
	reader := base64.NewDecoder(base64.StdEncoding, *data)
	buffer := bytes.Buffer{}

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		// TODO log?
		return err
	}

	err = os.WriteFile(filepath, buffer.Bytes(), 0644)
	return err
}

func (b *Backup) deleteBackup(client *nitro.Client, filename string) error {
	c := controllers.NewBackupController(client)
	_, err := c.Delete(filename)
	return err
}
