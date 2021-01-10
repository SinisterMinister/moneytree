/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/moneytree/pkg/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// placePairCmd represents the placePair command
var placePairCmd = &cobra.Command{
	Use:   "placePair [DIRECTION]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"up", "down"},
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.WithError(err).Fatal("could not get port")
		}
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			log.WithError(err).Fatal("could not get port")
		}
		timeout, err := cmd.Flags().GetString("timeout")
		if err != nil {
			log.WithError(err).Fatal("could not get timeout")
		}
		address := fmt.Sprintf("%s:%d", host, port)

		fmt.Println(fmt.Sprintf("placePair called with %s:%d", host, port))

		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := proto.NewMoneytreeClient(conn)

		// Contact the server and print out its response.
		to, err := time.ParseDuration(timeout)
		if err != nil {
			log.WithError(err).Fatal("could not parse timeout value")
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()
		r, err := c.PlacePair(ctx, &proto.PlacePairRequest{Direction: strings.ToUpper(args[0])})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Infof("Pair: %s", r.GetPair().Uuid)
	},
}

func init() {
	clientCmd.AddCommand(placePairCmd)
	placePairCmd.Flags().String("host", "localhost", "Host to connect to")
	placePairCmd.Flags().Int("port", 44444, "Port to connect to")
	placePairCmd.Flags().String("timeout", "15s", "Timeout")
}
