package cli

import (
	"fmt"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/baseline"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/spf13/cobra"
)

func NewBaselineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baseline",
		Short: "Manage system baselines",
		Long:  `Capture, list, compare, and manage system state baselines for health comparison.`,
	}

	cmd.AddCommand(NewBaselineCaptureCmd())
	cmd.AddCommand(NewBaselineListCmd())
	cmd.AddCommand(NewBaselineShowCmd())
	cmd.AddCommand(NewBaselineCompareCmd())
	cmd.AddCommand(NewBaselineDiffCmd())
	cmd.AddCommand(NewBaselineDeleteCmd())
	cmd.AddCommand(NewBaselineScheduleCmd())

	return cmd
}

func NewBaselineCaptureCmd() *cobra.Command {
	var (
		name        string
		description string
	)

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a new system baseline",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := baseline.NewEngine(nil)

			fmt.Print("Capturing system state... ")
			b, err := engine.Capture(cmd.Context())
			if err != nil {
				return fmt.Errorf("capture failed: %w", err)
			}
			fmt.Println("done")

			b.Name = name
			b.Description = description
			fmt.Printf("Baseline captured: %s (%s)\n", b.ID, b.Timestamp.Format(time.RFC3339))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "baseline name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "baseline description")
	return cmd
}

func NewBaselineListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all baselines",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available baselines:")
			fmt.Println("  (baseline storage not yet connected)")
			return nil
		},
	}
}

func NewBaselineShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [baseline-id]",
		Short: "Show baseline details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Baseline: %s\n", args[0])
			return nil
		},
	}
}

func NewBaselineCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare [baseline-id]",
		Short: "Compare current state against a baseline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCtx, err := plugin.NewExecutionContext(cmd.Context(), core.AllComponents())
			if err != nil {
				return fmt.Errorf("execution context: %w", err)
			}
			result := execCtx.Run()

			engine := baseline.NewEngine(nil)
			base := &baseline.Baseline{ID: args[0]}

			comp, err := engine.Compare(cmd.Context(), base, result)
			if err != nil {
				return fmt.Errorf("comparison failed: %w", err)
			}

			fmt.Printf("Comparison vs %s:\n", args[0])
			fmt.Printf("  Health Score: %.1f/100\n", comp.HealthScore)
			fmt.Printf("  Deviations: %d\n", len(comp.Deviations))
			for _, d := range comp.Deviations {
				fmt.Printf("  [%s] %s: expected %.2f, got %.2f (%.1f%% delta)\n",
					d.Severity, d.Metric, d.Expected, d.Actual, d.Delta*100)
			}
			return nil
		},
	}
}

func NewBaselineDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [baseline1] [baseline2]",
		Short: "Show differences between two baselines",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Diff between %s and %s (not yet implemented)\n", args[0], args[1])
			return nil
		},
	}
}

func NewBaselineDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [baseline-id]",
		Short: "Delete a baseline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Deleting baseline: %s\n", args[0])
			return nil
		},
	}
}

func NewBaselineScheduleCmd() *cobra.Command {
	var (
		cronExpr string
		remove   bool
	)

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Schedule automatic baseline captures",
		RunE: func(cmd *cobra.Command, args []string) error {
			if remove {
				fmt.Println("Scheduled baseline capture removed")
				return nil
			}
			if cronExpr != "" {
				fmt.Printf("Baseline capture scheduled: %s\n", cronExpr)
			} else {
				fmt.Println("Current schedule: not set")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&cronExpr, "cron", "c", "", "cron expression for scheduling")
	cmd.Flags().BoolVar(&remove, "remove", false, "remove scheduled capture")
	return cmd
}
