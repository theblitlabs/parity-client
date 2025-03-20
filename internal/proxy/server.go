package proxy

import (
	"fmt"
	"net/http"

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

	log.Info().
		Str("address", localAddr).
		Str("device_id", s.deviceID).
		Str("creator_address", s.creatorAddr).
		Int("port", s.port).
		Msg("Starting chain proxy server")

	http.HandleFunc("/", s.requestRouter.HandleRequest)
	return http.ListenAndServe(localAddr, nil)
}
