package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newScriptsSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage edge script secrets",
	}
	cmd.AddCommand(newScriptsSecretsListCmd())
	cmd.AddCommand(withFromFile(withInteractive(newScriptsSecretsAddCmd())))
	cmd.AddCommand(withFromFile(newScriptsSecretsUpdateCmd()))
	cmd.AddCommand(newScriptsSecretsDeleteCmd())
	return cmd
}

func newScriptsSecretsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <script_id>",
		Short: "List secrets for an edge script",
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

			secrets, err := c.ListEdgeScriptSecrets(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := edgeScriptSecretListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(secrets))
			for i, s := range secrets {
				items[i] = s
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

func newScriptsSecretsAddCmd() *cobra.Command {
	var (
		name   string
		secret string
	)

	cmd := &cobra.Command{
		Use:   "add <script_id>",
		Short: "Add a secret to an edge script",
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

			body := &client.EdgeScriptSecretCreate{}

			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("secret") {
				body.Secret = secret
			}

			s, err := c.AddEdgeScriptSecret(cmd.Context(), scriptId, body)
			if err != nil {
				return err
			}

			columns := edgeScriptSecretDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, s)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Secret name (required)")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret value (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("secret")

	return cmd
}

func newScriptsSecretsUpdateCmd() *cobra.Command {
	var secret string

	cmd := &cobra.Command{
		Use:   "update <script_id> <secret_id>",
		Short: "Update a secret on an edge script",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}
			secretId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid secret ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.EdgeScriptSecretUpdate{}

			if cmd.Flags().Changed("secret") {
				body.Secret = secret
			}

			if err := c.UpdateEdgeScriptSecret(cmd.Context(), scriptId, secretId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Secret updated.")
			return err
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "Secret value (required)")
	_ = cmd.MarkFlagRequired("secret")

	return cmd
}

func newScriptsSecretsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <script_id> <secret_id>",
		Short: "Delete a secret from an edge script",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}
			secretId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid secret ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete secret %d from edge script %d? [y/N] ", secretId, scriptId))
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

			if err := c.DeleteEdgeScriptSecret(cmd.Context(), scriptId, secretId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Secret deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Secret column definitions ---

func edgeScriptSecretListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptSecret]{
		output.IntColumn[*client.EdgeScriptSecret]("Id", func(s *client.EdgeScriptSecret) int { return int(s.Id) }),
		output.StringColumn[*client.EdgeScriptSecret]("Name", func(s *client.EdgeScriptSecret) string { return s.Name }),
		output.StringColumn[*client.EdgeScriptSecret]("Last Modified", func(s *client.EdgeScriptSecret) string { return s.LastModified }),
	})
}

func edgeScriptSecretDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptSecret]{
		output.IntColumn[*client.EdgeScriptSecret]("Id", func(s *client.EdgeScriptSecret) int { return int(s.Id) }),
		output.StringColumn[*client.EdgeScriptSecret]("Name", func(s *client.EdgeScriptSecret) string { return s.Name }),
		output.StringColumn[*client.EdgeScriptSecret]("Last Modified", func(s *client.EdgeScriptSecret) string { return s.LastModified }),
	})
}
