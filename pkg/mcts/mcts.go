package mcts

import (
	"context"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type Client struct {
	explorationParam float64
	nextMoveCache    *sync.Map
	threads          int
	iterations       int
	thinkTime        time.Duration
}

func New(threads, iterationsPerThread int) *Client {
	return &Client{
		explorationParam: 1.414,
		nextMoveCache:    &sync.Map{},
		threads:          threads,
		iterations:       iterationsPerThread,
		thinkTime:        time.Second,
	}
}

func (c *Client) UpdateExploationParam(ep float64) {
	c.explorationParam = ep
}

func (c *Client) UpdateThinkTime(t time.Duration) {
	c.thinkTime = t
}

func (c *Client) GetNextMove(ctx context.Context, rootBoard *tictactoe.Board, player tictactoe.Player) int {
	c.nextMoveCache = &sync.Map{}

	if move, ok := rootBoard.ForcedMove(player); ok {
		return move
	}

	results := make(chan map[int]int, c.threads)
	var wg sync.WaitGroup
	wg.Add(c.threads)
	for range c.threads {
		go func() {
			defer wg.Done()

			root := &node{
				UntriedMoves: rootBoard.LegalMoves(),
				client:       c,
			}

			c.mctsIteration(ctx, c.iterations, root, rootBoard.Clone(), player)

			visitMap := make(map[int]int)
			for _, c := range root.Children {
				visitMap[c.Move] = c.Visits
			}

			results <- visitMap
		}()
	}

	wg.Wait()
	close(results)

	totalVisits := make(map[int]int)
	for r := range results {
		for move, visits := range r {
			totalVisits[move] += visits
		}
	}

	bestMove := -1
	bestVisits := -1
	for move, visits := range totalVisits {
		if visits > bestVisits {
			bestMove = move
			bestVisits = visits
		}
	}

	return bestMove
}

func (c *Client) mctsIteration(ctx context.Context, iterations int, root *node, board *tictactoe.Board, player tictactoe.Player) {
	done := time.After(c.thinkTime)

mctsIteration:
	for range iterations {
		select {
		case <-ctx.Done():
			break mctsIteration
		case <-done:
			break mctsIteration
		default:
		}

		board := board.Clone()
		n := root
		current := player

		// Selection
		for len(n.UntriedMoves) == 0 && len(n.Children) > 0 {
			n = n.selectChild()
			err := board.ApplyMove(n.Move, current)
			if err != nil {
				panic("SELECTION ILLEGAL MOVE")
			}
			current = -current
		}

		// Expansion
		if len(n.UntriedMoves) > 0 && n.canExpand() {
			n = n.expand(board, current)
			current = -current
		}

		// Simulation
		winner := c.rollout(board, current)

		// Backprop
		n.backpropagate(winner)
	}
}

func (c *Client) rollout(board *tictactoe.Board, player tictactoe.Player) tictactoe.Player {
	current := player

	for {
		if winner := board.CheckWinner(); winner != tictactoe.Empty {
			return winner
		}

		if !board.AnyLegalMoves() {
			return tictactoe.Empty
		}

		//move := board.BiasedRandomMoveCache(c.nextMoveCache)
		move := board.BiasedRandomMove()
		err := board.ApplyMove(move, current)
		if err != nil {
			panic("ROLLOUT ILLEGAL MOVE")
		}
		current = -current
	}
}

func Iterations(N, K, t int) int {
	base := N * K * 1000
	alpha := 1.0
	remaining := float64(N*N - t)
	if remaining < 1 {
		remaining = 1
	}

	iters := float64(base) * (1 + alpha*(float64(N*N)/remaining)) * float64(K) / 3.0
	return int(iters)
}

func ExplorationParameter(N, K, turn int) float64 {
	base := 1.414
	factor := math.Sqrt(float64(K) / float64(N))
	remaining := float64(N*N - turn)
	if remaining <= 0 {
		remaining = 1
	}

	turnFactor := math.Min(1.0, float64(N*N)/remaining)
	return base * factor * turnFactor
}

func (c *Client) getNextMove(ctx context.Context, rootBoard *tictactoe.Board, player tictactoe.Player) int {
	c.nextMoveCache = &sync.Map{}

	if move, ok := rootBoard.ForcedMove(player); ok {
		return move
	}

	root := &node{
		UntriedMoves: rootBoard.LegalMoves(),
		client:       c,
	}

	c.mctsIteration(ctx, c.iterations, root, rootBoard.Clone(), player)

	rand.Shuffle(len(root.Children), func(i, j int) {
		root.Children[i], root.Children[j] = root.Children[j], root.Children[i]
	})

	best := root.Children[0]
	for _, c := range root.Children[1:] {
		if c.Visits > best.Visits {
			best = c
		}
	}

	return best.Move
}
