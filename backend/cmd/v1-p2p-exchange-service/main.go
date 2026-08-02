package main

import (
	"fmt"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"

	"go.uber.org/fx"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	fmt.Printf("NewServer: Name: %s\n", config.Name)
	fmt.Printf("NewServer: Host: %s\n", config.Host)
	fmt.Printf("NewServer: Port: %d\n", config.Port)
	fmt.Printf("NewServer: Mode: %s\n", config.Mode)
	return &Server{config: config}
}

func (s *Server) Start() {
	fmt.Printf("Server running at %s:%d\n", s.config.Host, s.config.Port)
}

func main() {
	fx.New(
		config.Module,
		fx.Provide(NewServer),
		fx.Invoke(func(s *Server, config *config.Config) {
			fmt.Printf("Name: %s\n", config.Name)
			fmt.Printf("Host: %s\n", config.Host)
			fmt.Printf("Port: %d\n", config.Port)
			fmt.Printf("Mode: %s\n", config.Mode)
			s.Start()
		}),
	).Run()
}
