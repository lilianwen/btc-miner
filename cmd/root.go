package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "noded",
		Short: "noded is a node service for btc p2p network",
		Long:  `noded is a node service for btc p2p network`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}
