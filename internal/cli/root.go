package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/linuxhealthdoctor/lhd/internal/config"
	"github.com/linuxhealthdoctor/lhd/pkg/version"
	"github.com/spf13/cobra"
)

type contextKey string

const cfgKey contextKey = "config"

func contextWithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, cfgKey, cfg)
}

func configFromContext(ctx context.Context) *config.Config {
	cfg, _ := ctx.Value(cfgKey).(*config.Config)
	return cfg
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lhd",
		Short: "Linux Health Doctor - Your Linux system's primary care physician",
		Long:  `Linux Health Doctor (lhd) is a local-first, offline-first health diagnostics and root-cause analysis platform for Linux systems.`,
		Version: version.Info(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			verbose, _ := cmd.Flags().GetBool("verbose")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
				}
				cfg = config.Default()
			}
			cmd.SetContext(contextWithConfig(cmd.Context(), cfg))
			return nil
		},
	}

	cmd.PersistentFlags().String("config", "", "config file path")
	cmd.PersistentFlags().String("format", "ansi", "output format (ansi, json, yaml, csv, markdown)")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	cmd.PersistentFlags().String("output", "", "output file path")

	cmd.AddCommand(NewDoctorCmd())
	cmd.AddCommand(NewDiagnoseCmd())
	cmd.AddCommand(NewMonitorCmd())
	cmd.AddCommand(NewDashboardCmd())
	cmd.AddCommand(NewReportCmd())
	cmd.AddCommand(NewBaselineCmd())
	cmd.AddCommand(NewSnapshotCmd())
	cmd.AddCommand(NewFleetCmd())
	cmd.AddCommand(NewConfigCmd())
	cmd.AddCommand(NewCompletionCmd())

	return cmd
}
