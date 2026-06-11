package agent

import (
	"database/sql"
	"log"

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
	framework *framework.Framework
	Logger    *log.Logger
}

func NewService(cfg config.Config, db *sql.DB) *Service {
	mem := memory.NewStore(db)
	providerList := providers.Enabled(cfg.Providers)
	r := router.NewModelRouter(cfg, db, providerList)
	return &Service{
		cfg:     cfg,
		db:      db,
		mem:     mem,
		router:  r,
		context: contextbuilder.New(cfg, db, mem).WithRouter(r),
		tools:   tools.NewRegistry(cfg).WithMemory(mem),
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
