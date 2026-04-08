package cmd

import "github.com/spf13/cobra"

func newStreamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Manage Stream video libraries and videos",
	}
	cmd.AddCommand(newStreamLibrariesCmd())
	cmd.AddCommand(newStreamVideosCmd())
	cmd.AddCommand(newStreamCollectionsCmd())
	cmd.AddCommand(newStreamCaptionsCmd())
	cmd.AddCommand(newStreamStatisticsCmd())
	cmd.AddCommand(newStreamHeatmapCmd())
	return cmd
}
