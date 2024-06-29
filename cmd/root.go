package cmd

import (
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "chatgpt-codeview",
	Short: "use chatgpt to help review code and comment on github-specified pr",
	Long:  `use chatgpt to help review code and comment on github-specified pr`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func Execute() {
	_ = rootCmd.Execute()
}
