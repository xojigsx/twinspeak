package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"jig.sx/twinspeak/srv"
)

var (
	addr string
)

var rootCmd = &cobra.Command{
	Use:   "twinspeak",
	Short: "Twinspeak - A drop-in replacement for Google's Gemini Live API",
	Long:  `Twinspeak provides real-time conversational AI capabilities over WebSocket connections, compatible with Google's Gemini Live API.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := srv.New()

		fmt.Printf("Starting Twinspeak server on %s\n", addr)
		log.Printf("Server listening on %s", addr)

		if err := http.ListenAndServe(addr, server.Handler()); err != nil {
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
