/*
Copyright The Ratify Authors.
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

package cmd

import (
	"context"

	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/httpserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	serveUse = "serve"
)

type ServeCmdOptions struct {
	configFilePath    string
	HttpServerAddress string
}

func NewCmdServe(argv ...string) *cobra.Command {

	var opts ServeCmdOptions

	cmd := &cobra.Command{
		Use:     serveUse,
		Short:   "Run ratify as a server",
		Example: "ratify server",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Serve(opts)

		},
	}

	flags := cmd.Flags()

	flags.StringVar(&opts.HttpServerAddress, "http", "", "HTTP Address")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	return cmd
}

func Serve(opts ServeCmdOptions) error {

	getExecutor, err := config.GetExecutorAndWatchForUpdate(opts.configFilePath)
	if err != nil {
		return err
	}

	if opts.HttpServerAddress != "" {
		server, err := httpserver.NewServer(context.Background(), opts.HttpServerAddress, getExecutor)
		if err != nil {
			return err
		}
		logrus.Infof("starting server at" + opts.HttpServerAddress)
		if err := server.Run(); err != nil {
			return err
		}
	}

	return nil
}
