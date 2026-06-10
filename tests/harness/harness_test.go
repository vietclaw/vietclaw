package harness_test

import (
	"context"
	"path/filepath"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/db"
	"vietclaw/internal/harness"
)

func TestCreateHarnessRunStoresCapsuleAndEvents(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	service := harness.New(cfg, database)
	run, err := service.Create(context.Background(), harness.CreateRequest{
		Goal: "fix failing auth test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.ID == "" || run.Status != harness.StatusPlanned {
		t.Fatalf("unexpected run: %#v", run)
	}
	if run.Mode != "agentless" || run.Risk != "low" {
		t.Fatalf("unexpected capsule defaults: %#v", run)
	}
	if len(run.Plan.Steps) == 0 {
		t.Fatalf("expected plan steps: %#v", run.Plan)
	}

	detail, err := service.Detail(context.Background(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Run.ID != run.ID {
		t.Fatalf("detail mismatch: %#v", detail)
	}
	if len(detail.Events) < 2 {
		t.Fatalf("expected ledger events, got %#v", detail.Events)
	}
}

func TestHarnessRiskGateMarksDangerousGoalsHigh(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	service := harness.New(cfg, database)
	run, err := service.Create(context.Background(), harness.CreateRequest{
		Goal: "deploy to production and push git changes",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.Risk != "high" {
		t.Fatalf("risk = %q, want high", run.Risk)
	}
}
