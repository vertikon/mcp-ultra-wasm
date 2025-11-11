package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/vertikon/mcp-ultra-templates/internal/models"
	templateservice "github.com/vertikon/mcp-ultra-templates/internal/services/template"
)

func renderCommand() *cobra.Command {
	var (
		templateName string
		outputDir    string
		valuesFile   string
		setValues    []string
		overwrite    bool
		interactive  bool
	)

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Renderiza um template para o diretório alvo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if templateName == "" {
				return fmt.Errorf("--template é obrigatório")
			}
			if outputDir == "" {
				return fmt.Errorf("--output é obrigatório")
			}

			app := MustApp(cmd)
			ctx := cmd.Context()

			values, err := buildValues(valuesFile, setValues)
			if err != nil {
				return err
			}

			if interactive {
				if err := promptMissingValues(cmd, app, ctx, templateName, values); err != nil {
					return err
				}
			}

			resp, err := app.TemplateService().Render(ctx, templateservice.RenderRequest{
				TemplateName: templateName,
				OutputDir:    outputDir,
				Values:       values,
				Overwrite:    overwrite,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Template %s renderizado em %s\n", resp.Template.DisplayName, resp.Output)
			return nil
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "", "Nome do template")
	cmd.Flags().StringVar(&outputDir, "output", "", "Diretório de destino")
	cmd.Flags().StringVar(&valuesFile, "values", "", "Arquivo YAML com variáveis")
	cmd.Flags().StringArrayVar(&setValues, "set", nil, "Definições no formato chave=valor")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Permitir sobrescrever diretório de destino")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Solicitar interativamente variáveis ausentes")

	return cmd
}

func buildValues(file string, sets []string) (map[string]string, error) {
	values := make(map[string]string)

	if file != "" {
		data, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, fmt.Errorf("ler values file: %w", err)
		}
		if err := yaml.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("parse values file: %w", err)
		}
	}

	for _, set := range sets {
		parts := strings.SplitN(set, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("flag --set inválida, use chave=valor: %s", set)
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return values, nil
}

func promptMissingValues(cmd *cobra.Command, app *App, ctx context.Context, templateName string, values map[string]string) error {
	templates, err := app.TemplateService().List(ctx)
	if err != nil {
		return fmt.Errorf("listar templates: %w", err)
	}

	var meta *models.TemplateMetadata
	for i := range templates {
		if templates[i].Name == templateName {
			meta = &templates[i]
			break
		}
	}

	if meta == nil {
		return fmt.Errorf("template %s não encontrado", templateName)
	}

	applyDefaults(meta, values)

	reader := bufio.NewReader(cmd.InOrStdin())
	for _, variable := range meta.Variables {
		if !variable.Required {
			continue
		}
		if strings.TrimSpace(values[variable.Key]) != "" {
			continue
		}

		for {
			prompt := fmt.Sprintf("Informe o valor para %s", variable.Key)
			if variable.Description != "" {
				prompt += fmt.Sprintf(" (%s)", variable.Description)
			}
			prompt += ": "

			fmt.Fprint(cmd.OutOrStdout(), prompt)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ler valor para %s: %w", variable.Key, err)
			}

			value := strings.TrimSpace(input)
			if value == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Valor obrigatório, tente novamente.")
				continue
			}
			values[variable.Key] = value
			break
		}
	}

	return nil
}

func applyDefaults(meta *models.TemplateMetadata, values map[string]string) {
	for k, v := range meta.Defaults {
		if _, ok := values[k]; !ok || strings.TrimSpace(values[k]) == "" {
			values[k] = v
		}
	}
}
