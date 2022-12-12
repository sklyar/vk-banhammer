package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/sklyar/vk-banhammer/internal/entity"
)

type service interface {
	// CheckComment checks comment and ban user if needed.
	CheckComment(comment *entity.Comment) (entity.BanReason, error)
}

// Server is a banhammer HTTP server.
// It handles VK callback API requests.
type Server struct {
	service                  service
	callbackConfirmationCode string

	srv *http.Server

	logger *zap.Logger
}

// NewServer creates a new banhammer HTTP server.
func NewServer(logger *zap.Logger, addr string, service service, callbackConfirmationCode string) *Server {
	mux := http.NewServeMux()

	srv := &Server{
		service:                  service,
		callbackConfirmationCode: callbackConfirmationCode,
		srv: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		logger: logger,
	}

	mux.HandleFunc("/new_message", srv.gatewayHandler)

	return srv
}

type request struct {
	GroupID int             `json:"group_id"`
	EventID string          `json:"event_id"`
	Type    string          `json:"type"`
	Object  json.RawMessage `json:"object"`

	HTTPRequest *http.Request `json:"-"`
}

func (s *Server) gatewayHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode message", zap.Error(err))
		_, _ = w.Write([]byte("ok"))
		return
	}
	req.HTTPRequest = r

	switch req.Type {
	case "confirmation":
		s.confirmationHandler(w, &req)
	case "wall_reply_new":
		s.wallReplyNewHandler(w, &req)
	default:
		s.unknownTypeHandler(w, &req)
	}
}

func (s *Server) confirmationHandler(w http.ResponseWriter, _ *request) {
	_, _ = w.Write([]byte(s.callbackConfirmationCode))
}

func (s *Server) wallReplyNewHandler(w http.ResponseWriter, r *request) {
	var comment entity.Comment
	if err := json.Unmarshal(r.Object, &comment); err != nil {
		log.Printf("failed to unmarshal comment: %v", err)
		return
	}

	reason, err := s.service.CheckComment(&comment)
	if err != nil {
		log.Printf("failed to check comment: %v", err)
		_, _ = w.Write([]byte("ok"))
		return
	}

	log.Printf("comment checked: %v", reason)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) unknownTypeHandler(w http.ResponseWriter, msgRequest *request) {
	log.Printf("unknown type: %s", msgRequest.Type)
	_, _ = w.Write([]byte("ok"))
}

// ListenAndServe starts HTTP server.
// It blocks until the context is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.srv.ListenAndServe()
	}()

	s.logger.Info("server started", zap.String("addr", s.srv.Addr))

	select {
	case <-ctx.Done():
		return s.srv.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}
