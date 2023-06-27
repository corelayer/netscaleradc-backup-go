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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/corelayer/netscaleradc-backup/cmd"
	"github.com/corelayer/netscaleradc-backup/pkg/config"
)

// Banner generated at https://patorjk.com/software/taag/#p=display&v=3&f=Ivrit&t=NetScaler%20ADC%20-%20Backup
var banner = "  _   _      _   ____            _                _    ____   ____           ____             _                \n | \\ | | ___| |_/ ___|  ___ __ _| | ___ _ __     / \\  |  _ \\ / ___|         | __ )  __ _  ___| | ___   _ _ __  \n |  \\| |/ _ \\ __\\___ \\ / __/ _` | |/ _ \\ '__|   / _ \\ | | | | |      _____  |  _ \\ / _` |/ __| |/ / | | | '_ \\ \n | |\\  |  __/ |_ ___) | (_| (_| | |  __/ |     / ___ \\| |_| | |___  |_____| | |_) | (_| | (__|   <| |_| | |_) |\n |_| \\_|\\___|\\__|____/ \\___\\__,_|_|\\___|_|    /_/   \\_\\____/ \\____|         |____/ \\__,_|\\___|_|\\_\\\\__,_| .__/ \n                                                                                                        |_| "
var c config.Application

func main() {
	fmt.Println(banner)

	// Setup Viper
	viper.SetConfigName("netscaleradc-backup")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/corelayer/netscaleradc-backup/")
	viper.AddConfigPath("$HOME/.netscaleradc-backup")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	fmt.Println(c)

	cmd.Execute(c)
}
