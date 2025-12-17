package mcts

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type node struct {
	Parent   *node
	Children []*node

	Move   int
	Player tictactoe.Player

	Wins   float64
	Visits int

	UntriedMoves []int

	client *Client
}

func (n *node) canExpand() bool {
	maxChildren := int(2.0 * math.Sqrt(float64(n.Visits)))
	return len(n.UntriedMoves) > 0 && len(n.Children) < maxChildren
}

func (n *node) uctValue() float64 {
	if n.Visits == 0 {
		return math.Inf(1)
	}

	explorationParam := n.client.explorationParam
	parentVisits := n.Parent.Visits
	nWinRate := n.Wins / float64(n.Visits)
	logPVisit := math.Log(float64(parentVisits))

	return nWinRate + explorationParam*math.Sqrt(logPVisit/float64(n.Visits))
}

func (n *node) selectChild() *node {
	best := n.Children[0]
	bestVal := best.uctValue()

	for _, c := range n.Children[1:] {
		v := c.uctValue()
		if v > bestVal {
			best = c
			bestVal = v
		}
	}

	return best
}

func (n *node) expand(board *tictactoe.Board, player tictactoe.Player) *node {
	rand.Shuffle(len(n.UntriedMoves), func(i, j int) {
		n.UntriedMoves[i], n.UntriedMoves[j] = n.UntriedMoves[j], n.UntriedMoves[i]
	})

	var tacticalMoves []int
	var nearbyMoves []int

	move := -1
	for _, untriedMove := range n.UntriedMoves {
		if board.TacticalStone(untriedMove) {
			tacticalMoves = append(tacticalMoves, untriedMove)
			continue
		}

		if board.HasNeighbor(untriedMove, 2) {
			nearbyMoves = append(nearbyMoves, untriedMove)
		}
	}

	if len(tacticalMoves) > 0 {
		move = tacticalMoves[rand.N(len(tacticalMoves))]
	}

	eps := math.Max(0.05, 0.3*math.Exp(-0.1*float64(board.Turn)))
	if move == -1 && len(nearbyMoves) > 0 && rand.Float64() < eps {
		move = nearbyMoves[rand.N(len(nearbyMoves))]
	}

	if move == -1 {
		move = n.UntriedMoves[0]
	}

	n.UntriedMoves = slices.DeleteFunc(n.UntriedMoves, func(cmp int) bool {
		return cmp == move
	})

	board.ApplyMove(move, player)

	child := &node{
		Parent:       n,
		Move:         move,
		Player:       player,
		UntriedMoves: board.LegalMoves(),
		client:       n.client,
	}

	n.Children = append(n.Children, child)
	return child
}

const winValue = 1
const drawValue = 0.6

func (n *node) backpropagate(winner tictactoe.Player) {
	for n != nil {
		n.Visits++
		switch winner {
		case tictactoe.Empty:
			n.Wins += drawValue
		case n.Player:
			n.Wins += winValue
		}

		n = n.Parent
	}
}

func (n *node) deepCopy() *node {
	newNode := &node{
		Move:   n.Move,
		Player: n.Player,
		Wins:   n.Wins,
		client: n.client,
	}

	if len(n.UntriedMoves) > 0 {
		newNode.UntriedMoves = make([]int, len(n.UntriedMoves))
		copy(newNode.UntriedMoves, n.UntriedMoves)
	}

	if len(n.Children) > 0 {
		newNode.Children = make([]*node, len(n.Children))
		for i, child := range n.Children {
			newNode.Children[i] = child.deepCopy()
			newNode.Children[i].Parent = newNode
		}
	}

	return newNode
}
