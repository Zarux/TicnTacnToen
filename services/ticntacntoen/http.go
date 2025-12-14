package ticntacntoen

import (
	"net/http"
	"time"

	"github.com/Zarux/ticntacntoen/internal/logger"
)

type httpHandler struct {
	svc *Service
}

func HTTPHandler(s *Service) http.Handler {
	h := &httpHandler{
		svc: s,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /{gameID}/moves/", h.HandleNewMove)
	mux.HandleFunc("POST /", h.HandleNewGame)

	return mux
}

type moveRequest struct {
	X            int           `json:"x"`
	Y            int           `json:"y"`
	Hash         uint64        `json:"hash"`
	ThinkingTime time.Duration `json:"thinkingTime"`
}

type board struct {
	State  []int8 `json:"state"`
	Hash   uint64 `json:"hash"`
	Winner int8   `json:"winner,omitempty"`
}

func (h *httpHandler) HandleNewMove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.Info("yo")
}

type newGameRequest struct {
	N      int `json:"n"`
	K      int `json:"k"`
	Player int `json:"player"`
}

func (h *httpHandler) HandleNewGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.Info("yo")
}
