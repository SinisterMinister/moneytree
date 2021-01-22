/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/miraclegrow/pkg/grow"
	"github.com/spf13/cobra"
)

// growCmd represents the grow command
var growCmd = &cobra.Command{
	Use:   "grow",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.WithError(err).Fatal("could not get port")
		}
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			log.WithError(err).Fatal("could not get port")
		}
		updateFrequencyRaw, err := cmd.Flags().GetString("updateFrequency")
		if err != nil {
			log.WithError(err).Fatal("could not get update frequency")
		}
		updateFrequency, err := time.ParseDuration(updateFrequencyRaw)
		if err != nil {
			log.WithError(err).Fatal("could not parse update frequency")
		}
		address := fmt.Sprintf("%s:%d", host, port)

		svc := grow.NewService(address, updateFrequency)

		svc.Grow(make(chan bool))
	},
}

func init() {
	rootCmd.AddCommand(growCmd)

	growCmd.Flags().String("host", "moneytree.sinimini.com", "Host to connect to")
	growCmd.Flags().Int("port", 44444, "Port to connect to")
	growCmd.Flags().String("updateFrequency", "5s", "Timeout")
}
