package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newShieldWafCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waf",
		Short: "Manage WAF rules and configuration",
	}
	cmd.AddCommand(newShieldWafRulesCmd())
	cmd.AddCommand(newShieldWafCustomRulesCmd())
	cmd.AddCommand(newShieldWafProfilesListCmd())
	cmd.AddCommand(newShieldWafEngineCmd())
	cmd.AddCommand(newShieldWafTriggeredCmd())
	return cmd
}

// --- Managed WAF Rules ---

func newShieldWafRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage WAF rules",
	}
	cmd.AddCommand(newShieldWafRulesListCmd())
	return cmd
}

func newShieldWafRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <shield_zone_id>",
		Short: "List managed WAF rules for a Shield zone",
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

			groups, err := c.ListWafRules(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			// Flatten rule groups into individual rules for display
			type flatRule struct {
				MainGroup   string
				GroupName   string
				GroupCode   string
				RuleId      int
				Description string
			}

			var rules []flatRule
			for _, mg := range groups {
				for _, rg := range mg.RuleGroups {
					for _, r := range rg.Rules {
						rules = append(rules, flatRule{
							MainGroup:   mg.Name,
							GroupName:   rg.Name,
							GroupCode:   rg.Code,
							RuleId:      r.RuleId,
							Description: r.Description,
						})
					}
				}
			}

			columns := output.ToColumns([]output.TypedColumn[*flatRule]{
				output.StringColumn[*flatRule]("Main Group", func(r *flatRule) string { return r.MainGroup }),
				output.StringColumn[*flatRule]("Group", func(r *flatRule) string { return r.GroupName }),
				output.StringColumn[*flatRule]("Code", func(r *flatRule) string { return r.GroupCode }),
				output.IntColumn[*flatRule]("Rule Id", func(r *flatRule) int { return r.RuleId }),
				output.StringColumn[*flatRule]("Description", func(r *flatRule) string { return r.Description }),
			})

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(rules))
			for i := range rules {
				items[i] = &rules[i]
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

// --- Custom WAF Rules ---

func newShieldWafCustomRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "custom-rules",
		Aliases: []string{"custom-rule"},
		Short:   "Manage custom WAF rules",
	}
	cmd.AddCommand(newShieldWafCustomRulesListCmd())
	cmd.AddCommand(withWatch(newShieldWafCustomRulesGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newShieldWafCustomRulesCreateCmd())))
	cmd.AddCommand(withFromFile(newShieldWafCustomRulesUpdateCmd()))
	cmd.AddCommand(newShieldWafCustomRulesDeleteCmd())
	return cmd
}

func newShieldWafCustomRulesListCmd() *cobra.Command {
	var (
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list <shield_zone_id>",
		Short: "List custom WAF rules for a Shield zone",
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

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.CustomWafRule], error) {
				return c.ListCustomWafRules(cmd.Context(), shieldZoneId, page, perPage)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := customWafRuleListColumns()

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

func newShieldWafCustomRulesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get custom WAF rule details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid custom WAF rule ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			rule, err := c.GetCustomWafRule(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := customWafRuleDetailColumns()

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

func newShieldWafCustomRulesCreateCmd() *cobra.Command {
	var (
		shieldZoneId int64
		ruleName     string
		ruleDesc     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom WAF rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.CustomWafRuleCreate{}
			if cmd.Flags().Changed("shield-zone-id") {
				body.ShieldZoneId = shieldZoneId
			}
			if cmd.Flags().Changed("rule-name") {
				body.RuleName = ruleName
			}
			if cmd.Flags().Changed("rule-description") {
				body.RuleDescription = ruleDesc
			}

			rule, err := c.CreateCustomWafRule(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := customWafRuleDetailColumns()

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

func newShieldWafCustomRulesUpdateCmd() *cobra.Command {
	var (
		ruleName string
		ruleDesc string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom WAF rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid custom WAF rule ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.CustomWafRuleUpdate{}
			if cmd.Flags().Changed("rule-name") {
				body.RuleName = &ruleName
			}
			if cmd.Flags().Changed("rule-description") {
				body.RuleDescription = &ruleDesc
			}

			rule, err := c.UpdateCustomWafRule(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := customWafRuleDetailColumns()

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

func newShieldWafCustomRulesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a custom WAF rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid custom WAF rule ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete custom WAF rule %d? [y/N] ", id))
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

			if err := c.DeleteCustomWafRule(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Custom WAF rule deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- WAF Profiles ---

func newShieldWafProfilesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "List WAF profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			profiles, err := c.ListWafProfiles(cmd.Context())
			if err != nil {
				return err
			}

			columns := wafProfileColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(profiles))
			for i, p := range profiles {
				items[i] = p
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

// --- WAF Engine Config ---

func newShieldWafEngineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "engine",
		Short: "Get WAF engine configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			vars, err := c.GetWafEngineConfig(cmd.Context())
			if err != nil {
				return err
			}

			columns := wafEngineConfigColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(vars))
			for i := range vars {
				items[i] = &vars[i]
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

// --- Triggered Rules ---

func newShieldWafTriggeredCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "triggered",
		Short: "Manage triggered WAF rules",
	}
	cmd.AddCommand(newShieldWafTriggeredListCmd())
	cmd.AddCommand(newShieldWafTriggeredUpdateCmd())
	return cmd
}

func newShieldWafTriggeredListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <shield_zone_id>",
		Short: "List triggered WAF rules for a Shield zone",
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

			rules, err := c.ListTriggeredWafRules(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := triggeredRuleColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(rules))
			for i, r := range rules {
				items[i] = r
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

func newShieldWafTriggeredUpdateCmd() *cobra.Command {
	var (
		ruleId string
		action int
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id>",
		Short: "Update a triggered WAF rule review action",
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

			body := &client.TriggeredRuleUpdate{
				RuleId: ruleId,
				Action: action,
			}

			if err := c.UpdateTriggeredWafRule(cmd.Context(), shieldZoneId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Triggered rule updated.")
			return err
		},
	}

	cmd.Flags().StringVar(&ruleId, "rule-id", "", "Rule ID to update (required)")
	cmd.Flags().IntVar(&action, "action", 0, "Review action")
	_ = cmd.MarkFlagRequired("rule-id")

	return cmd
}

// --- Column definitions ---

func customWafRuleListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.CustomWafRule]{
		output.IntColumn[*client.CustomWafRule]("Id", func(r *client.CustomWafRule) int { return int(r.Id) }),
		output.IntColumn[*client.CustomWafRule]("Shield Zone Id", func(r *client.CustomWafRule) int { return int(r.ShieldZoneId) }),
		output.StringColumn[*client.CustomWafRule]("Rule Name", func(r *client.CustomWafRule) string { return r.RuleName }),
		output.StringColumn[*client.CustomWafRule]("Description", func(r *client.CustomWafRule) string { return r.RuleDescription }),
	})
}

func customWafRuleDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.CustomWafRule]{
		output.IntColumn[*client.CustomWafRule]("Id", func(r *client.CustomWafRule) int { return int(r.Id) }),
		output.IntColumn[*client.CustomWafRule]("Shield Zone Id", func(r *client.CustomWafRule) int { return int(r.ShieldZoneId) }),
		output.StringColumn[*client.CustomWafRule]("Rule Name", func(r *client.CustomWafRule) string { return r.RuleName }),
		output.StringColumn[*client.CustomWafRule]("Description", func(r *client.CustomWafRule) string { return r.RuleDescription }),
		output.StringColumn[*client.CustomWafRule]("Rule JSON", func(r *client.CustomWafRule) string { return r.RuleJson }),
	})
}

func wafProfileColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.WafProfile]{
		output.IntColumn[*client.WafProfile]("Id", func(p *client.WafProfile) int { return p.Id }),
		output.StringColumn[*client.WafProfile]("Name", func(p *client.WafProfile) string { return p.Name }),
		output.BoolColumn[*client.WafProfile]("Premium", func(p *client.WafProfile) bool { return p.IsPremium }),
		output.StringColumn[*client.WafProfile]("Category", func(p *client.WafProfile) string { return p.ProfileCategory }),
		output.StringColumn[*client.WafProfile]("Description", func(p *client.WafProfile) string { return p.Description }),
	})
}

func wafEngineConfigColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.WafConfigVariable]{
		output.StringColumn[*client.WafConfigVariable]("Name", func(v *client.WafConfigVariable) string { return v.Name }),
		output.StringColumn[*client.WafConfigVariable]("Value", func(v *client.WafConfigVariable) string { return v.ValueEncoded }),
	})
}

func triggeredRuleColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.TriggeredRule]{
		output.StringColumn[*client.TriggeredRule]("Rule Id", func(r *client.TriggeredRule) string { return r.RuleId }),
		output.StringColumn[*client.TriggeredRule]("Description", func(r *client.TriggeredRule) string { return r.RuleDescription }),
		output.IntColumn[*client.TriggeredRule]("Total Triggered", func(r *client.TriggeredRule) int { return int(r.TotalTriggeredRequests) }),
	})
}
