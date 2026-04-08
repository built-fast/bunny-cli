package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShieldUploadScanningCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upload-scanning",
		Aliases: []string{"scanning"},
		Short:   "Manage upload scanning configuration",
	}
	cmd.AddCommand(newShieldUploadScanningGetCmd())
	cmd.AddCommand(withFromFile(newShieldUploadScanningUpdateCmd()))
	return cmd
}

func newShieldUploadScanningGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <shield_zone_id>",
		Short: "Get upload scanning configuration",
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

			config, err := c.GetUploadScanning(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := uploadScanningColumns()

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

func newShieldUploadScanningUpdateCmd() *cobra.Command {
	var (
		isEnabled             bool
		csamScanningMode      int
		antivirusScanningMode int
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id>",
		Short: "Update upload scanning configuration",
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

			body := &client.UploadScanningUpdate{}
			if cmd.Flags().Changed("enabled") {
				body.IsEnabled = &isEnabled
			}
			if cmd.Flags().Changed("csam-scanning-mode") {
				body.CsamScanningMode = &csamScanningMode
			}
			if cmd.Flags().Changed("antivirus-scanning-mode") {
				body.AntivirusScanningMode = &antivirusScanningMode
			}

			if err := c.UpdateUploadScanning(cmd.Context(), shieldZoneId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Upload scanning configuration updated.")
			return err
		},
	}

	cmd.Flags().BoolVar(&isEnabled, "enabled", false, "Enable or disable upload scanning")
	cmd.Flags().IntVar(&csamScanningMode, "csam-scanning-mode", 0, "CSAM scanning mode")
	cmd.Flags().IntVar(&antivirusScanningMode, "antivirus-scanning-mode", 0, "Antivirus scanning mode")

	return cmd
}

// --- Column definitions ---

func uploadScanningColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.UploadScanningConfig]{
		output.IntColumn[*client.UploadScanningConfig]("Shield Zone Id", func(c *client.UploadScanningConfig) int { return int(c.ShieldZoneId) }),
		output.BoolColumn[*client.UploadScanningConfig]("Enabled", func(c *client.UploadScanningConfig) bool { return c.IsEnabled }),
		output.IntColumn[*client.UploadScanningConfig]("Antivirus Mode", func(c *client.UploadScanningConfig) int { return c.AntivirusScanningMode }),
		output.IntColumn[*client.UploadScanningConfig]("CSAM Mode", func(c *client.UploadScanningConfig) int { return c.CsamScanningMode }),
	})
}
