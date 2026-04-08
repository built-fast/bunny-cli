package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newPullZonePurgeCmd() *cobra.Command {
	var (
		tag string
		yes bool
	)

	cmd := &cobra.Command{
		Use:   "purge <pull_zone_id>",
		Short: "Purge the cache for a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			msg := fmt.Sprintf("Are you sure you want to purge the cache for pull zone %d? [y/N] ", id)
			if tag != "" {
				msg = fmt.Sprintf("Are you sure you want to purge cache tag %q for pull zone %d? [y/N] ", tag, id)
			}

			if !yes {
				confirmed, err := confirm(cmd, msg)
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Purge canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.PurgePullZoneCache(cmd.Context(), id, tag); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Cache purged.")
			return err
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Purge only items with this cache tag")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}
