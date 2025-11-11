package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista os templates disponíveis",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := MustApp(cmd)
			ctx := cmd.Context()

			templates, err := app.TemplateService().List(ctx)
			if err != nil {
				return err
			}

			if asJSON {
				data, err := json.MarshalIndent(templates, "", "  ")
				if err != nil {
					return fmt.Errorf("serializar templates: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s %-20s %s\n", "NAME", "VERSION", "DESCRIPTION")
			for _, tmpl := range templates {
				fmt.Fprintf(out, "%-20s %-20s %s\n", tmpl.Name, tmpl.Version, tmpl.Description)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Formato JSON")
	return cmd
}

