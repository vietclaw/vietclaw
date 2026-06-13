package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
)

const (
	toolFileRead  = "file_read"
	toolFileWrite = "file_write"
	toolShellExec = "shell_exec"
)

type ToolRegistry struct {
	policy      Policy
	cfg         config.Config
	mem         *memory.Store
	agents      *agentfs.Registry
	tools       map[string]Tool
	mcp         map[string]mcpToolRef
	defs        []providers.ToolDefinition
	extraDefs   []providers.ToolDefinition
	agentCustom map[string][]providers.ToolDefinition
}

type mcpToolRef struct {
	client *MCPClient
	name   string
}

func NewRegistry(cfg config.Config) *ToolRegistry {
	p := NewPolicy(cfg)
	r := &ToolRegistry{
		policy:      p,
		cfg:         cfg,
		tools:       make(map[string]Tool),
		mcp:         make(map[string]mcpToolRef),
		agentCustom: make(map[string][]providers.ToolDefinition),
	}
	r.tools[toolFileRead] = FileRead{Policy: p}
	r.tools[toolFileWrite] = FileWrite{Policy: p}
	r.tools[toolShellExec] = ShellExec{Policy: p}

	r.tools["web_search"] = WebSearch{}
	r.tools["web_fetch"] = WebFetch{Policy: p}
	r.tools["dir_list"] = DirList{Policy: p}
	r.tools["file_grep"] = FileGrep{Policy: p}
	r.tools["file_find"] = FileFind{Policy: p}
	r.tools["system_info"] = SystemInfo{}
	r.tools["network_ping"] = NetworkPing{}
	r.tools["env_get"] = EnvGet{}
	r.tools["hash_calc"] = HashCalc{Policy: p}
	r.tools["json_format"] = JSONFormat{}
	r.tools["string_transform"] = StringTransform{}
	r.tools["time_current"] = TimeCurrent{}
	r.tools["math_calc"] = MathCalc{}
	r.tools["process_list"] = ProcessList{}
	r.tools["ip_lookup"] = IPLookup{}
	registerExtraTools(r.tools, p)

	r.discoverMCP(context.Background())
	return r
}

func (r *ToolRegistry) Register(tool Tool, def providers.ToolDefinition) {
	name := normalizeToolName(def.Function.Name)
	if name == "" {
		name = normalizeToolName(tool.Name())
		def.Function.Name = name
	}
	r.tools[name] = tool
	r.extraDefs = append(r.extraDefs, def)
}

func (r *ToolRegistry) WithAgentRegistry(registry *agentfs.Registry) *ToolRegistry {
	r.agents = registry
	r.ReloadAgentTools(registry)
	return r
}

func (r *ToolRegistry) ReloadAgentTools(registry *agentfs.Registry) {
	if registry == nil {
		return
	}
	custom := make(map[string][]providers.ToolDefinition)
	for _, def := range registry.List() {
		for _, tool := range def.CustomTools {
			name := normalizeToolName(tool.Name)
			params := tool.Parameters
			if params == nil {
				params = map[string]any{"type": "object", "properties": map[string]any{}}
			}
			custom[def.ID] = append(custom[def.ID], providers.ToolDefinition{
				Type: "function",
				Function: providers.FunctionDetail{
					Name:        name,
					Description: tool.Description,
					Parameters:  params,
				},
			})
			r.tools[name] = customToolRunner{
				registry: registry,
				agentID:  def.ID,
				tool:     tool,
				policy:   r.policy,
				cfg:      r.cfg,
				mcp:      r.mcp,
			}
		}
	}
	r.agentCustom = custom
}

func (r *ToolRegistry) WithMemory(mem *memory.Store) *ToolRegistry {
	if mem == nil {
		return r
	}
	r.mem = mem
	if r.cfg.Agent.MemoryTools.Enabled {
		r.tools[toolMemoryRecall] = MemoryRecall{Store: mem}
		r.tools[toolMemoryStore] = MemoryStore{Store: mem}
	}
	return r
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, argsJSON string) (string, error) {
	normalized := normalizeToolName(name)
	t, ok := r.tools[normalized]
	if !ok {
		if ref, ok := r.mcp[normalized]; ok {
			return ref.client.Execute(ctx, ref.name, argsJSON)
		}
		return "", fmt.Errorf("tool not found: %s", name)
	}

	switch normalized {
	case toolFileRead:
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return t.Run(ctx, argsJSON)
		}
		return t.Run(ctx, args.Path)
	case toolFileWrite:
		var args struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", err
		}
		return t.Run(ctx, args.Path+"\n"+args.Content)
	case toolShellExec:
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return t.Run(ctx, argsJSON)
		}
		return t.Run(ctx, args.Command)
	case "web_search":
		return r.runWebSearch(ctx, argsJSON)
	default:
		return t.Run(ctx, argsJSON)
	}
}

func (r *ToolRegistry) runWebSearch(ctx context.Context, argsJSON string) (string, error) {
	if ref, ok := r.mcpOpenWebSearch(); ok {
		out, err := ref.client.Execute(ctx, ref.name, argsJSON)
		if err == nil && strings.TrimSpace(out) != "" && !isEmptySearchPayload(out) {
			return out, nil
		}
	}
	tool, ok := r.tools["web_search"]
	if !ok {
		return "", fmt.Errorf("tool not found: web_search")
	}
	return tool.Run(ctx, argsJSON)
}

func (r *ToolRegistry) mcpOpenWebSearch() (mcpToolRef, bool) {
	ref, ok := r.mcp["mcp_open_websearch_search"]
	return ref, ok
}

func isEmptySearchPayload(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "[]" || trimmed == "{}" {
		return true
	}
	if strings.Contains(trimmed, "No web results found") || strings.Contains(trimmed, "no results") {
		return true
	}
	return false
}

func (r *ToolRegistry) GetDefinitions() []providers.ToolDefinition {
	return r.GetDefinitionsForProfile(config.AgentProfileConfig{}, true)
}

func (r *ToolRegistry) GetDefinitionsForProfile(profile config.AgentProfileConfig, includeDelegate bool) []providers.ToolDefinition {
	lang := profile.Language
	if lang == "" {
		lang = r.cfg.Agent.Language
	}
	var list []providers.ToolDefinition

	if r.cfg.Tools.Files.Enabled {
		list = append(list, providers.ToolDefinition{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        toolFileRead,
				Description: i18n.T(lang, i18n.ToolFileRead),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": i18n.T(lang, i18n.ToolPathParam),
						},
					},
					"required": []string{"path"},
				},
			},
		})

		list = append(list, providers.ToolDefinition{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        toolFileWrite,
				Description: i18n.T(lang, i18n.ToolFileWrite),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": i18n.T(lang, i18n.ToolPathParam),
						},
						"content": map[string]any{
							"type":        "string",
							"description": i18n.T(lang, i18n.ToolContentParam),
						},
					},
					"required": []string{"path", "content"},
				},
			},
		})
	}

	if r.cfg.Tools.Shell.Enabled {
		list = append(list, providers.ToolDefinition{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        toolShellExec,
				Description: i18n.T(lang, i18n.ToolShellExec),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]any{
							"type":        "string",
							"description": i18n.T(lang, i18n.ToolCommandParam),
						},
					},
					"required": []string{"command"},
				},
			},
		})
	}

	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "web_search",
			Description: i18n.T(lang, i18n.ToolWebSearch),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": i18n.T(lang, i18n.ToolQueryParam),
					},
				},
				"required": []string{"query"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "web_fetch",
			Description: i18n.T(lang, i18n.ToolWebFetch),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": i18n.T(lang, i18n.ToolURLParam),
					},
				},
				"required": []string{"url"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "dir_list",
			Description: "List files and folders in a directory path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The directory path to list (defaults to workspace root).",
					},
				},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "file_grep",
			Description: "Search file contents for lines matching a pattern.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The path to file or directory to search.",
					},
					"pattern": map[string]any{
						"type":        "string",
						"description": "The regular expression pattern to search for.",
					},
				},
				"required": []string{"path", "pattern"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "file_find",
			Description: "Find files matching a glob pattern under a path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The base path to start the search.",
					},
					"pattern": map[string]any{
						"type":        "string",
						"description": "The glob pattern matching filenames (e.g. *.txt).",
					},
				},
				"required": []string{"path", "pattern"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "system_info",
			Description: "Get current host OS name, architecture, CPU count, etc.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "network_ping",
			Description: "Ping a remote host to test connectivity.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"host": map[string]any{
						"type":        "string",
						"description": "The IP address or domain name to ping.",
					},
				},
				"required": []string{"host"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "env_get",
			Description: "Read an environment variable from the system.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "The environment variable key name.",
					},
				},
				"required": []string{"key"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "hash_calc",
			Description: "Calculate MD5, SHA-1, or SHA-256 hash of a file.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The file path.",
					},
					"algo": map[string]any{
						"type":        "string",
						"description": "The hash algorithm: md5, sha1, or sha256 (default).",
					},
				},
				"required": []string{"path"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "json_format",
			Description: "Pretty print or compact JSON text.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": "The JSON string to format.",
					},
					"minify": map[string]any{
						"type":        "boolean",
						"description": "Set true to compact/minify JSON instead of pretty printing.",
					},
				},
				"required": []string{"text"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "string_transform",
			Description: "Perform url-encode/decode, base64-encode/decode, or casing.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": "The string to transform.",
					},
					"op": map[string]any{
						"type":        "string",
						"description": "Operation: base64_encode, base64_decode, url_encode, url_decode, upper, lower.",
					},
				},
				"required": []string{"text", "op"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "time_current",
			Description: "Get the current date and time in local, UTC, or a specific timezone.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tz": map[string]any{
						"type":        "string",
						"description": "Optional IANA timezone name (e.g. America/New_York).",
					},
				},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "math_calc",
			Description: "Evaluate basic mathematical expressions.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"expr": map[string]any{
						"type":        "string",
						"description": "The arithmetic expression to evaluate.",
					},
				},
				"required": []string{"expr"},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "process_list",
			Description: "Retrieve a list of active processes on the host.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	})
	list = append(list, providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        "ip_lookup",
			Description: "Query the external public IP and geolocation of the host.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	})

	if r.cfg.Agent.MemoryTools.Enabled && r.mem != nil {
		list = append(list, memoryToolDefinitions(r.cfg.Agent.Language)...)
	}
	list = append(list, extraToolDefinitions()...)
	list = append(list, r.extraDefs...)
	list = append(list, r.defs...)
	if includeDelegate && r.cfg.Framework.Enabled && r.cfg.Framework.DelegateEnabled {
		list = append(list, FrameworkToolDefinitions(r.cfg.Framework.AllowAutoCreate)...)
	}
	if profile.ID != "" {
		list = append(list, r.agentCustom[profile.ID]...)
	}
	if len(profile.Tools) > 0 {
		list = filterDefinitions(list, profile.Tools)
	}
	return list
}

func normalizeToolName(name string) string {
	switch name {
	case "file.read":
		return toolFileRead
	case "file.write":
		return toolFileWrite
	case "shell.exec":
		return toolShellExec
	default:
		if strings.HasPrefix(name, mcpToolPrefix+"_") {
			return sanitizeToolName(name)
		}
		return name
	}
}

func (r *ToolRegistry) discoverMCP(ctx context.Context) {
	for _, server := range r.cfg.Tools.MCP {
		if !server.Enabled {
			continue
		}
		client := NewMCPClient(server)
		discovered, err := client.Discover(ctx)
		if err != nil {
			continue
		}
		for _, tool := range discovered {
			name := tool.Definition.Function.Name
			r.mcp[name] = mcpToolRef{client: client, name: tool.Name}
			r.defs = append(r.defs, tool.Definition)
		}
	}
}
