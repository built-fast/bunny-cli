package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newShieldRateLimitsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rate-limits",
		Aliases: []string{"rate-limit", "ratelimits"},
		Short:   "Manage Shield rate limit rules",
	}
	cmd.AddCommand(newShieldRateLimitsListCmd())
	cmd.AddCommand(withWatch(newShieldRateLimitsGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newShieldRateLimitsCreateCmd())))
	cmd.AddCommand(withFromFile(newShieldRateLimitsUpdateCmd()))
	cmd.AddCommand(newShieldRateLimitsDeleteCmd())
	return cmd
}

func newShieldRateLimitsListCmd() *cobra.Command {
	var (
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list <shield_zone_id>",
		Short: "List rate limit rules for a Shield zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.RateLimitRule], error) {
				return c.ListRateLimits(cmd.Context(), shieldZoneId, page, perPage)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := rateLimitRuleListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, r := range result.Items {
				items[i] = r
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

	return cmd
}

func newShieldRateLimitsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get rate limit rule details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid rate limit rule ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			rule, err := c.GetRateLimit(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := rateLimitRuleDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, rule)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldRateLimitsCreateCmd() *cobra.Command {
	var (
		shieldZoneId int64
		ruleName     string
		ruleDesc     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a rate limit rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.RateLimitRuleCreate{}
			if cmd.Flags().Changed("shield-zone-id") {
				body.ShieldZoneId = shieldZoneId
			}
			if cmd.Flags().Changed("rule-name") {
				body.RuleName = ruleName
			}
			if cmd.Flags().Changed("rule-description") {
				body.RuleDescription = ruleDesc
			}

			rule, err := c.CreateRateLimit(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := rateLimitRuleDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, rule)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().Int64Var(&shieldZoneId, "shield-zone-id", 0, "Shield Zone ID (required)")
	cmd.Flags().StringVar(&ruleName, "rule-name", "", "Rule name (required)")
	cmd.Flags().StringVar(&ruleDesc, "rule-description", "", "Rule description")
	_ = cmd.MarkFlagRequired("shield-zone-id")
	_ = cmd.MarkFlagRequired("rule-name")

	return cmd
}

func newShieldRateLimitsUpdateCmd() *cobra.Command {
	var (
		ruleName string
		ruleDesc string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a rate limit rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid rate limit rule ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.RateLimitRuleUpdate{}
			if cmd.Flags().Changed("rule-name") {
				body.RuleName = &ruleName
			}
			if cmd.Flags().Changed("rule-description") {
				body.RuleDescription = &ruleDesc
			}

			rule, err := c.UpdateRateLimit(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := rateLimitRuleDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, rule)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&ruleName, "rule-name", "", "Rule name")
	cmd.Flags().StringVar(&ruleDesc, "rule-description", "", "Rule description")

	return cmd
}

func newShieldRateLimitsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a rate limit rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid rate limit rule ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete rate limit rule %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteRateLimit(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Rate limit rule deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Column definitions ---

func rateLimitRuleListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.RateLimitRule]{
		output.IntColumn[*client.RateLimitRule]("Id", func(r *client.RateLimitRule) int { return int(r.Id) }),
		output.IntColumn[*client.RateLimitRule]("Shield Zone Id", func(r *client.RateLimitRule) int { return int(r.ShieldZoneId) }),
		output.StringColumn[*client.RateLimitRule]("Rule Name", func(r *client.RateLimitRule) string { return r.RuleName }),
		output.StringColumn[*client.RateLimitRule]("Description", func(r *client.RateLimitRule) string { return r.RuleDescription }),
	})
}

func rateLimitRuleDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.RateLimitRule]{
		output.IntColumn[*client.RateLimitRule]("Id", func(r *client.RateLimitRule) int { return int(r.Id) }),
		output.IntColumn[*client.RateLimitRule]("Shield Zone Id", func(r *client.RateLimitRule) int { return int(r.ShieldZoneId) }),
		output.StringColumn[*client.RateLimitRule]("Rule Name", func(r *client.RateLimitRule) string { return r.RuleName }),
		output.StringColumn[*client.RateLimitRule]("Description", func(r *client.RateLimitRule) string { return r.RuleDescription }),
		output.StringColumn[*client.RateLimitRule]("Rule JSON", func(r *client.RateLimitRule) string { return r.RuleJson }),
	})
}
