// Package main provides the Twinspeak server executable.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"jig.sx/twinspeak/srv"
)

var (
	addr string
)

var rootCmd = &cobra.Command{
	Use:   "twinspeak",
	Short: "Twinspeak - Real-time conversational AI over WebSocket connections",
	Long: `Twinspeak provides real-time conversational AI capabilities over WebSocket connections ` +
		`with support for text and audio communication.`,
	Run: func(_ *cobra.Command, _ []string) {
		server := srv.New()

		fmt.Printf("Starting Twinspeak server on %s\n", addr)
		log.Printf("Server listening on %s", addr)

		httpServer := &http.Server{
			Addr:              addr,
			Handler:           server.Handler(),
			ReadHeaderTimeout: 30 * time.Second,
		}
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().StringVar(&addr, "addr", ":8080", "Address to listen on (default :8080)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
