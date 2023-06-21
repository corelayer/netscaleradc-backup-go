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
	"time"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro"
	nitroControllers "github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type Backup struct {
	basePath    string
	prefix      string
	environment registry.Environment
	level       string
}

//func (b *Backup) Execute(c config.Application) {
//	err := b.initializeOutputPaths(c)
//	if err != nil {
//		// Error handling
//	}
//
//	var wg sync.WaitGroup
//	for _, o := range c.Organization {
//		for _, e := range o.Environments {
//			wg.Add(1)
//			outputPath := b.getOutputPath(c.Backup.BasePath, o.Name, e.Name, c.Backup.FolderPerEnvironment)
//			go b.backupEnvironment(o.Name, e, outputPath, c.Backup.Level, &wg)
//		}
//	}
//	wg.Wait()
//}

//func (b *Backup) initializeOutputPaths() error {
//	var err error
//
//	for _, n := range b.environment.Nodes {
//		path := filepath.Join(b.basePath, b.prefix, b.environment.Name, n.Name)
//		err = b.createDirectory(path)
//		if err != nil {
//			// log?
//			return err
//		}
//	}
//
//	return nil
//}

//func (b *Backup) createDirectory(path string) error {
//	src, err := os.Stat(path)
//
//	if os.IsNotExist(err) {
//		return os.MkdirAll(path, 0755)
//	} else if src.Mode().IsRegular() {
//		return os.ErrExist
//	} else {
//		return nil
//	}
//}

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

//func (b *Backup) getOutputPath(basePath string, organization string, environment string, folderPerEnvironment bool) string {
//	if folderPerEnvironment {
//		return filepath.Join(basePath, organization, environment)
//	}
//	return filepath.Join(basePath, organization)
//}

func (b *Backup) Execute() error {
	var err error
	var nitroClients map[string]*nitro.Client
	nitroClients, err = b.environment.GetAllNitroClients()
	if err != nil {
		return err
	}

	var primary string
	primary, err = b.environment.GetPrimaryNodeName()
	if err != nil {
		return err
	}

	backupName := b.getTimestamp() + ".tgz"
	err = b.Create(nitroClients[primary], backupName, b.level)
	if err != nil {
		return err
	}

	for _, n := range b.environment.Nodes {
		var r *io.Reader
		r, err = b.Download(nitroClients[n.Name], backupName)
		if err != nil {
			return err
		}

		var output string
		if b.prefix != "" {
			output = filepath.Join(b.basePath, n.Name, b.prefix+"_"+b.environment.Name+"_"+n.Name+"_"+backupName)
		} else {
			output = filepath.Join(b.basePath, n.Name, b.environment.Name+"_"+n.Name+"_"+backupName)
		}
		err = b.WriteToDisk(output, r)
		if err != nil {
			return err
		}

		err = b.Delete(nitroClients[n.Name], backupName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Backup) Create(client *nitro.Client, name string, level string) error {
	// Filename must have no extension
	name = strings.TrimSuffix(name, ".tgz")

	c := nitroControllers.NewBackupController(client)
	_, err := c.Create(name, level)
	return err
}

func (b *Backup) Download(client *nitro.Client, name string) (*io.Reader, error) {
	c := nitroControllers.NewBackupController(client)
	return c.Download(name)

}

func (b *Backup) WriteToDisk(filepath string, data *io.Reader) error {
	reader := base64.NewDecoder(base64.StdEncoding, *data)
	buffer := bytes.Buffer{}

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, buffer.Bytes(), 0644)
	return err
}

func (b *Backup) Delete(client *nitro.Client, filename string) error {
	c := nitroControllers.NewBackupController(client)
	_, err := c.Delete(filename)
	return err
}
