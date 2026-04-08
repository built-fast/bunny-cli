package cmd

import (
	"fmt"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCountriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "countries",
		Short: "List countries",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewCountryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			countries, err := c.ListCountries(cmd.Context())
			if err != nil {
				return err
			}

			columns := countryColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(countries))
			for i, co := range countries {
				items[i] = co
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

func countryColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Country]{
		output.StringColumn[*client.Country]("Name", func(c *client.Country) string { return c.Name }),
		output.StringColumn[*client.Country]("ISO Code", func(c *client.Country) string { return c.IsoCode }),
		output.BoolColumn[*client.Country]("EU", func(c *client.Country) bool { return c.IsEU }),
		output.FloatColumn[*client.Country]("Tax Rate", func(c *client.Country) float64 { return c.TaxRate }),
		output.StringColumn[*client.Country]("POPs", func(c *client.Country) string { return client.FormatPopList(c.PopList) }),
	})
}
