package mcts

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type LastMoveStats struct {
	RealThinkTime    time.Duration
	ActualThinkTime  time.Duration
	NumIterations    int
	BestMove         int
	MoveVisits       int
	MoveWins         float64
	TacticalOverride bool
}

type Client struct {
	explorationParam float64
	workers          int
	iterations       int
	thinkTime        time.Duration
	lastNode         *node

	lastMoveStats *LastMoveStats
}

func New(workers, iterationsPerThread int) *Client {
	return &Client{
		explorationParam: 1.414,
		workers:          workers,
		iterations:       iterationsPerThread,
		thinkTime:        time.Second,
	}
}

func (c *Client) UpdateThinkTime(t time.Duration) {
	c.thinkTime = t
}

func (c *Client) Stats() *LastMoveStats {
	return c.lastMoveStats
}

type threadResult struct {
	numIters  int
	thinkTime time.Duration
	visitMap  map[int]*node
}

func (c *Client) getNewRoot(b *tictactoe.Board) (newRoot *node) {
	newRoot = &node{
		client: c,
	}

	defer func() {
		newRoot.Parent = nil
		var untriedMoves []int
		legalMoves := b.LegalMoves()
	legalMoveLoop:
		for _, move := range legalMoves {
			for _, child := range newRoot.Children {
				if child.Move == move {
					continue legalMoveLoop
				}
			}

			untriedMoves = append(untriedMoves, move)
		}

		newRoot.UntriedMoves = untriedMoves
	}()

	if c.lastNode == nil {
		return newRoot
	}

	if c.lastNode.Move == b.LastMove {
		c.lastNode.Parent = nil
		return c.lastNode
	}

	for _, childNode := range c.lastNode.Children {
		if childNode.Move == b.LastMove {
			childNode.Parent = nil
			return childNode
		}
	}

	return newRoot
}

func (c *Client) GetNextMove(ctx context.Context, rootBoard *tictactoe.Board, player tictactoe.Player) (int, error) {
	c.lastMoveStats = nil

	rootBoard.Turn = (rootBoard.N * rootBoard.N) - len(rootBoard.LegalMoves())
	c.explorationParam = 0.9 + 0.6*math.Log(float64(rootBoard.N))

	root := c.getNewRoot(rootBoard)

	tacticalMoves, win := rootBoard.TacticalMoves(player)
	if len(tacticalMoves) == 1 {
		move := tacticalMoves[0]
		if win {
			return move, nil
		}

		for _, child := range root.Children {
			if child.Move == move {
				c.lastNode = child
				break
			}
		}

		c.lastMoveStats = &LastMoveStats{
			BestMove:         move,
			TacticalOverride: true,
		}

		return move, nil
	}

	results := make(chan threadResult, c.workers)

	var wg sync.WaitGroup
	wg.Add(c.workers)

	workerRoots := make([]*node, c.workers)
	for i := range workerRoots {
		workerRoots[i] = root.deepCopy()
	}

	g, gCtx := errgroup.WithContext(ctx)
	t := time.Now()
	for i := range c.workers {
		g.Go(func() error {
			defer wg.Done()

			root := workerRoots[i]
			thinkStart := time.Now()
			numIters, err := c.mctsIteration(gCtx, c.iterations, root, rootBoard.Clone(), player)
			if err != nil {
				return err
			}

			visitMap := make(map[int]*node)
			for _, c := range root.Children {
				visitMap[c.Move] = c
			}

			results <- threadResult{
				numIters:  numIters,
				visitMap:  visitMap,
				thinkTime: time.Since(thinkStart),
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	close(results)

	realThinkTime := time.Since(t)
	actualThinkTime := time.Duration(0)
	totalIters := 0

	nodes := map[int]*node{}
	totalVisits := make(map[int]int)
	for r := range results {
		totalIters += r.numIters
		actualThinkTime += r.thinkTime

		for move, node := range r.visitMap {
			totalVisits[move] += node.Visits
			n, ok := nodes[move]
			if !ok || node.Visits > n.Visits {
				nodes[move] = node
			}
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

	tacticalOverride := false
	if len(tacticalMoves) > 0 && !slices.Contains(tacticalMoves, bestMove) {
		tacticalOverride = true
		bestMove = tacticalMoves[rand.IntN(len(tacticalMoves))]
	}

	bestNode, ok := nodes[bestMove]
	if ok {
		c.lastNode = bestNode
	} else {
		bestNode = &node{}
	}

	c.lastMoveStats = &LastMoveStats{
		RealThinkTime:    realThinkTime,
		ActualThinkTime:  actualThinkTime,
		NumIterations:    totalIters,
		BestMove:         bestMove,
		MoveVisits:       bestNode.Visits,
		MoveWins:         bestNode.Wins,
		TacticalOverride: tacticalOverride,
	}

	return bestMove, nil
}

func (c *Client) mctsIteration(ctx context.Context, iterations int, root *node, board *tictactoe.Board, player tictactoe.Player) (int, error) {
	done := time.After(c.thinkTime)

	iterationsDone := 0
mctsIteration:
	for {
		select {
		case <-ctx.Done():
			break mctsIteration
		case <-done:
			break mctsIteration
		default:
		}

		if iterations != 0 && iterationsDone >= iterations {
			break
		}

		board := board.Clone()
		n := root
		current := player

		// Selection
		for !n.canExpand() && len(n.Children) > 0 {
			n = n.selectChild()
			err := board.ApplyMove(n.Move, current)
			if err != nil {
				return 0, fmt.Errorf("illegal move during selection: iteration %d, move %d, err %w", iterationsDone, n.Move, err)
			}

			current = -current
		}

		// Expansion
		if n.canExpand() {
			n = n.expand(board, current)
			current = -current
		}

		// Simulation
		winner := c.rollout(board, current)

		// Backprop
		n.backpropagate(winner)

		iterationsDone++
	}

	return iterationsDone, nil
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

		moves := board.LegalMoves()
		move := moves[rand.N(len(moves))]

		err := board.ApplyMove(move, current)
		if err != nil {
			panic("Illegal move during rollout")
		}

		current = -current
	}
}
