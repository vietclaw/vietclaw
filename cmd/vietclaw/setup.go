package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vietclaw/internal/config"
	"vietclaw/internal/db"
	"vietclaw/internal/providers"
)

func runSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\033[1;36m==================================================\033[0m")
	fmt.Println("\033[1;36m       🚀 CHƯƠNG TRÌNH THIẾT LẬP VIETCLAW TUI     \033[0m")
	fmt.Println("\033[1;36m==================================================\033[0m")
	fmt.Println("Hệ thống sẽ hướng dẫn bạn cấu hình các thông số cơ bản.")
	fmt.Println("Sử dụng các phím mũi tên và phím cách để chọn nhanh.")
	fmt.Println()

	paths, err := config.DefaultPaths()
	if err != nil {
		return err
	}

	// Load existing config or start with defaults
	var cfg config.Config
	if _, err := os.Stat(paths.ConfigFile); err == nil {
		loaded, loadErr := config.Load(paths.ConfigFile)
		if loadErr == nil {
			cfg = loaded
			fmt.Println("\033[32m[ok] Đã tìm thấy file cấu hình hiện có, tiến hành chỉnh sửa.\033[0m")
		} else {
			cfg = config.Default(paths)
		}
	} else {
		cfg = config.Default(paths)
	}

	envVars := make(map[string]string)

	// --- 1. SERVER CONFIG ---
	fmt.Println("\n\033[1;33m--- 1. Cấu hình Server Daemon ---\033[0m")
	cfg.Server.Host = promptString(reader, "Địa chỉ Host", cfg.Server.Host)
	cfg.Server.Port = promptInt(reader, "Cổng Port chạy daemon", cfg.Server.Port)

	// --- 2. PROVIDERS CONFIG ---
	fmt.Println("\n\033[1;33m--- 2. Chọn LLM Providers muốn sử dụng ---\033[0m")
	fmt.Println("\033[2m  Tip: OpenCode Zen là cách nhanh nhất để dùng AI miễn phí demo (https://opencode.ai/auth)\033[0m")

	providerOpts := []string{"OpenCode Zen (recommended, free demo)", "OpenAI", "Google Gemini", "Anthropic Claude", "Ollama (local, hoàn toàn miễn phí)"}
	providerDefaults := []bool{true, false, false, false, false}

	selectedProviders, err := promptMultiSelect("Chọn LLM Providers bạn muốn bật", providerOpts, providerDefaults)
	if err != nil {
		return err
	}

	enableZen := selectedProviders[0]
	enableOpenAI := selectedProviders[1]
	enableGemini := selectedProviders[2]
	enableAnthropic := selectedProviders[3]
	enableOllama := selectedProviders[4]

	var enabledIDs []string
	var enabledModels []string

	// OpenCode Zen configuration
	if enableZen {
		fmt.Println("\n\033[36m>> Cấu hình OpenCode Zen:\033[0m")
		fmt.Println("   Đăng ký tại: https://opencode.ai/auth")
		apiKey := promptSecret(reader, "Nhập Zen API Key (bắt đầu bằng zen_..., Enter nếu dùng biến môi trường)")
		if apiKey != "" {
			envVars["OPENCODE_ZEN_KEY"] = apiKey
		}

		zenCfg := findOrCreateProvider(&cfg, "zen", "opencode-zen")
		zenCfg.Enabled = true
		zenCfg.APIKeyEnv = "OPENCODE_ZEN_KEY"
		zenCfg.BaseURL = providers.ZenBaseURL

		// Fetch live model list from Zen
		pick := fetchAndPickModel(zenCfg.BaseURL, "OPENCODE_ZEN_KEY", apiKey, "deepseek/deepseek-chat")
		zenCfg.DefaultModel = pick
		updateProvider(&cfg, *zenCfg)

		enabledIDs = append(enabledIDs, "zen")
		enabledModels = append(enabledModels, zenCfg.DefaultModel)
	} else {
		disableProvider(&cfg, "zen")
	}

	// OpenAI configuration
	if enableOpenAI {
		fmt.Println("\n\033[36m>> Cấu hình OpenAI:\033[0m")
		apiKey := promptSecret(reader, "Nhập OpenAI API Key (Enter nếu dùng biến môi trường)")
		if apiKey != "" {
			envVars["OPENAI_API_KEY"] = apiKey
		}

		openaiCfg := findOrCreateProvider(&cfg, "openai", "openai")
		openaiCfg.Enabled = true
		openaiCfg.APIKeyEnv = "OPENAI_API_KEY"
		openaiCfg.BaseURL = "https://api.openai.com/v1"

		pick := fetchAndPickModel(openaiCfg.BaseURL, "OPENAI_API_KEY", apiKey, "gpt-4o-mini")
		openaiCfg.DefaultModel = pick
		updateProvider(&cfg, *openaiCfg)

		enabledIDs = append(enabledIDs, "openai")
		enabledModels = append(enabledModels, openaiCfg.DefaultModel)
	} else {
		disableProvider(&cfg, "openai")
	}

	// Gemini configuration
	if enableGemini {
		fmt.Println("\n\033[36m>> Cấu hình Google Gemini:\033[0m")
		apiKey := promptSecret(reader, "Nhập Gemini API Key (Enter nếu dùng biến môi trường)")
		if apiKey != "" {
			envVars["GEMINI_API_KEY"] = apiKey
		}

		geminiCfg := findOrCreateProvider(&cfg, "gemini", "gemini")
		geminiCfg.Enabled = true
		geminiCfg.APIKeyEnv = "GEMINI_API_KEY"

		// Gemini doesn't expose OpenAI-compatible /models — offer curated list
		geminiModels := []string{
			"gemini-2.5-flash-preview-05-20",
			"gemini-2.5-pro-preview-06-05",
			"gemini-2.0-flash",
			"gemini-1.5-flash",
			"gemini-1.5-pro",
		}
		pick := pickModelFromList(geminiModels, "gemini-2.5-flash-preview-05-20", "Chọn model Gemini mặc định")
		geminiCfg.DefaultModel = pick
		updateProvider(&cfg, *geminiCfg)

		enabledIDs = append(enabledIDs, "gemini")
		enabledModels = append(enabledModels, geminiCfg.DefaultModel)
	} else {
		disableProvider(&cfg, "gemini")
	}

	// Anthropic configuration
	if enableAnthropic {
		fmt.Println("\n\033[36m>> Cấu hình Anthropic Claude:\033[0m")
		apiKey := promptSecret(reader, "Nhập Anthropic API Key (Enter nếu dùng biến môi trường)")
		if apiKey != "" {
			envVars["ANTHROPIC_API_KEY"] = apiKey
		}

		anthropicCfg := findOrCreateProvider(&cfg, "anthropic", "anthropic")
		anthropicCfg.Enabled = true
		anthropicCfg.APIKeyEnv = "ANTHROPIC_API_KEY"

		anthropicModels := []string{
			"claude-sonnet-4-5",
			"claude-opus-4-5",
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
			"claude-3-opus-20240229",
		}
		pick := pickModelFromList(anthropicModels, "claude-3-5-sonnet-20241022", "Chọn model Claude mặc định")
		anthropicCfg.DefaultModel = pick
		updateProvider(&cfg, *anthropicCfg)

		enabledIDs = append(enabledIDs, "anthropic")
		enabledModels = append(enabledModels, anthropicCfg.DefaultModel)
	} else {
		disableProvider(&cfg, "anthropic")
	}

	// Ollama (local) configuration
	if enableOllama {
		fmt.Println("\n\033[36m>> Cấu hình Ollama (local):\033[0m")
		fmt.Println("   Cài đặt Ollama tại: https://ollama.com")
		baseURL := promptString(reader, "Ollama base URL", "http://localhost:11434/v1")

		ollamaCfg := findOrCreateProvider(&cfg, "ollama", "openai-compatible")
		ollamaCfg.Enabled = true
		ollamaCfg.BaseURL = baseURL
		ollamaCfg.APIKeyEnv = ""

		// Fetch live from Ollama /models
		pick := fetchAndPickModel(baseURL, "", "", "qwen2.5-coder:7b")
		ollamaCfg.DefaultModel = pick
		updateProvider(&cfg, *ollamaCfg)

		enabledIDs = append(enabledIDs, "ollama")
		enabledModels = append(enabledModels, ollamaCfg.DefaultModel)
	} else {
		disableProvider(&cfg, "ollama")
	}

	// --- 3. ROUTER CONFIG ---
	fmt.Println("\n\033[1;33m--- 3. Định tuyến Mặc định (Default Routing) ---\033[0m")

	// Automatically set default provider/model based on selection
	if len(enabledIDs) == 1 {
		cfg.Router.DefaultProvider = enabledIDs[0]
		cfg.Router.DefaultModel = enabledModels[0]
		fmt.Printf("\033[32m[ok] Chỉ có 1 provider được bật. Tự động chọn Provider mặc định: %s, Model mặc định: %s\033[0m\n", enabledIDs[0], enabledModels[0])
	} else if len(enabledIDs) > 1 {
		defaultIdx, selectErr := promptSingleSelect("Chọn LLM Provider mặc định cho hội thoại chính", enabledIDs)
		if selectErr != nil {
			return selectErr
		}
		cfg.Router.DefaultProvider = enabledIDs[defaultIdx]
		cfg.Router.DefaultModel = enabledModels[defaultIdx]
		fmt.Printf("\033[32m[ok] Đã chọn Provider mặc định: %s, Model mặc định: %s\033[0m\n", cfg.Router.DefaultProvider, cfg.Router.DefaultModel)
	} else {
		cfg.Router.DefaultProvider = "mock"
		cfg.Router.DefaultModel = "mock-small"
		fmt.Println("\033[31m[!] Không có provider nào được bật. Hệ thống sẽ tự động dùng Mock Provider.\033[0m")
	}

	// Select Intent Mode & Agent Routing
	modes := []string{"hybrid", "llm", "rule"}
	intentIdx, _ := promptSingleSelect("Chọn chế độ phân loại Intent (Intent Mode)", modes)
	cfg.Router.IntentMode = modes[intentIdx]

	agentIdx, _ := promptSingleSelect("Chọn chế độ phân phối Agent (Agent Routing)", modes)
	cfg.Router.AgentRouting = modes[agentIdx]

	// --- 4. SHELL EXECUTION SANDBOX ---
	fmt.Println("\n\033[1;33m--- 4. Quyền Thực thi Lệnh hệ thống (Shell Exec) ---\033[0m")
	cfg.Tools.Shell.Enabled = promptBool(reader, "Bật công cụ chạy shell (shell_exec)?", false)
	if cfg.Tools.Shell.Enabled {
		useSandbox := promptBool(reader, "Chạy lệnh trong môi trường cách ly Docker Sandbox?", true)
		if useSandbox {
			cfg.Tools.Shell.Sandbox = "docker"
			cfg.Tools.Shell.DockerImage = promptString(reader, "Docker Image sử dụng", "alpine:3.20")
			cfg.Tools.Shell.WorkspaceMode = promptString(reader, "Quyền ghi thư mục làm việc (ro: chỉ đọc / rw: đọc-ghi)", "ro")
		} else {
			cfg.Tools.Shell.Sandbox = "none"
		}
	}

	// --- 5. CHANNELS BOTS CONFIG ---
	fmt.Println("\n\033[1;33m--- 5. Cấu hình Chat Bots (Tùy chọn) ---\033[0m")

	// Telegram Bot
	cfg.Channels.Telegram.Enabled = promptBool(reader, "Kích hoạt Telegram Bot?", false)
	if cfg.Channels.Telegram.Enabled {
		token := promptSecret(reader, "Nhập Telegram Bot Token")
		if token != "" {
			envVars["VIETCLAW_TELEGRAM_TOKEN"] = token
		}
		cfg.Channels.Telegram.TokenEnv = "VIETCLAW_TELEGRAM_TOKEN"
	}

	// Discord Bot
	cfg.Channels.Discord.Enabled = promptBool(reader, "Kích hoạt Discord Bot?", false)
	if cfg.Channels.Discord.Enabled {
		token := promptSecret(reader, "Nhập Discord Token")
		if token != "" {
			envVars["VIETCLAW_DISCORD_TOKEN"] = token
		}
		cfg.Channels.Discord.TokenEnv = "VIETCLAW_DISCORD_TOKEN"
	}

	// --- 6. SAVE CONFIG & ENV ---
	fmt.Println("\n\033[1;33m--- 6. Lưu Cấu hình & Khởi tạo Môi trường ---\033[0m")

	// Create directories
	if err := os.MkdirAll(paths.LogDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.Agent.Workspace, 0o755); err != nil {
		return err
	}

	// Save config.json
	cfg = config.MergeDefault(cfg, config.Default(paths))
	if err := config.Save(paths.ConfigFile, cfg); err != nil {
		return fmt.Errorf("lưu file config thất bại: %w", err)
	}
	fmt.Printf("\033[32m[ok] Đã lưu file cấu hình: %s\033[0m\n", paths.ConfigFile)

	// Save .env file in the data dir
	envPath := filepath.Join(paths.DataDir, ".env")
	if err := writeEnvFile(envPath, envVars); err != nil {
		return fmt.Errorf("lưu file .env thất bại: %w", err)
	}
	fmt.Printf("\033[32m[ok] Đã lưu các khóa API vào: %s\033[0m\n", envPath)

	// Initialize SQLite Database
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		return fmt.Errorf("lập lược đồ CSDL thất bại: %w", err)
	}
	fmt.Printf("\033[32m[ok] Khởi tạo cơ sở dữ liệu SQLite thành công tại: %s\033[0m\n", cfg.Database.Path)

	fmt.Println("\n\033[1;32m==================================================\033[0m")
	fmt.Println("\033[1;32m       🎉 THIẾT LẬP VIETCLAW THÀNH CÔNG!         \033[0m")
	fmt.Println("\033[1;32m==================================================\033[0m")
	fmt.Println("VietClaw đã sẵn sàng hoạt động với các cấu hình của bạn.")
	fmt.Println("Để khởi động máy chủ nền (Daemon Server), hãy chạy lệnh:")
	fmt.Println("\n    \033[1;37mvietclaw daemon\033[0m")
	fmt.Println("==================================================")

	return nil
}

func promptString(reader *bufio.Reader, question string, defaultValue string) string {
	fmt.Printf("\033[36m%s\033[0m [\033[1;37m%s\033[0m]: ", question, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptSecret(reader *bufio.Reader, question string) string {
	fmt.Printf("\033[36m%s\033[0m: ", question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptBool(reader *bufio.Reader, question string, defaultValue bool) bool {
	defaultStr := "y"
	if !defaultValue {
		defaultStr = "n"
	}
	fmt.Printf("\033[36m%s\033[0m (\033[1;37my/n\033[0m) [\033[1;37m%s\033[0m]: ", question, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

func promptInt(reader *bufio.Reader, question string, defaultValue int) int {
	fmt.Printf("\033[36m%s\033[0m [\033[1;37m%d\033[0m]: ", question, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("\033[31mLỗi: nhập số nguyên hợp lệ. Sử dụng mặc định: %d\033[0m\n", defaultValue)
		return defaultValue
	}
	return val
}

func findOrCreateProvider(cfg *config.Config, id, providerType string) *config.ProviderConfig {
	for i, p := range cfg.Providers {
		if p.ID == id {
			return &cfg.Providers[i]
		}
	}
	newProvider := config.ProviderConfig{
		ID:   id,
		Type: providerType,
	}
	cfg.Providers = append(cfg.Providers, newProvider)
	return &cfg.Providers[len(cfg.Providers)-1]
}

func updateProvider(cfg *config.Config, provider config.ProviderConfig) {
	for i, p := range cfg.Providers {
		if p.ID == provider.ID {
			cfg.Providers[i] = provider
			return
		}
	}
	cfg.Providers = append(cfg.Providers, provider)
}

func disableProvider(cfg *config.Config, id string) {
	for i, p := range cfg.Providers {
		if p.ID == id {
			cfg.Providers[i].Enabled = false
			return
		}
	}
}

func writeEnvFile(path string, vars map[string]string) error {
	existing := make(map[string]string)
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existing[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		file.Close()
	}

	for k, v := range vars {
		existing[k] = v
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for k, v := range existing {
		if _, err := writer.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func readKey() (string, error) {
	var buf [16]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}
	if n == 1 {
		switch buf[0] {
		case 13, 10:
			return "enter", nil
		case 32:
			return "space", nil
		case 3:
			return "ctrlc", nil
		case 27:
			return "escape", nil
		}
	}
	if n >= 2 && buf[0] == 224 {
		switch buf[1] {
		case 72:
			return "up", nil
		case 80:
			return "down", nil
		}
	}
	if n >= 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return "up", nil
		case 66:
			return "down", nil
		}
	}
	return "", nil
}

func promptMultiSelect(question string, options []string, defaults []bool) ([]bool, error) {
	cleanup, err := setTerminalRaw()
	if err != nil {
		// Fallback if raw mode fails
		return defaults, nil
	}
	defer cleanup()

	cursor := 0
	selected := make([]bool, len(options))
	copy(selected, defaults)

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	fmt.Printf("\033[1;36m%s\033[0m (di chuyển ↑/↓, Space để chọn, Enter để tiếp tục)\n", question)
	printedLines := 0

	render := func() {
		if printedLines > 0 {
			fmt.Print(strings.Repeat("\033[F\033[K", printedLines))
		}
		printedLines = 0
		for i, opt := range options {
			checkbox := "[ ]"
			if selected[i] {
				checkbox = "[\033[32m*\033[0m]"
			}
			prefix := "  "
			if i == cursor {
				prefix = "\033[33m> \033[0m"
			}
			fmt.Printf("%s%s %s\n", prefix, checkbox, opt)
			printedLines++
		}
	}

	render()

	for {
		key, err := readKey()
		if err != nil {
			return nil, err
		}
		switch key {
		case "up":
			cursor--
			if cursor < 0 {
				cursor = len(options) - 1
			}
			render()
		case "down":
			cursor++
			if cursor >= len(options) {
				cursor = 0
			}
			render()
		case "space":
			selected[cursor] = !selected[cursor]
			render()
		case "enter":
			return selected, nil
		case "ctrlc", "escape":
			os.Exit(0)
		}
	}
}

func promptSingleSelect(question string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, nil
	}
	cleanup, err := setTerminalRaw()
	if err != nil {
		return 0, nil
	}
	defer cleanup()

	cursor := 0
	fmt.Printf("\033[1;36m%s\033[0m (di chuyển ↑/↓, Enter để chọn)\n", question)
	printedLines := 0

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	render := func() {
		if printedLines > 0 {
			fmt.Print(strings.Repeat("\033[F\033[K", printedLines))
		}
		printedLines = 0
		for i, opt := range options {
			prefix := "  "
			if i == cursor {
				prefix = "\033[33m> \033[0m"
			}
			fmt.Printf("%s%s\n", prefix, opt)
			printedLines++
		}
	}

	render()

	for {
		key, err := readKey()
		if err != nil {
			return -1, err
		}
		switch key {
		case "up":
			cursor--
			if cursor < 0 {
				cursor = len(options) - 1
			}
			render()
		case "down":
			cursor++
			if cursor >= len(options) {
				cursor = 0
			}
			render()
		case "enter":
			return cursor, nil
		case "ctrlc", "escape":
			os.Exit(0)
		}
	}
}

// fetchAndPickModel fetches live model list from baseURL/models.
// apiKeyDirect is used directly if non-empty (stored temporarily in env so FetchZenModels can pick it up).
// Falls back to fallback if fetch fails or the list is empty.
func fetchAndPickModel(baseURL, apiKeyEnv, apiKeyDirect, fallback string) string {
	fmt.Print("\033[36m  Đang tải danh sách model...\033[0m")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if apiKeyDirect != "" && apiKeyEnv != "" {
		_ = os.Setenv(apiKeyEnv, apiKeyDirect)
	}

	models, err := providers.FetchZenModels(ctx, baseURL, apiKeyEnv)
	fmt.Print("\r\033[K") // clear loading line

	if err != nil || len(models) == 0 {
		if err != nil {
			fmt.Printf("  \033[33m[!] Không lấy được danh sách model: %v\033[0m\n", err)
		} else {
			fmt.Println("  \033[33m[!] Danh sách model trống\033[0m")
		}
		return fallback
	}

	fmt.Printf("  \033[32m[ok] Lấy được %d model\033[0m\n", len(models))
	return pickModelFromList(models, fallback, "Chọn model mặc định")
}

// pickModelFromList shows an interactive ↑/↓ select over a list of model names.
func pickModelFromList(models []string, defaultModel, question string) string {
	if len(models) == 0 {
		return defaultModel
	}

	start := 0
	for i, m := range models {
		if m == defaultModel {
			start = i
			break
		}
	}

	cleanup, err := setTerminalRaw()
	if err != nil {
		return defaultModel
	}
	defer cleanup()

	cursor := start
	fmt.Printf("\033[1;36m%s\033[0m (↑/↓ di chuyển, Enter để chọn)\n", question)
	printedLines := 0

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	render := func() {
		if printedLines > 0 {
			fmt.Print(strings.Repeat("\033[F\033[K", printedLines))
		}
		printedLines = 0

		winStart := cursor - 4
		if winStart < 0 {
			winStart = 0
		}
		winEnd := winStart + 10
		if winEnd > len(models) {
			winEnd = len(models)
			winStart = winEnd - 10
			if winStart < 0 {
				winStart = 0
			}
		}

		if winStart > 0 {
			fmt.Printf("  \033[2m... %d model phía trên\033[0m\n", winStart)
			printedLines++
		}
		for i := winStart; i < winEnd; i++ {
			prefix := "  "
			suffix := ""
			if i == cursor {
				prefix = "\033[33m> \033[0m"
				suffix = " \033[32m✓\033[0m"
			}
			fmt.Printf("%s%s%s\n", prefix, models[i], suffix)
			printedLines++
		}
		if winEnd < len(models) {
			fmt.Printf("  \033[2m... %d model phía dưới\033[0m\n", len(models)-winEnd)
			printedLines++
		}
	}

	render()

	for {
		key, err := readKey()
		if err != nil {
			return models[cursor]
		}
		switch key {
		case "up":
			if cursor > 0 {
				cursor--
			}
			render()
		case "down":
			if cursor < len(models)-1 {
				cursor++
			}
			render()
		case "enter":
			fmt.Printf("\033[32m  ✓ Đã chọn: %s\033[0m\n", models[cursor])
			return models[cursor]
		case "ctrlc", "escape":
			os.Exit(0)
		}
	}
}
