package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShieldBotDetectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bot-detection",
		Aliases: []string{"bot"},
		Short:   "Manage bot detection configuration",
	}
	cmd.AddCommand(newShieldBotDetectionGetCmd())
	cmd.AddCommand(withFromFile(newShieldBotDetectionUpdateCmd()))
	return cmd
}

func newShieldBotDetectionGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <shield_zone_id>",
		Short: "Get bot detection configuration",
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

			config, err := c.GetBotDetection(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := botDetectionColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, config)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldBotDetectionUpdateCmd() *cobra.Command {
	var (
		executionMode    int
		requestIntegrity int
		ipSensitivity    int
		fpSensitivity    int
		fpAggression     int
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id>",
		Short: "Update bot detection configuration",
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

			body := &client.BotDetectionUpdate{}
			if cmd.Flags().Changed("execution-mode") {
				body.ExecutionMode = &executionMode
			}
			if cmd.Flags().Changed("request-integrity") {
				body.RequestIntegrity = &client.BotDetectionSensitivityConfig{Sensitivity: requestIntegrity}
			}
			if cmd.Flags().Changed("ip-sensitivity") {
				body.IpAddress = &client.BotDetectionSensitivityConfig{Sensitivity: ipSensitivity}
			}
			if cmd.Flags().Changed("fingerprint-sensitivity") || cmd.Flags().Changed("fingerprint-aggression") {
				body.BrowserFingerprint = &client.BrowserFingerprintConfig{
					Sensitivity: fpSensitivity,
					Aggression:  fpAggression,
				}
			}

			if err := c.UpdateBotDetection(cmd.Context(), shieldZoneId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Bot detection configuration updated.")
			return err
		},
	}

	cmd.Flags().IntVar(&executionMode, "execution-mode", 0, "Execution mode (0=Learn, 1=Protect)")
	cmd.Flags().IntVar(&requestIntegrity, "request-integrity", 0, "Request integrity sensitivity (0=Off, 1=Low, 2=Medium, 3=High)")
	cmd.Flags().IntVar(&ipSensitivity, "ip-sensitivity", 0, "IP address sensitivity (0=Off, 1=Low, 2=Medium, 3=High)")
	cmd.Flags().IntVar(&fpSensitivity, "fingerprint-sensitivity", 0, "Browser fingerprint sensitivity (0=Off, 1=Low, 2=Medium, 3=High)")
	cmd.Flags().IntVar(&fpAggression, "fingerprint-aggression", 0, "Browser fingerprint aggression level")

	return cmd
}

// --- Column definitions ---

func botDetectionColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.BotDetectionConfig]{
		output.IntColumn[*client.BotDetectionConfig]("Shield Zone Id", func(c *client.BotDetectionConfig) int { return int(c.ShieldZoneId) }),
		output.StringColumn[*client.BotDetectionConfig]("Execution Mode", func(c *client.BotDetectionConfig) string {
			return client.ShieldExecutionModeName(c.ExecutionMode)
		}),
		output.IntColumn[*client.BotDetectionConfig]("Request Integrity", func(c *client.BotDetectionConfig) int {
			return c.RequestIntegrity.Sensitivity
		}),
		output.IntColumn[*client.BotDetectionConfig]("IP Sensitivity", func(c *client.BotDetectionConfig) int {
			return c.IpAddress.Sensitivity
		}),
		output.IntColumn[*client.BotDetectionConfig]("FP Sensitivity", func(c *client.BotDetectionConfig) int {
			return c.BrowserFingerprint.Sensitivity
		}),
		output.IntColumn[*client.BotDetectionConfig]("FP Aggression", func(c *client.BotDetectionConfig) int {
			return c.BrowserFingerprint.Aggression
		}),
	})
}
