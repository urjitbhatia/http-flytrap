package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/urjitbhatia/http-flytrap/internal"
)

var capturePort = "9000"
var queryPort = "9001"
var ttl time.Duration

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flytrap",
	Short: "Flytrap for http requests",
	Long: `Flytrap captures http requests that are sent to it and stores for a some time.
You can ask flytrap what requests it has captured so far, however its stickiness decays over time.
Any path that hasn't seen a request for more than a TTL duration, will forget the requests it saw previously.
(This TTL can be configured)`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		internal.Trap(capturePort, queryPort, ttl)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&capturePort, "capturePort", "c", "9000", "capture port - all requests to this endpoint are captured")
	rootCmd.PersistentFlags().StringVarP(&queryPort, "queryPort", "q", "9001", "query interface port")
	rootCmd.PersistentFlags().DurationVarP(&ttl, "ttl", "t", time.Minute*30, "Time to remember captured requests (use go time.duration format. Eg: 10m)")
}
