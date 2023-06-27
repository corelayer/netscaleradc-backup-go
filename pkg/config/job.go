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

package config

import (
	"fmt"

	"github.com/corelayer/netscaleradc-nitro-go/pkg/registry"
)

type Job struct {
	Name           string                 `json:"name" yaml:"name" mapstructure:"name"`
	Environments   []registry.Environment `json:"environments" yaml:"environments" mapstructure:"environments"`
	BackupSettings BackupSettings         `json:"backupSettings" yaml:"backupSettings" mapstructure:"backupSettings"`
}

func (j *Job) GetEnvironment(name string) (registry.Environment, error) {
	for _, e := range j.Environments {
		if e.Name == name {
			return e, nil
		}
	}
	return registry.Environment{}, fmt.Errorf("environment '%s' not found", name)
}
