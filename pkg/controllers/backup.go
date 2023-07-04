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
	"github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/config"
	nitroControllers "github.com/corelayer/netscaleradc-nitro-go/pkg/nitro/resource/controllers"
	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type Backup struct {
	path        string
	prefix      string
	level       string
	environment registry.Environment
}

func NewBackupController(path string, prefix string, level string, environment registry.Environment) Backup {
	return Backup{
		path:        path,
		prefix:      prefix,
		level:       level,
		environment: environment,
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

	timestamp := b.getTimestamp()
	backupName := timestamp + ".tgz"
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
			output = filepath.Join(b.path, timestamp+"_"+b.prefix+"_"+n.Name+".tgz")
		} else {
			output = filepath.Join(b.path, timestamp+"_"+n.Name+".tgz")
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
	// c := nitroControllers.NewBackupController(client)
	// return c.Download(name)

	var err error
	c := nitroControllers.NewSystemFileController(client)
	var res *nitro.Response[config.SystemFile]
	res, err = c.Get(name, "/var/ns_sys_backup", nil)
	if err != nil {
		return nil, err
	}

	if len(res.Data) != 1 {
		err = fmt.Errorf("invalid amount of files in data")
		return nil, err
	}

	var output io.Reader
	output = base64.NewDecoder(base64.StdEncoding, strings.NewReader(res.Data[0].Content))

	return &output, nil
}

func (b *Backup) WriteToDisk(path string, data *io.Reader) error {
	buffer := bytes.Buffer{}

	_, err := buffer.ReadFrom(*data)
	if err != nil {
		fmt.Println("Error reading")
		return err
	}

	err = os.WriteFile(path, buffer.Bytes(), 0644)
	return err
}

func (b *Backup) Delete(client *nitro.Client, filename string) error {
	c := nitroControllers.NewBackupController(client)
	_, err := c.Delete(filename)
	return err
}
