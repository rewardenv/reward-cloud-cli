/*
Copyright © 2021-2023 JANOS MIKO <info@janosmiko.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"reward-cloud-cli/cmd/root"
	"reward-cloud-cli/internal/config"
)

var (
	APPNAME = "cloud"
	VERSION = "v0.0.1"
)

func main() {
	c := config.New(APPNAME, VERSION)

	cobra.OnInitialize(func() {
		c.Init()
	})

	err := root.NewCmdRoot(c).Execute()
	if err != nil {
		log.Error(err)
	}
}
