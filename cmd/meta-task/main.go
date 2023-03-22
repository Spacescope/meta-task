package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Spacescore/observatory-task/internal/busi"

	"github.com/spf13/cobra"
)

// @title spacescope data extraction notify backend
// @version 1.0
// @description spacescope data extraction api backend
// @termsOfService http://swagger.io/terms/

// @host meta-task-api.spacescope.io
// @BasePath /

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
	busi.NewServer(context.Background()).Start()
	return nil
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
