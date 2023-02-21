package main

import (
	"fmt"
	"os"

	"github.com/Spacescore/observatory-task/internal/busi"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta-task",
		Short: "mt",
		Run: func(cmd *cobra.Command, args []string) {
			if err := entry(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&busi.Flags.Config, "conf", "", "path of the configuration file")

	return cmd
}

func entry() error {
	return busi.Start()
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
