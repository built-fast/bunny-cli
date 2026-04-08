package cmd

import (
	"fmt"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regions",
		Short: "List CDN regions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewRegionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			regions, err := c.ListRegions(cmd.Context())
			if err != nil {
				return err
			}

			columns := regionColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(regions))
			for i, r := range regions {
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

func regionColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Region]{
		output.IntColumn[*client.Region]("Id", func(r *client.Region) int { return int(r.Id) }),
		output.StringColumn[*client.Region]("Name", func(r *client.Region) string { return r.Name }),
		output.StringColumn[*client.Region]("Region Code", func(r *client.Region) string { return r.RegionCode }),
		output.StringColumn[*client.Region]("Continent", func(r *client.Region) string { return r.ContinentCode }),
		output.StringColumn[*client.Region]("Country", func(r *client.Region) string { return r.CountryCode }),
		output.FloatColumn[*client.Region]("Price/GB", func(r *client.Region) float64 { return r.PricePerGigabyte }),
		output.BoolColumn[*client.Region]("Latency Routing", func(r *client.Region) bool { return r.AllowLatencyRouting }),
	})
}
