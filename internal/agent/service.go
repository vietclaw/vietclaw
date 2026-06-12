package agent

import (
	"database/sql"
	"log"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
	contextbuilder "vietclaw/internal/context"
	"vietclaw/internal/framework"
	"vietclaw/internal/i18n"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/router"
	"vietclaw/internal/tools"
)

type Service struct {
	cfg       config.Config
	db        *sql.DB
	mem       *memory.Store
	router    *router.ModelRouter
	context   *contextbuilder.Builder
	tools     *tools.ToolRegistry
	agents    *agentfs.Registry
	pool      *RunPool
	framework *framework.Framework
	Logger    *log.Logger
	dataDir   string
}

func NewService(cfg config.Config, db *sql.DB) *Service {
	return NewServiceWithDataDir(cfg, db, "")
}

func NewServiceWithDataDir(cfg config.Config, db *sql.DB, dataDir string) *Service {
	mem := memory.NewStore(db)
	providerList := providers.Enabled(cfg.Providers)
	r := router.NewModelRouter(cfg, db, providerList)
	root := agentfs.DefaultRoot(dataDir)
	if dataDir == "" {
		if paths, err := config.DefaultPaths(); err == nil {
			root = agentfs.DefaultRoot(paths.DataDir)
			dataDir = paths.DataDir
		}
	}
	registry := agentfs.NewRegistry(root, cfg)
	_ = registry.Reload()
	toolReg := tools.NewRegistry(cfg).WithMemory(mem).WithAgentRegistry(registry)
	return &Service{
		cfg:     cfg,
		db:      db,
		mem:     mem,
		router:  r,
		context: contextbuilder.New(cfg, db, mem).WithRouter(r).WithAgentRegistry(registry),
		tools:   toolReg,
		agents:  registry,
		pool:    NewRunPool(cfg.Runtime.MaxConcurrentTasks, cfg.Framework.MaxConcurrentSpawns),
		dataDir: dataDir,
	}
}

func (s *Service) WithLogger(logger *log.Logger) *Service {
	s.Logger = logger
	return s
}

func (s *Service) WithFramework(fw *framework.Framework) *Service {
	s.framework = fw
	return s
}

func (s *Service) Framework() *framework.Framework {
	return s.framework
}

func (s *Service) logf(format string, args ...any) {
	if s.Logger != nil {
		s.Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (s *Service) Memory() *memory.Store {
	return s.mem
}

func (s *Service) Language() string {
	return string(i18n.Normalize(s.cfg.Agent.Language))
}

func (s *Service) text(id i18n.MessageID, args ...any) string {
	return i18n.T(s.cfg.Agent.Language, id, args...)
}

func (s *Service) Router() *router.ModelRouter {
	return s.router
}

func (s *Service) AgentRegistry() *agentfs.Registry {
	return s.agents
}

func (s *Service) ApplyRuntimeConfig(cfg config.Config) {
	s.cfg = cfg
	s.pool.UpdateLimits(cfg.Runtime.MaxConcurrentTasks, cfg.Framework.MaxConcurrentSpawns)
	s.router = router.NewModelRouter(cfg, s.db, providers.Enabled(cfg.Providers))
	s.context = contextbuilder.New(cfg, s.db, s.mem).WithRouter(s.router).WithAgentRegistry(s.agents)
	s.tools = tools.NewRegistry(cfg).WithMemory(s.mem).WithAgentRegistry(s.agents)
}

func (s *Service) ReloadAgents() error {
	return s.agents.Reload()
}
