package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newScriptsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scripts",
		Aliases: []string{"compute"},
		Short:   "Manage edge scripts",
	}
	cmd.AddCommand(newScriptsListCmd())
	cmd.AddCommand(withWatch(newScriptsGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newScriptsCreateCmd())))
	cmd.AddCommand(withFromFile(newScriptsUpdateCmd()))
	cmd.AddCommand(newScriptsDeleteCmd())
	cmd.AddCommand(newScriptsStatisticsCmd())
	cmd.AddCommand(newScriptsRotateKeyCmd())
	cmd.AddCommand(newScriptsCodeCmd())
	cmd.AddCommand(newScriptsPublishCmd())
	cmd.AddCommand(newScriptsReleasesCmd())
	cmd.AddCommand(newScriptsVariablesCmd())
	cmd.AddCommand(newScriptsSecretsCmd())
	return cmd
}

func newScriptsListCmd() *cobra.Command {
	var (
		limit      int
		all        bool
		search     string
		scriptType string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List edge scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			var scriptTypes []int
			if cmd.Flags().Changed("type") {
				t, err := client.ScriptTypeFromName(scriptType)
				if err != nil {
					return err
				}
				scriptTypes = []int{t}
			}

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.EdgeScript], error) {
				return c.ListEdgeScripts(cmd.Context(), page, perPage, search, scriptTypes)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := edgeScriptListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, s := range result.Items {
				items[i] = s
			}

			formatted, err := output.FormatList(cfg, columns, items, result.HasMore)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results to return (default 20)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages of results")
	cmd.Flags().StringVar(&search, "search", "", "Filter results by search term")
	cmd.Flags().StringVar(&scriptType, "type", "", "Filter by script type (DNS, CDN, Middleware)")

	return cmd
}

func newScriptsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get edge script details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			s, err := c.GetEdgeScript(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := edgeScriptDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, s)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newScriptsCreateCmd() *cobra.Command {
	var (
		name       string
		scriptType string
		code       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an edge script",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.EdgeScriptCreate{}

			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("type") {
				t, err := client.ScriptTypeFromName(scriptType)
				if err != nil {
					return err
				}
				body.ScriptType = t
			}
			if cmd.Flags().Changed("code") {
				body.Code = code
			}

			s, err := c.CreateEdgeScript(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := edgeScriptDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, s)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Script name (required)")
	cmd.Flags().StringVar(&scriptType, "type", "", "Script type: DNS, CDN, Middleware (required)")
	cmd.Flags().StringVar(&code, "code", "", "Initial script code")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func newScriptsUpdateCmd() *cobra.Command {
	var (
		name       string
		scriptType string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an edge script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.EdgeScriptUpdate{}

			if cmd.Flags().Changed("name") {
				body.Name = &name
			}
			if cmd.Flags().Changed("type") {
				t, err := client.ScriptTypeFromName(scriptType)
				if err != nil {
					return err
				}
				body.ScriptType = &t
			}

			s, err := c.UpdateEdgeScript(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := edgeScriptDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, s)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Script name")
	cmd.Flags().StringVar(&scriptType, "type", "", "Script type: DNS, CDN, Middleware")

	return cmd
}

func newScriptsDeleteCmd() *cobra.Command {
	var (
		yes                   bool
		deleteLinkedPullZones bool
	)

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an edge script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete edge script %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteEdgeScript(cmd.Context(), id, deleteLinkedPullZones); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge script deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&deleteLinkedPullZones, "delete-linked-pullzones", false, "Also delete linked pull zones")

	return cmd
}

func newScriptsStatisticsCmd() *cobra.Command {
	var (
		dateFrom   string
		dateTo     string
		hourly     bool
		loadLatest bool
	)

	cmd := &cobra.Command{
		Use:   "statistics <id>",
		Short: "Get edge script statistics",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			stats, err := c.GetEdgeScriptStatistics(cmd.Context(), id, dateFrom, dateTo, loadLatest, hourly)
			if err != nil {
				return err
			}

			columns := edgeScriptStatisticsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, stats)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&dateFrom, "date-from", "", "Start date for statistics (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "date-to", "", "End date for statistics (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&hourly, "hourly", false, "Return statistics in hourly grouping")
	cmd.Flags().BoolVar(&loadLatest, "load-latest", false, "Load most recent data as soon as available")

	return cmd
}

func newScriptsRotateKeyCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "rotate-key <id>",
		Short: "Rotate edge script deployment key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to rotate the deployment key for edge script %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Key rotation canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.RotateEdgeScriptDeploymentKey(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deployment key rotated.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Column definitions ---

func edgeScriptListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScript]{
		output.IntColumn[*client.EdgeScript]("Id", func(s *client.EdgeScript) int { return int(s.Id) }),
		output.StringColumn[*client.EdgeScript]("Name", func(s *client.EdgeScript) string { return s.Name }),
		output.StringColumn[*client.EdgeScript]("Type", func(s *client.EdgeScript) string { return client.ScriptTypeName(s.ScriptType) }),
		output.IntColumn[*client.EdgeScript]("Variables", func(s *client.EdgeScript) int { return len(s.EdgeScriptVariables) }),
		output.IntColumn[*client.EdgeScript]("Pull Zones", func(s *client.EdgeScript) int { return len(s.LinkedPullZones) }),
		output.FloatColumn[*client.EdgeScript]("Monthly Cost", func(s *client.EdgeScript) float64 { return s.MonthlyCost }),
	})
}

func edgeScriptDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScript]{
		output.IntColumn[*client.EdgeScript]("Id", func(s *client.EdgeScript) int { return int(s.Id) }),
		output.StringColumn[*client.EdgeScript]("Name", func(s *client.EdgeScript) string { return s.Name }),
		output.StringColumn[*client.EdgeScript]("Type", func(s *client.EdgeScript) string { return client.ScriptTypeName(s.ScriptType) }),
		output.IntColumn[*client.EdgeScript]("Current Release Id", func(s *client.EdgeScript) int { return int(s.CurrentReleaseId) }),
		output.StringColumn[*client.EdgeScript]("Default Hostname", func(s *client.EdgeScript) string { return s.DefaultHostname }),
		output.StringColumn[*client.EdgeScript]("System Hostname", func(s *client.EdgeScript) string { return s.SystemHostname }),
		output.StringColumn[*client.EdgeScript]("Deployment Key", func(s *client.EdgeScript) string { return s.DeploymentKey }),
		output.IntColumn[*client.EdgeScript]("Variables", func(s *client.EdgeScript) int { return len(s.EdgeScriptVariables) }),
		output.IntColumn[*client.EdgeScript]("Pull Zones", func(s *client.EdgeScript) int { return len(s.LinkedPullZones) }),
		output.FloatColumn[*client.EdgeScript]("Monthly Cost", func(s *client.EdgeScript) float64 { return s.MonthlyCost }),
		output.IntColumn[*client.EdgeScript]("Monthly Requests", func(s *client.EdgeScript) int { return int(s.MonthlyRequestCount) }),
		output.IntColumn[*client.EdgeScript]("Monthly CPU Time", func(s *client.EdgeScript) int { return int(s.MonthlyCpuTime) }),
		output.StringColumn[*client.EdgeScript]("Last Modified", func(s *client.EdgeScript) string { return s.LastModified }),
	})
}

func edgeScriptStatisticsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptStatistics]{
		output.IntColumn[*client.EdgeScriptStatistics]("Total Requests Served", func(s *client.EdgeScriptStatistics) int { return int(s.TotalRequestsServed) }),
		output.FloatColumn[*client.EdgeScriptStatistics]("Total CPU Used", func(s *client.EdgeScriptStatistics) float64 { return s.TotalCpuUsed }),
		output.FloatColumn[*client.EdgeScriptStatistics]("Total Monthly Cost", func(s *client.EdgeScriptStatistics) float64 { return s.TotalMonthlyCost }),
		output.FloatColumn[*client.EdgeScriptStatistics]("Avg CPU Time/Execution", func(s *client.EdgeScriptStatistics) float64 { return s.AverageCpuTimePerExecution }),
	})
}
