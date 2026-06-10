package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/db"
	"vietclaw/internal/harness"
)

func runHarness(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("harness command is required: create|list|show")
	}
	_, cfg, database, cleanup, err := localDatabase()
	if err != nil {
		return err
	}
	defer cleanup()
	service := harness.New(cfg, database)
	ctx := context.Background()

	switch args[0] {
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("harness create goal is required")
		}
		run, err := service.Create(ctx, harness.CreateRequest{Goal: strings.Join(args[1:], " ")})
		if err != nil {
			return err
		}
		return printJSON(run)
	case "list":
		runs, err := service.List(ctx, 20)
		if err != nil {
			return err
		}
		for _, run := range runs {
			fmt.Printf("%s [%s/%s] %s\n", run.ID, run.Status, run.Risk, run.Goal)
		}
		return nil
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("harness show run_id is required")
		}
		detail, err := service.Detail(ctx, args[1])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("harness run not found: %s", args[1])
		}
		if err != nil {
			return err
		}
		return printJSON(detail)
	default:
		return fmt.Errorf("unknown harness command %q", args[0])
	}
}

func localDatabase() (config.Paths, config.Config, *sql.DB, func(), error) {
	paths, cfg, err := loadOrCreateConfig()
	if err != nil {
		return config.Paths{}, config.Config{}, nil, nil, err
	}
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return config.Paths{}, config.Config{}, nil, nil, err
	}
	if err := db.ApplySchema(database); err != nil {
		_ = database.Close()
		return config.Paths{}, config.Config{}, nil, nil, err
	}
	return paths, cfg, database, func() { _ = database.Close() }, nil
}

func printJSON(value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}
