package main

import (
	"encoding/json"
	"fmt"

	"vietclaw/internal/channels"
	_ "vietclaw/internal/channels/discord"
	_ "vietclaw/internal/channels/telegram"
	"vietclaw/internal/framework"
	"vietclaw/internal/plugins"
)

func runFramework(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return printFrameworkCatalog()
	}
	return fmt.Errorf("usage: vietclaw framework list")
}

func printFrameworkCatalog() error {
	catalog := map[string]any{
		"extensions": plugins.BuiltinRegistry(),
		"channels":   channels.RegisteredAdapters(),
		"hooks": []string{
			string(framework.EventBeforeChat),
			string(framework.EventAfterChat),
			string(framework.EventBeforeTool),
			string(framework.EventAfterTool),
			string(framework.EventRunStart),
			string(framework.EventRunFinish),
		},
	}
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
