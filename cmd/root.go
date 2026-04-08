package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/config"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version, commit, and date are set via ldflags at build time.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// NewRootCmd creates and returns the root command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bunny",
		Short:   "Bunny CLI — manage bunny.net resources from the command line",
		Version: version,
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("bunny-cli {{.Version}} (commit: %s, built: %s)\n", commit, date))
	rootCmd.SilenceErrors = true

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		RefreshSkillsIfVersionChanged()

		cfg := &output.Config{}

		// Parse --field flag
		fieldStr, _ := cmd.Flags().GetString("field")
		if fieldStr != "" {
			fields := strings.Split(fieldStr, ",")
			for i, f := range fields {
				fields[i] = strings.TrimSpace(f)
			}
			cfg.Fields = fields
		}

		// Parse --jq flag
		jqExpr, _ := cmd.Flags().GetString("jq")
		if jqExpr != "" {
			outputChanged := cmd.Flags().Changed("output")
			outputFormat, _ := cmd.Flags().GetString("output")

			if outputChanged && outputFormat == "table" {
				return fmt.Errorf("--jq and --output table are mutually exclusive")
			}

			if !outputChanged {
				viper.Set("output", "json")
			}

			query, err := gojq.Parse(jqExpr)
			if err != nil {
				return fmt.Errorf("invalid jq expression: %w", err)
			}

			code, err := gojq.Compile(query)
			if err != nil {
				return fmt.Errorf("compiling jq expression: %w", err)
			}

			cfg.JQ = code
		}

		// Set format from viper (flag or config file)
		cfg.Format = viper.GetString("output")

		// Store config on command context
		cmd.SetContext(output.NewContext(cmd.Context(), cfg))

		if err := config.Init(); err != nil {
			return err
		}

		return nil
	}

	rootCmd.PersistentFlags().String("api-key", "", "bunny.net API key")
	rootCmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression (built-in, no external jq required)")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table, json, json-pretty)")
	rootCmd.PersistentFlags().StringP("field", "f", "", "Comma-separated list of fields to display")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	_ = viper.BindEnv("api_key", "BUNNY_API_KEY")

	rootCmd.AddCommand(newConfigureCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newPullZonesCmd())
	rootCmd.AddCommand(newStorageZonesCmd())
	rootCmd.AddCommand(newStorageCmd())
	rootCmd.AddCommand(newDnsCmd())
	rootCmd.AddCommand(newStreamCmd())
	rootCmd.AddCommand(newScriptsCmd())
	rootCmd.AddCommand(newShieldCmd())
	rootCmd.AddCommand(newAccountCmd())
	rootCmd.AddCommand(newBillingCmd())
	rootCmd.AddCommand(newStatisticsCmd())
	rootCmd.AddCommand(newRegionsCmd())
	rootCmd.AddCommand(newCountriesCmd())
	rootCmd.AddCommand(newSkillCmd())
	rootCmd.AddCommand(newCompletionCmd())

	// Store the default App on the root command's context so all subcommands
	// can retrieve it. This must happen before PersistentPreRunE runs.
	rootCmd.SetContext(NewAppContext(context.Background(), DefaultApp()))

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), client.FormatError(err))
		os.Exit(1)
	}
}
