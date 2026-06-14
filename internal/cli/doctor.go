package cli

import (
	"fmt"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/output"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/spf13/cobra"
)

func NewDoctorCmd() *cobra.Command {
	var (
		category []string
		check    []string
		severity string
		fix      bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run health diagnostics on your system",
		Long:  `Execute health checks across all components and display the health status of your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			categories := make([]core.Component, len(category))
			for i, c := range category {
				categories[i] = core.Component(c)
			}
			if len(categories) == 0 {
				categories = core.AllComponents()
			}

			if len(check) > 0 {
				checks := plugin.List(plugin.WithIDs(check...))
				if len(checks) == 0 {
					fmt.Println("No checks found for specified IDs")
					return nil
				}
				// Extract categories from filtered checks
				catSet := make(map[core.Component]bool)
				for _, ch := range checks {
					catSet[ch.Category()] = true
				}
				categories = make([]core.Component, 0, len(catSet))
				for c := range catSet {
					categories = append(categories, c)
				}
			}

			execCtx, err := plugin.NewExecutionContext(cmd.Context(), categories)
			if err != nil {
				return fmt.Errorf("execution context: %w", err)
			}
			if len(check) > 0 {
				execCtx = plugin.NewFilteredExecutionContext(execCtx, check)
			}

			go func() {
				for p := range execCtx.Progress() {
					fmt.Printf("\r[%d/%d] %-40s %s", p.Completed, p.Total, p.Current, p.Status)
				}
				fmt.Println()
			}()

			result := execCtx.Run()

			printer := output.NewPrinter(true)
			if c := configFromContext(cmd.Context()); c != nil {
				printer.Verbose = c.Global.Verbose
			}
			printer.PrintResult(result)

			if fix {
				for _, r := range result.AllResults {
					for _, rem := range r.Remediation {
						if rem.AutoFixable {
							fmt.Printf("  Auto-fixing: %s\n", rem.Action)
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&category, "category", nil, "categories to check (cpu, memory, disk, network, etc.)")
	cmd.Flags().StringSliceVar(&check, "check", nil, "specific checks to run")
	cmd.Flags().StringVar(&severity, "severity", "warning", "minimum severity level")
	cmd.Flags().BoolVar(&fix, "fix", false, "attempt auto-fix for remediable issues")

	return cmd
}
