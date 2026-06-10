CREATE TABLE IF NOT EXISTS harness_runs (
  id TEXT PRIMARY KEY,
  session_id TEXT,
  goal TEXT NOT NULL,
  mode TEXT NOT NULL,
  risk TEXT NOT NULL,
  status TEXT NOT NULL,
  budget_json TEXT NOT NULL,
  allowed_tools_json TEXT NOT NULL,
  forbidden_tools_json TEXT NOT NULL,
  success_checks_json TEXT NOT NULL,
  provider TEXT,
  model TEXT,
  summary TEXT,
  plan_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS harness_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL,
  type TEXT NOT NULL,
  payload TEXT NOT NULL,
  created_at TEXT NOT NULL
);
