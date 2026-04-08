package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newPullZoneEdgeRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge-rules",
		Short: "Manage pull zone edge rules",
	}
	cmd.AddCommand(newPullZoneEdgeRulesListCmd())
	cmd.AddCommand(withFromFile(newPullZoneEdgeRulesAddCmd()))
	cmd.AddCommand(newPullZoneEdgeRulesDeleteCmd())
	cmd.AddCommand(newPullZoneEdgeRulesEnableCmd())
	cmd.AddCommand(newPullZoneEdgeRulesDisableCmd())
	return cmd
}

func newPullZoneEdgeRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <pull_zone_id>",
		Short: "List edge rules for a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			pz, err := c.GetPullZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := edgeRuleColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(pz.EdgeRules))
			for i := range pz.EdgeRules {
				items[i] = &pz.EdgeRules[i]
			}

			formatted, err := output.FormatList(cfg, columns, items, false)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newPullZoneEdgeRulesAddCmd() *cobra.Command {
	var (
		actionType          int
		actionParameter1    string
		actionParameter2    string
		triggerMatchingType int
		description         string
		enabled             bool
	)

	cmd := &cobra.Command{
		Use:   "add <pull_zone_id>",
		Short: "Add or update an edge rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			rule := &client.EdgeRule{}

			if cmd.Flags().Changed("action-type") {
				rule.ActionType = actionType
			}
			if cmd.Flags().Changed("action-parameter1") {
				rule.ActionParameter1 = actionParameter1
			}
			if cmd.Flags().Changed("action-parameter2") {
				rule.ActionParameter2 = actionParameter2
			}
			if cmd.Flags().Changed("trigger-matching-type") {
				rule.TriggerMatchingType = triggerMatchingType
			}
			if cmd.Flags().Changed("description") {
				rule.Description = description
			}
			if cmd.Flags().Changed("enabled") {
				rule.Enabled = enabled
			}

			if err := c.AddOrUpdateEdgeRule(cmd.Context(), id, rule); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge rule saved.")
			return err
		},
	}

	cmd.Flags().IntVar(&actionType, "action-type", 0, "Action type (0=ForceSSL, 1=Redirect, 2=OriginUrl, etc.)")
	cmd.Flags().StringVar(&actionParameter1, "action-parameter1", "", "First action parameter")
	cmd.Flags().StringVar(&actionParameter2, "action-parameter2", "", "Second action parameter")
	cmd.Flags().IntVar(&triggerMatchingType, "trigger-matching-type", 0, "Trigger matching (0=MatchAny, 1=MatchAll, 2=MatchNone)")
	cmd.Flags().StringVar(&description, "description", "", "Rule description")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the rule is enabled")

	return cmd
}

func newPullZoneEdgeRulesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <pull_zone_id> <edge_rule_id>",
		Short: "Delete an edge rule",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pullZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}
			edgeRuleId := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete edge rule %q from pull zone %d? [y/N] ", edgeRuleId, pullZoneId))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteEdgeRule(cmd.Context(), pullZoneId, edgeRuleId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge rule deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newPullZoneEdgeRulesEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <pull_zone_id> <edge_rule_id>",
		Short: "Enable an edge rule",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pullZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}
			edgeRuleId := args[1]

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.SetEdgeRuleEnabled(cmd.Context(), pullZoneId, edgeRuleId, true); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge rule enabled.")
			return err
		},
	}

	return cmd
}

func newPullZoneEdgeRulesDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <pull_zone_id> <edge_rule_id>",
		Short: "Disable an edge rule",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pullZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}
			edgeRuleId := args[1]

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.SetEdgeRuleEnabled(cmd.Context(), pullZoneId, edgeRuleId, false); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge rule disabled.")
			return err
		},
	}

	return cmd
}

func edgeRuleColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeRule]{
		output.StringColumn[*client.EdgeRule]("Guid", func(r *client.EdgeRule) string { return r.Guid }),
		output.StringColumn[*client.EdgeRule]("Description", func(r *client.EdgeRule) string { return r.Description }),
		output.IntColumn[*client.EdgeRule]("Action Type", func(r *client.EdgeRule) int { return r.ActionType }),
		output.StringColumn[*client.EdgeRule]("Parameter 1", func(r *client.EdgeRule) string { return r.ActionParameter1 }),
		output.BoolColumn[*client.EdgeRule]("Enabled", func(r *client.EdgeRule) bool { return r.Enabled }),
	})
}
