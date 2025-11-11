package depguardlite

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type Rules struct {
	Deny               map[string]string `json:"deny"`
	ExcludePaths       []string          `json:"excludePaths"`
	InternalLayerRules []struct {
		Name    string   `json:"name"`
		From    string   `json:"from"`
		AllowTo []string `json:"allowTo"`
		DenyTo  []string `json:"denyTo"`
		Message string   `json:"message"`
	} `json:"internalLayerRules"`
}

func loadRules() (*Rules, error) {
	// Try to find the config file starting from the current directory
	cfgPath := "internal/config/dep_rules.json"

	// Try current directory first
	if _, err := os.Stat(cfgPath); err == nil {
		b, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, err
		}
		var r Rules
		return &r, json.Unmarshal(b, &r)
	}

	// If not found, try to find go.mod and use that directory as root
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		gomodPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(gomodPath); err == nil {
			cfgPath = filepath.Join(dir, "internal/config/dep_rules.json")
			b, err := os.ReadFile(cfgPath)
			if err != nil {
				// If config file doesn't exist, return empty rules (no errors)
				return &Rules{}, nil
			}
			var r Rules
			return &r, json.Unmarshal(b, &r)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod, return empty rules
			return &Rules{}, nil
		}
		dir = parent
	}
}

func isExcluded(filePath string, excludes []string) bool {
	fp := filepath.ToSlash(filePath)
	for _, e := range excludes {
		if strings.Contains(fp, filepath.ToSlash(e)) {
			return true
		}
	}
	return false
}

func matchAnyPrefix(s string, prefixes []string) bool {
	s = filepath.ToSlash(s)
	for _, p := range prefixes {
		if strings.HasPrefix(s, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}

func checkLayerRule(filePath, importPath string, r *Rules) (string, bool) {
	fp := filepath.ToSlash(filePath)
	for _, lr := range r.InternalLayerRules {
		// Se o arquivo pertence ao "from" da regra
		if strings.HasPrefix(fp, filepath.ToSlash(lr.From)) {
			// Negações primeiro
			for _, d := range lr.DenyTo {
				if strings.HasPrefix(importPath, filepath.ToSlash(d)) {
					if lr.Message != "" {
						return lr.Message, true
					}
					return "import não permitido por regra de camada: " + lr.Name, true
				}
			}
			// Se existir allow-list, restringe
			if len(lr.AllowTo) > 0 && !matchAnyPrefix(importPath, lr.AllowTo) {
				return "import não permitido (apenas " + strings.Join(lr.AllowTo, ", ") + ")", true
			}
		}
	}
	return "", false
}

var Analyzer = &analysis.Analyzer{
	Name: "depguardlite",
	Doc:  "Valida imports proibidos e regras de camadas (facades + internos)",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rules, err := loadRules()
		if err != nil {
			return nil, err
		}

		for _, f := range pass.Files {
			file := pass.Fset.Position(f.Pos()).Filename
			if file == "" || isExcluded(file, rules.ExcludePaths) {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				imp, ok := n.(*ast.ImportSpec)
				if !ok || imp.Path == nil {
					return true
				}
				ip := strings.Trim(imp.Path.Value, `"`)

				// Denylist explícita por módulo/pacote
				for blocked, msg := range rules.Deny {
					if ip == blocked || strings.HasPrefix(ip, blocked) {
						pass.Reportf(imp.Pos(), "import proibido: %s (%s)", ip, msg)
						return true
					}
				}

				// Regras de camadas internas por caminho do arquivo e do import
				if msg, violated := checkLayerRule(file, ip, rules); violated {
					pass.Reportf(imp.Pos(), "violação de camada: %s → %s (%s)", file, ip, msg)
				}
				return true
			})
		}
		return nil, nil
	},
}
