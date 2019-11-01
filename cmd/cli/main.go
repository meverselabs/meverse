package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var hostURL string
	var rootCmd = &cobra.Command{Use: "cli"}
	rootCmd.PersistentFlags().StringVar(&hostURL, "host", "http://localhost:58000", "url of the node to access")
	rootCmd.AddCommand(keyCommand(&hostURL))
	rootCmd.AddCommand(accountCommand(&hostURL))
	rootCmd.AddCommand(txCommand(&hostURL))
	rootCmd.AddCommand(chainCommand(&hostURL))
	rootCmd.Execute()
}
