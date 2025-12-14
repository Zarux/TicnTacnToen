package main

import (
	"context"
	"net/http"
	"slices"

	"github.com/Zarux/ticntacntoen/internal/logger"
	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/services/ticntacntoen"
)

func main() {
	log := logger.New()

	bot := mcts.New(4, 100_000)
	svc := ticntacntoen.New(bot)

	h := ticntacntoen.HTTPHandler(svc)
	handler := rootHandler("/game/v1", h)

	middlewares := []func(http.Handler) http.Handler{
		logger.NewMiddleware(),
	}

	slices.Reverse(middlewares)

	var ok bool
	for _, mw := range middlewares {
		handler, ok = mw(handler).(http.HandlerFunc)
		if !ok {
			panic("bad middleware")
		}
	}

	svc.Play(context.Background())
	return

	addr := "127.0.0.1:3000"
	log.Info("listening on", "addr", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Error(err.Error())
	}
}

func rootHandler(root string, h http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(root+"/", http.StripPrefix(root, h))
	return mux
}
