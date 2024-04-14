package cmd

import (
	"context"
	assets "github.com/danielmichaels/tawny"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

func Execute(ctx context.Context) int {
	rootCmd := &cobra.Command{
		Use:   assets.AppName,
		Short: "",
	}

	rootCmd.AddCommand(ServeCmd(ctx))

	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}
