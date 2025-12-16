package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "tmpo",
	Short: "Minimal CLI time tracker for developers",
	Long: `tmpo - Set the tmpo

A minimal, developer-friendly time tracking tool that lives in your terminal.
Track time effortlessly with automatic project detection and simple commands.`,
	Version: Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(GetVersionOutput())
}
