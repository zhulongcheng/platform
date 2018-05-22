package http

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Server is an abstraction around the http.Server that handles a server process.
type Server struct {
	srv      *http.Server
	signalCh chan os.Signal
	wg       sync.WaitGroup
}

// NewServer returns a new server struct that can be used.
func NewServer(handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Handler: handler,
		},
	}
}

func (s *Server) Serve(listener net.Listener) error {
	// When we return, wait for all pending goroutines to finish.
	defer s.wg.Wait()

	errCh := s.serve(listener)
	select {
	case err := <-errCh:
		// The server has failed and reported an error.
		return err
	case <-s.signalCh:
		// We have received an interrupt. Signal the shutdown process.
		return s.shutdown()
	}
}

func (s *Server) serve(listener net.Listener) <-chan error {
	s.wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		defer s.wg.Done()
		if err := s.srv.Serve(listener); err != nil {
			errCh <- err
		}
		close(errCh)
	}()
	return errCh
}

func (s *Server) shutdown() error {
	// The shutdown needs to succeed in 20 seconds or less.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Wait for another signal to cancel the shutdown.
	done := make(chan struct{})
	defer close(done)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		select {
		case <-s.signalCh:
			cancel()
		case <-done:
		}
	}()
	return s.srv.Shutdown(ctx)
}

func (s *Server) ListenForSignals(signals ...os.Signal) {
	if s.signalCh == nil {
		s.signalCh = make(chan os.Signal, 4)
	}
	signal.Notify(s.signalCh, signals...)
}
