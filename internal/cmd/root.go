package cmd

import (
	"context"

	assets "github.com/danielmichaels/tawny"
	"github.com/spf13/cobra"
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

//func cmdLogger() *zerolog.Logger {
//	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
//	return &logger
//}
