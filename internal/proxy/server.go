package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/handlers"
)

type Server struct {
	config        *config.Config
	deviceID      string
	creatorAddr   string
	port          int
	requestRouter *handlers.RequestRouter
}

func NewServer(cfg *config.Config, deviceID, creatorAddr string, port int) *Server {
	return &Server{
		config:        cfg,
		deviceID:      deviceID,
		creatorAddr:   creatorAddr,
		port:          port,
		requestRouter: handlers.NewRequestRouter(cfg, deviceID, creatorAddr),
	}
}

func (s *Server) Start() error {
	log := gologger.Get().With().Str("component", "proxy").Logger()

	localAddr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.port)
	localBaseURL := handlers.LocalProxyBaseURL(s.config.Server.Host, s.port)
	upstreamBaseURL := handlers.NormalizeServerBaseURL(s.config.Runner.ServerURL)

	log.Info().
		Str("address", localAddr).
		Str("local_proxy_url", localBaseURL).
		Str("dashboard_url", fmt.Sprintf("%s/dashboard", localBaseURL)).
		Str("health_url", fmt.Sprintf("%s/health", localBaseURL)).
		Str("info_url", fmt.Sprintf("%s/", localBaseURL)).
		Str("registry_task_url", fmt.Sprintf("%s/api/v1/tasks", localBaseURL)).
		Str("local_upload_url", fmt.Sprintf("%s/api/v1/tasks/local", localBaseURL)).
		Str("upstream_server_url", upstreamBaseURL).
		Str("device_id", s.deviceID).
		Str("creator_address", s.creatorAddr).
		Int("port", s.port).
		Msg("Starting Parity client proxy")

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.requestRouter.HandleRequest)

	server := &http.Server{
		Addr:              localAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server.ListenAndServe()
}
