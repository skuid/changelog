package cmd

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skuid/changelog/webhooks/github"
	"github.com/skuid/spec"
	"github.com/skuid/spec/lifecycle"
	_ "github.com/skuid/spec/metrics"
	"github.com/skuid/spec/middlewares"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a webhook endpoint for PR validation",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		l, _ := spec.NewStandardLogger()
		zap.ReplaceGlobals(l)
		var webhookHandler http.Handler

		switch viper.GetString("provider") {
		case "github":
			webhookHandler = github.New(viper.GetString("secret"), viper.GetString("token"))
		default:
			zap.L().Fatal(
				fmt.Sprintf("webhook for provider %s isn't supported", viper.GetString("provider")),
				zap.String("provider", viper.GetString("provider")),
			)
		}

		handler := middlewares.Apply(
			webhookHandler,
			middlewares.InstrumentRoute(),
			middlewares.Logging(),
		)

		internalMux := http.NewServeMux()
		internalMux.Handle("/", handler)
		internalMux.Handle("/metrics", promhttp.Handler())
		internalMux.HandleFunc("/live", lifecycle.LivenessHandler)
		internalMux.HandleFunc("/ready", lifecycle.ReadinessHandler)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", viper.GetInt("port")),
			Handler: internalMux,
		}
		lifecycle.ShutdownOnTerm(server)

		zap.L().Info("starting changelog webhook server", zap.Int("port", viper.GetInt("port")))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal(err.Error())
		}

	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("secret", "s", "", "webhook secret")
	serveCmd.Flags().IntP("port", "n", 3000, "webhook server port")
	viper.BindPFlags(serveCmd.Flags())
}
