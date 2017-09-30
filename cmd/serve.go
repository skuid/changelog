// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/skuid/changelog/webhooks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a webhook endpoint for PR validation",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		var handler func(http.ResponseWriter, *http.Request)

		switch viper.GetString("provider") {
		case "github":
			handler = webhooks.GithubWebhook(viper.GetString("secret"), viper.GetString("token"))
		default:
			fmt.Errorf("webhook for provider %s isn't supported", viper.GetString("provider"))
			os.Exit(1)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/webhook", handler)
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", viper.GetInt("port")),
			Handler: mux,
		}

		fmt.Printf("Starting changelog webhook server on port %d", viper.GetInt("port"))
		if err := server.ListenAndServe(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("secret", "s", "", "webhook secret")
	serveCmd.Flags().IntP("port", "n", 3000, "webhook server port")
	viper.BindPFlags(serveCmd.Flags())
}
