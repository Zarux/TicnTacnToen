package mcts

import (
	"math"

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
	move := n.UntriedMoves[len(n.UntriedMoves)-1]
	n.UntriedMoves = n.UntriedMoves[:len(n.UntriedMoves)-1]

	board.Cells[move] = player

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
