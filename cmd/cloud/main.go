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
	"os"
	"os/signal"
	"syscall"

	"github.com/rewardenv/reward-cloud-cli/cmd/root"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
)

var (
	APPNAME       = "cloud"
	PARENTAPPNAME = "reward"
	VERSION       = "v0.0.1"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)

	app := config.New(APPNAME, PARENTAPPNAME, VERSION)

	cobra.OnInitialize(func() {
		app.Init()
	})

	go func() {
		<-sig

		if err := app.Cleanup(); err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	err := root.NewCmdRoot(app).Execute()
	if err != nil {
		log.Error(err)
	}
}
