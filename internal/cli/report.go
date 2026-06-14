package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/report"
	"github.com/spf13/cobra"
)

func NewReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate and manage reports",
		Long:  `Generate health reports in various formats from check results, baselines, or snapshots.`,
	}

	cmd.AddCommand(NewReportGenerateCmd())
	cmd.AddCommand(NewReportCompareCmd())
	cmd.AddCommand(NewReportListCmd())
	cmd.AddCommand(NewReportViewCmd())
	cmd.AddCommand(NewReportExportCmd())

	return cmd
}

func NewReportGenerateCmd() *cobra.Command {
	var (
		from       string
		format     string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a health report",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := report.NewEngine()
			engine.RegisterFormatter("ansi", &report.ANSITerminalFormatter{Color: true})
			engine.RegisterFormatter("json", &report.JSONFormatter{Pretty: true})

			data := &report.ReportData{
				Title:       "Linux Health Doctor Report",
				GeneratedAt: time.Now(),
				Hostname:    "localhost",
				Distro:      "unknown",
				Kernel:      "unknown",
			}

			opts := &report.FormatOptions{Color: true}

			output, err := engine.Generate(context.Background(), data, format, opts)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, output, 0644); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Printf("Report saved to %s\n", outputFile)
			} else {
				fmt.Println(string(output))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "generate from baseline or snapshot (baseline:<id>, snapshot:<id>)")
	cmd.Flags().StringVarP(&format, "format", "f", "ansi", "output format (ansi, json, yaml, csv, markdown, html)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")

	return cmd
}

func NewReportCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare [ref1] [ref2]",
		Short: "Compare two baselines or snapshots",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Comparison between %s and %s (not yet implemented)\n", args[0], args[1])
			return nil
		},
	}
}

func NewReportListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List generated reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Generated reports:")
			fmt.Println("  No reports found")
			return nil
		},
	}
}

func NewReportViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view [report-id]",
		Short: "View a report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Viewing report: %s (not yet implemented)\n", args[0])
			return nil
		},
	}
}

func NewReportExportCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export [report-id]",
		Short: "Export a report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Exporting report %s as %s (not yet implemented)\n", args[0], format)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "pdf", "export format (pdf, html, markdown)")
	return cmd
}
