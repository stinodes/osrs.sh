package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"osrs.sh/wiki/ssh/src/config"
	"osrs.sh/wiki/ssh/src/views/layout"
)

func main() {
	config, err := config.LoadAppConfig()
	if err != nil {
		log.Error("No valid config found.", "err", err)
	}

	log.Info("Starting server with config", "config", config)

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(config.Host, config.Port)),
		wish.WithHostKeyPath(config.IDFile),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			logging.Middleware(),
		),
	)

	if err != nil {
		log.Error("Unable to start server.", "err", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Info("Server starting with", "host", config.Host, "port", config.Port)
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Unable to start server.", "err", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Server stopped.")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Unable to stop server.", "err", err)
	}

}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	renderer := bubbletea.MakeRenderer(s)
	return layout.New(renderer), []tea.ProgramOption{tea.WithAltScreen()}
}
