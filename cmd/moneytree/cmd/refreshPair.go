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
	"context"
	"fmt"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/moneytree/pkg/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// refreshPairCmd represents the refreshPair command
var refreshPairCmd = &cobra.Command{
	Use:   "refreshPair",
	Short: "A brief description of your command",
	Long:  `Refreshes the orders in an order pair`,
	Args:  cobra.ExactArgs(1),
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

		uuid := args[0]
		fmt.Println(fmt.Sprintf("attempting to refresh pair %s", uuid))

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
		r, err := c.RefreshPair(ctx, &proto.PairRequest{Uuid: uuid})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Infof("pair %s was refreshed", r.Uuid)
	},
}

func init() {
	clientCmd.AddCommand(refreshPairCmd)
	refreshPairCmd.Flags().String("host", "localhost", "Host to connect to")
	refreshPairCmd.Flags().Int("port", 44444, "Port to connect to")
	refreshPairCmd.Flags().String("timeout", "15s", "Timeout")
}
