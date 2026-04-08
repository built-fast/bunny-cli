package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func newScriptsCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "Manage edge script code",
	}
	cmd.AddCommand(newScriptsCodeGetCmd())
	cmd.AddCommand(newScriptsCodeSetCmd())
	return cmd
}

func newScriptsCodeGetCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get edge script code",
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

			code, err := c.GetEdgeScriptCode(cmd.Context(), id)
			if err != nil {
				return err
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, []byte(code.Code), 0644); err != nil {
					return fmt.Errorf("writing code file: %w", err)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Code written to %s.\n", outputFile)
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), code.Code)
			return err
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Write code to file instead of stdout")

	return cmd
}

func newScriptsCodeSetCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Set edge script code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			var reader io.Reader
			if file == "-" {
				reader = cmd.InOrStdin()
			} else {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening code file: %w", err)
				}
				defer func() { _ = f.Close() }()
				reader = f
			}

			data, err := io.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("reading code: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.SetEdgeScriptCode(cmd.Context(), id, string(data)); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Code updated.")
			return err
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to code file (use '-' for stdin) (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
