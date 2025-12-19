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
	var centerMoves []int
	var invalidMoves []int

	center := float64(board.N) / 2.0
	radius := math.Max(1, float64(board.N)/3)

	move := -1
	for _, untriedMove := range n.UntriedMoves {
		if board.Cells[untriedMove] != tictactoe.Empty {
			invalidMoves = append(invalidMoves, untriedMove)
			continue
		}

		if board.TacticalStone(untriedMove) {
			tacticalMoves = append(tacticalMoves, untriedMove)
			continue
		}

		if board.HasNeighbor(untriedMove, 2) {
			nearbyMoves = append(nearbyMoves, untriedMove)
		}

		x := float64(untriedMove % board.N)
		y := float64(untriedMove / board.N)

		if math.Abs(x+1-center)+math.Abs(y+1-center) <= radius {
			centerMoves = append(centerMoves, untriedMove)
		}
	}

	if len(tacticalMoves) > 0 {
		move = tacticalMoves[rand.N(len(tacticalMoves))]
	}

	epsNearby := math.Max(0.05, 0.3*math.Exp(-0.08*float64(board.Turn)))
	if move == -1 && len(nearbyMoves) > 0 && rand.Float64() < epsNearby {
		move = nearbyMoves[rand.N(len(nearbyMoves))]
	}

	epsCenter := math.Max(0.05, 0.2*math.Exp(-0.05*float64(board.Turn)))
	if move == -1 && len(centerMoves) > 0 && rand.Float64() < epsCenter {
		move = centerMoves[rand.N(len(centerMoves))]
	}

	if move == -1 {
		move = n.UntriedMoves[0]
	}

	n.UntriedMoves = slices.DeleteFunc(n.UntriedMoves, func(cmp int) bool {
		return cmp == move || slices.Contains(invalidMoves, cmp)
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

func (n *node) deepCopy(validMoves []int) *node {
	newNode := &node{
		Move:   n.Move,
		Player: n.Player,
		Wins:   n.Wins,
		client: n.client,
	}

	if len(n.UntriedMoves) > 0 {
		newNode.UntriedMoves = make([]int, len(n.UntriedMoves))
		copy(newNode.UntriedMoves, n.UntriedMoves)

		newNode.UntriedMoves = slices.DeleteFunc(newNode.UntriedMoves, func(cmp int) bool {
			return !slices.Contains(validMoves, cmp)
		})
	}

	if len(n.Children) > 0 {
		newNode.Children = make([]*node, len(n.Children))
		for i, child := range n.Children {
			newNode.Children[i] = child.deepCopy(validMoves)
			newNode.Children[i].Parent = newNode
		}
	}

	return newNode
}
