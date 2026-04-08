package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newScriptsVariablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variables",
		Aliases: []string{"vars"},
		Short:   "Manage edge script variables",
	}
	cmd.AddCommand(newScriptsVariablesListCmd())
	cmd.AddCommand(newScriptsVariablesGetCmd())
	cmd.AddCommand(withFromFile(withInteractive(newScriptsVariablesAddCmd())))
	cmd.AddCommand(withFromFile(newScriptsVariablesUpdateCmd()))
	cmd.AddCommand(newScriptsVariablesDeleteCmd())
	return cmd
}

func newScriptsVariablesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <script_id>",
		Short: "List variables for an edge script",
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

			script, err := c.GetEdgeScript(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := edgeScriptVariableListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(script.EdgeScriptVariables))
			for i := range script.EdgeScriptVariables {
				items[i] = &script.EdgeScriptVariables[i]
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

func newScriptsVariablesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <script_id> <variable_id>",
		Short: "Get a variable for an edge script",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}
			variableId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid variable ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			v, err := c.GetEdgeScriptVariable(cmd.Context(), scriptId, variableId)
			if err != nil {
				return err
			}

			columns := edgeScriptVariableDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newScriptsVariablesAddCmd() *cobra.Command {
	var (
		name         string
		defaultValue string
		required     bool
	)

	cmd := &cobra.Command{
		Use:   "add <script_id>",
		Short: "Add a variable to an edge script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.EdgeScriptVariableCreate{}

			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("default-value") {
				body.DefaultValue = defaultValue
			}
			if cmd.Flags().Changed("required") {
				body.Required = required
			}

			v, err := c.AddEdgeScriptVariable(cmd.Context(), scriptId, body)
			if err != nil {
				return err
			}

			columns := edgeScriptVariableDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Variable name (required)")
	cmd.Flags().StringVar(&defaultValue, "default-value", "", "Default value")
	cmd.Flags().BoolVar(&required, "required", false, "Whether the variable is required")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newScriptsVariablesUpdateCmd() *cobra.Command {
	var (
		defaultValue string
		required     bool
	)

	cmd := &cobra.Command{
		Use:   "update <script_id> <variable_id>",
		Short: "Update a variable on an edge script",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}
			variableId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid variable ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.EdgeScriptVariableUpdate{}

			if cmd.Flags().Changed("default-value") {
				body.DefaultValue = &defaultValue
			}
			if cmd.Flags().Changed("required") {
				body.Required = &required
			}

			if err := c.UpdateEdgeScriptVariable(cmd.Context(), scriptId, variableId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Variable updated.")
			return err
		},
	}

	cmd.Flags().StringVar(&defaultValue, "default-value", "", "Default value")
	cmd.Flags().BoolVar(&required, "required", false, "Whether the variable is required")

	return cmd
}

func newScriptsVariablesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <script_id> <variable_id>",
		Short: "Delete a variable from an edge script",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}
			variableId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid variable ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete variable %d from edge script %d? [y/N] ", variableId, scriptId))
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

			if err := c.DeleteEdgeScriptVariable(cmd.Context(), scriptId, variableId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Variable deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Variable column definitions ---

func edgeScriptVariableListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptVariable]{
		output.IntColumn[*client.EdgeScriptVariable]("Id", func(v *client.EdgeScriptVariable) int { return int(v.Id) }),
		output.StringColumn[*client.EdgeScriptVariable]("Name", func(v *client.EdgeScriptVariable) string { return v.Name }),
		output.BoolColumn[*client.EdgeScriptVariable]("Required", func(v *client.EdgeScriptVariable) bool { return v.Required }),
		output.StringColumn[*client.EdgeScriptVariable]("Default Value", func(v *client.EdgeScriptVariable) string { return v.DefaultValue }),
	})
}

func edgeScriptVariableDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptVariable]{
		output.IntColumn[*client.EdgeScriptVariable]("Id", func(v *client.EdgeScriptVariable) int { return int(v.Id) }),
		output.StringColumn[*client.EdgeScriptVariable]("Name", func(v *client.EdgeScriptVariable) string { return v.Name }),
		output.BoolColumn[*client.EdgeScriptVariable]("Required", func(v *client.EdgeScriptVariable) bool { return v.Required }),
		output.StringColumn[*client.EdgeScriptVariable]("Default Value", func(v *client.EdgeScriptVariable) string { return v.DefaultValue }),
	})
}
