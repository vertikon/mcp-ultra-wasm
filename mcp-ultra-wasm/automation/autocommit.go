package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Config represents the configuration for the auto-commit tool
type Config struct {
	GitHubToken string `json:"github_token"`
	GitHubOrg   string `json:"github_org"`
	RepoName    string `json:"repo_name"`
	BasePath    string `json:"base_path"`
	CommitMsg   string `json:"commit_message"`
	Branch      string `json:"branch"`
	UserName    string `json:"git_user_name"`
	UserEmail   string `json:"git_user_email"`
	AutoPush    bool   `json:"auto_push"`
	CreateDirs  bool   `json:"create_directories"`
	GitIgnore   string `json:"gitignore_template"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		GitHubOrg:  "vertikon",
		BasePath:   "E:\\vertikon\\business\\CPaaS\\WhatsApp\\moda-b2b-platform",
		CommitMsg:  "🤖 Auto-commit via MCP Ultra",
		Branch:     "main",
		UserName:   "Vertikon MCP Ultra",
		UserEmail:  "mcp@vertikon.com",
		AutoPush:   true,
		CreateDirs: true,
		GitIgnore:  "node_modules/\n*.log\n.env\n.DS_Store\ndist/\nbuild/\n.nyc_output/\ncoverage/\n*.tgz\n*.tar.gz\n",
	}
}

// ensureDirectory creates directory structure if it doesn't exist
func ensureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("🔨 Creating directory: %s", path)
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// runCommand executes a shell command and returns output
func runCommand(dir, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	log.Printf("🔧 Running: %s %s", command, strings.Join(args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Command failed: %s", string(output))
		return string(output), err
	}

	log.Printf("✅ Command successful: %s", strings.TrimSpace(string(output)))
	return string(output), nil
}

// initializeGitRepo initializes a git repository if it doesn't exist
func initializeGitRepo(config Config) error {
	repoPath := filepath.Join(config.BasePath, config.RepoName)

	// Create directory structure
	if config.CreateDirs {
		if err := ensureDirectory(repoPath); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Check if it's already a git repo
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		log.Printf("🚀 Initializing Git repository in %s", repoPath)

		// Initialize git repo
		if _, err := runCommand(repoPath, "git", "init"); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}

		// Set git config
		if _, err := runCommand(repoPath, "git", "config", "user.name", config.UserName); err != nil {
			return fmt.Errorf("failed to set git user.name: %w", err)
		}

		if _, err := runCommand(repoPath, "git", "config", "user.email", config.UserEmail); err != nil {
			return fmt.Errorf("failed to set git user.email: %w", err)
		}

		// Create .gitignore
		gitignorePath := filepath.Join(repoPath, ".gitignore")
		if err := os.WriteFile(gitignorePath, []byte(config.GitIgnore), 0644); err != nil {
			log.Printf("⚠️ Failed to create .gitignore: %v", err)
		}

		// Create initial README
		readmePath := filepath.Join(repoPath, "README.md")
		readmeContent := fmt.Sprintf("# %s\n\n✨ Repositório criado automaticamente via **MCP Ultra** by Vertikon.\n\n🤖 **MCP Ultra Features:**\n- ✅ Criação automática de repositórios GitHub\n- ✅ Automação completa de commits e push\n- ✅ Integração MCP Server <-> GitHub API\n- ✅ Gerenciamento de diretórios locais\n- ✅ Scripts de setup automático\n- ✅ Pipeline de testes end-to-end\n\n⏰ **Criado em:** %s\n🏢 **Organização:** %s\n🔧 **Template:** [MCP Ultra](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm)\n\n---\n\n🚀 **Próximos passos:**\n1. Clone o repositório: `git clone %s`\n2. Adicione seus arquivos e código\n3. Use `autocommit commit %s` para commits automáticos\n4. Explore as ferramentas MCP Ultra disponíveis\n\n💡 **Dica:** Este repositório foi criado com MCP Ultra, um template completo para automação GitHub desenvolvido pela Vertikon.\n",
			config.RepoName,
			time.Now().Format("2006-01-02 15:04:05"),
			config.GitHubOrg,
			fmt.Sprintf("https://github.com/%s/%s.git", config.GitHubOrg, config.RepoName),
			config.RepoName,
		)

		if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
			log.Printf("⚠️ Failed to create README.md: %v", err)
		}

		// Set remote origin
		remoteURL := fmt.Sprintf("https://%s@github.com/%s/%s.git",
			config.GitHubToken, config.GitHubOrg, config.RepoName)

		if _, err := runCommand(repoPath, "git", "remote", "add", "origin", remoteURL); err != nil {
			log.Printf("⚠️ Failed to add remote origin: %v", err)
		}
	}

	return nil
}

// commitAndPush commits changes and pushes to GitHub
func commitAndPush(config Config) error {
	repoPath := filepath.Join(config.BasePath, config.RepoName)

	log.Printf("📁 Working in directory: %s", repoPath)

	// Check if directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoPath)
	}

	// Pull latest changes first
	log.Printf("📥 Pulling latest changes...")
	if _, err := runCommand(repoPath, "git", "pull", "origin", config.Branch); err != nil {
		log.Printf("⚠️ Pull failed, might be initial commit: %v", err)
	}

	// Check git status
	status, err := runCommand(repoPath, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(status) == "" {
		log.Printf("ℹ️ No changes to commit")
		return nil
	}

	log.Printf("📝 Found changes:\n%s", status)

	// Add all changes
	if _, err := runCommand(repoPath, "git", "add", "."); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Commit changes
	commitMessage := fmt.Sprintf("%s\n\n🚀 Automated commit via MCP Ultra\n⏰ %s\n🔧 Template: github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm",
		config.CommitMsg, time.Now().Format("2006-01-02 15:04:05"))

	if _, err := runCommand(repoPath, "git", "commit", "-m", commitMessage); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push to remote if enabled
	if config.AutoPush {
		log.Printf("📤 Pushing to GitHub...")
		if _, err := runCommand(repoPath, "git", "push", "origin", config.Branch); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}

		log.Printf("🎉 Successfully pushed to GitHub!")
		log.Printf("🔗 Repository URL: https://github.com/%s/%s", config.GitHubOrg, config.RepoName)
	}

	return nil
}

// loadConfigFromFile loads configuration from JSON file
func loadConfigFromFile(filename string) (Config, error) {
	config := DefaultConfig()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("📋 Config file not found, using defaults")
		return config, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	log.Printf("✅ Configuration loaded from %s", filename)
	return config, nil
}

// saveConfigToFile saves configuration to JSON file
func saveConfigToFile(config Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("💾 Configuration saved to %s", filename)
	return nil
}

// interactiveConfig allows user to input configuration interactively
func interactiveConfig() Config {
	reader := bufio.NewReader(os.Stdin)
	config := DefaultConfig()

	fmt.Printf("🔧 Configuração Interativa do MCP Ultra Auto-Commit\n")
	fmt.Printf("================================================\n\n")

	fmt.Printf("📋 GitHub Token (necessário): ")
	if token, _ := reader.ReadString('\n'); strings.TrimSpace(token) != "" {
		config.GitHubToken = strings.TrimSpace(token)
	}

	fmt.Printf("🏢 Organização GitHub [%s]: ", config.GitHubOrg)
	if org, _ := reader.ReadString('\n'); strings.TrimSpace(org) != "" {
		config.GitHubOrg = strings.TrimSpace(org)
	}

	fmt.Printf("📁 Nome do repositório: ")
	if repo, _ := reader.ReadString('\n'); strings.TrimSpace(repo) != "" {
		config.RepoName = strings.TrimSpace(repo)
	}

	fmt.Printf("📂 Caminho base [%s]: ", config.BasePath)
	if path, _ := reader.ReadString('\n'); strings.TrimSpace(path) != "" {
		config.BasePath = strings.TrimSpace(path)
	}

	fmt.Printf("💬 Mensagem de commit [%s]: ", config.CommitMsg)
	if msg, _ := reader.ReadString('\n'); strings.TrimSpace(msg) != "" {
		config.CommitMsg = strings.TrimSpace(msg)
	}

	return config
}

func main() {
	log.Printf("🚀 MCP Ultra Auto-Commit Tool v1.0")
	log.Printf("==================================\n")

	configFile := "autocommit-config.json"

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [init|commit|config] [repo-name]\n\n", os.Args[0])
		fmt.Printf("Commands:\n")
		fmt.Printf("  init <repo-name>    - Initialize new repository\n")
		fmt.Printf("  commit <repo-name>  - Commit and push changes\n")
		fmt.Printf("  config              - Interactive configuration\n")
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s config\n", os.Args[0])
		fmt.Printf("  %s init meu-novo-repo\n", os.Args[0])
		fmt.Printf("  %s commit meu-novo-repo\n", os.Args[0])
		fmt.Printf("\n🔗 MCP Ultra: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "config":
		config := interactiveConfig()
		if err := saveConfigToFile(config, configFile); err != nil {
			log.Fatalf("❌ Failed to save config: %v", err)
		}
		fmt.Printf("\n✅ Configuration saved! Now run 'init' or 'commit' commands.\n")

	case "init":
		if len(os.Args) < 3 {
			log.Fatalf("❌ Repository name required for init command")
		}

		config, err := loadConfigFromFile(configFile)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}

		config.RepoName = os.Args[2]

		if config.GitHubToken == "" {
			log.Fatalf("❌ GitHub token required! Run 'config' command first.")
		}

		if err := initializeGitRepo(config); err != nil {
			log.Fatalf("❌ Failed to initialize repository: %v", err)
		}

		log.Printf("✅ Repository '%s' initialized successfully!", config.RepoName)
		log.Printf("📁 Location: %s", filepath.Join(config.BasePath, config.RepoName))
		log.Printf("🔗 GitHub: https://github.com/%s/%s", config.GitHubOrg, config.RepoName)

	case "commit":
		if len(os.Args) < 3 {
			log.Fatalf("❌ Repository name required for commit command")
		}

		config, err := loadConfigFromFile(configFile)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}

		config.RepoName = os.Args[2]

		if config.GitHubToken == "" {
			log.Fatalf("❌ GitHub token required! Run 'config' command first.")
		}

		if err := commitAndPush(config); err != nil {
			log.Fatalf("❌ Failed to commit and push: %v", err)
		}

	default:
		log.Fatalf("❌ Unknown command: %s", command)
	}
}
