package cli

import (
	"fmt"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/knowledge"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/spf13/cobra"
)

func NewDiagnoseCmd() *cobra.Command {
	var symptom []string

	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Deep root-cause analysis of system issues",
		Long:  `Run doctor checks combined with knowledge engine analysis to identify root causes and provide recommendations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			execCtx, err := plugin.NewExecutionContext(cmd.Context(), core.AllComponents())
			if err != nil {
				return fmt.Errorf("execution context: %w", err)
			}

			go func() {
				for p := range execCtx.Progress() {
					fmt.Printf("\r[%d/%d] %-40s %s", p.Completed, p.Total, p.Current, p.Status)
				}
				fmt.Println()
			}()

			result := execCtx.Run()

			engine := knowledge.NewEngine()
			engine.AddRules(knowledge.BuiltinCPURules())
			engine.AddRules(knowledge.BuiltinMemoryRules())
			engine.AddRules(knowledge.BuiltinDiskRules())
			engine.AddRules(knowledge.BuiltinNetworkRules())

			if err := engine.IngestFacts(cmd.Context(), result.AllResults); err != nil {
				return fmt.Errorf("ingesting facts: %w", err)
			}

			analysis, err := engine.Analyze(cmd.Context())
			if err != nil {
				return fmt.Errorf("analysis: %w", err)
			}

			fmt.Println()
			fmt.Println("=== Diagnosis Results ===")
			fmt.Printf("Overall Severity: %s\n", analysis.OverallSeverity)

			if analysis.RootCause != nil {
				fmt.Printf("\nRoot Cause: %s\n", analysis.RootCause.Title)
				fmt.Printf("  %s\n", analysis.RootCause.Description)
				fmt.Printf("  Certainty: %.0f%%\n", analysis.RootCause.Certainty*100)
			}

			if len(analysis.Conclusions) > 0 {
				fmt.Println("\nFindings:")
				for _, c := range analysis.Conclusions {
					fmt.Printf("  [%s] %s (%.0f%% certainty)\n", c.Severity, c.Title, c.Certainty*100)
					if len(c.Remediations) > 0 {
						for _, r := range c.Remediations {
							fmt.Printf("    → %s\n", r.Action)
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&symptom, "symptom", nil, "specific symptoms to analyze")

	return cmd
}
