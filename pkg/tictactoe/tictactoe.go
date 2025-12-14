package tictactoe

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/Zarux/ticntacntoen/pkg/zobrist"
)

var errIllegalMove = errors.New("illegal move")

type Player int8

const (
	Empty Player = 0
	P1    Player = 1
	P2    Player = -1
)

func (p Player) Mark() string {
	s := " "
	if p == P1 {
		s = "X"
	}

	if p == P2 {
		s = "O"
	}

	return s
}

func (p Player) Idx() int {
	if p == P1 {
		return 0
	}

	if p == P2 {
		return 1
	}

	return -1
}

type Move struct {
	X int
	Y int
}

type Board struct {
	N        int
	K        int
	Cells    []Player
	LastMove int
	Hash     uint64

	xList      []int
	yList      []int
	emptyCells []int

	game *Game
}

func (b *Board) GetIdx(x, y int) int {
	return y*b.N + x
}

func (b *Board) GetMove(idx int) Move {
	return Move{
		X: idx % b.N,
		Y: idx / b.N,
	}
}

func (b *Board) Get(x, y int) Player {
	idx := b.GetIdx(x, y)
	return b.Cells[idx]
}

func (b *Board) Play(p Player, m Move) {
	idx := b.GetIdx(m.X, m.Y)
	b.ApplyMove(idx, p)
}

func (b *Board) ApplyMove(idx int, p Player) error {
	if b.Cells[idx] != Empty {
		b.Print()
		fmt.Println(idx)
		fmt.Println(b.emptyCells)
		return errIllegalMove
	}

	b.Cells[idx] = p

	b.LastMove = idx
	b.Hash ^= b.game.ZobristKeys[idx][p.Idx()]

	b.emptyCells = slices.DeleteFunc(b.emptyCells, func(cmp int) bool {
		return cmp == idx
	})

	return nil
}

func (b *Board) UndoMove(idx int) {
	if b.Cells[idx] == Empty {
		panic("UndoMove on empty cell")
	}

	p := b.Cells[idx]
	b.Cells[idx] = Empty
	b.emptyCells = append(b.emptyCells, idx)
	b.Hash ^= b.game.ZobristKeys[idx][p.Idx()]
}

func (b *Board) AnyLegalMoves() bool {
	return slices.Contains(b.Cells, Empty)
}

func (b *Board) LegalMoves() []int {
	emptyCells := make([]int, 0, len(b.Cells))
	for m, p := range b.Cells {
		if p != Empty {
			continue
		}
		emptyCells = append(emptyCells, m)
	}

	return emptyCells
}

func (b *Board) CheckWinner() Player {
	winner := b.checkFrom(b.LastMove)
	return winner
}

func (b *Board) Clone() *Board {
	cells := make([]Player, len(b.Cells))
	copy(cells, b.Cells)

	emptyCells := make([]int, len(b.emptyCells))
	copy(emptyCells, b.emptyCells)

	return &Board{
		N:          b.N,
		K:          b.K,
		Cells:      cells,
		xList:      b.xList,
		yList:      b.yList,
		emptyCells: emptyCells,
		Hash:       b.Hash,
		LastMove:   b.LastMove,
		game:       b.game,
	}
}

type Dir struct {
	dx, dy int
}

var directions = []Dir{
	{1, 0},  // horizontal
	{0, 1},  // vertical
	{1, 1},  // diag
	{1, -1}, // anti-diag
}

func (b *Board) checkFrom(idx int) Player {
	p := b.Cells[idx]
	if p == Empty {
		return Empty
	}

	N := b.N
	x := b.xList[idx]
	y := b.yList[idx]

	for _, d := range directions {
		count := 1

		for _, dir := range []int{-1, 1} {
			nx := x + d.dx*dir
			ny := y + d.dy*dir

			for nx >= 0 && ny >= 0 && nx < N && ny < N {
				nidx := ny*N + nx
				if b.Cells[nidx] != p {
					break
				}

				count++
				if count >= b.K {
					return p
				}

				nx += d.dx * dir
				ny += d.dy * dir
			}
		}
	}
	return Empty
}

func (b *Board) DoesPlacingAnyStoneHereEndTheGame(idx int) bool {
	x := idx % b.N
	y := idx / b.N

	t := false
	for _, d := range directions {
		if b.checkOneColorFromSide(x, y, d.dx, d.dy) || b.checkOneColorFromSide(x, y, -d.dx, -d.dy) {
			t = true
			break
		}
	}

	return t
}

func (b *Board) checkOneColorFromSide(x, y, dx, dy int) bool {
	nx := x + dx
	ny := y + dy

	if nx < 0 || ny < 0 || nx >= b.N || ny >= b.N {
		return false
	}

	color := b.Cells[ny*b.N+nx]
	if color == Empty {
		return false
	}

	count := 1 // stone we "place" at center

	// forward
	fx, fy := nx, ny
	for {
		count++
		fx += dx
		fy += dy
		if fx < 0 || fy < 0 || fx >= b.N || fy >= b.N {
			break
		}
		if b.Cells[fy*b.N+fx] != color {
			break
		}
	}

	// backward
	bx := x - dx
	by := y - dy
	for {
		if bx < 0 || by < 0 || bx >= b.N || by >= b.N {
			break
		}
		if b.Cells[by*b.N+bx] != color {
			break
		}
		count++
		bx -= dx
		by -= dy
	}

	return count >= b.K
}

func (b *Board) hasNeighbor(idx int) (hasNeighbor bool) {
	x := idx % b.N
	y := idx / b.N

	for _, dx := range []int{-1, 1} {
		for _, dy := range []int{-1, 1} {
			nx := x + dx
			ny := y + dy

			if nx < 0 || ny < 0 || nx >= b.N || ny >= b.N {
				continue
			}

			if b.Cells[b.GetIdx(nx, ny)] != Empty {
				return true
			}
		}
	}

	return false
}

func (b *Board) ForcedMove(player Player) (int, bool) {
	blockingMoves := []int{}

	for i, c := range b.Cells {
		if c != Empty {
			continue
		}

		b.ApplyMove(i, player)
		isWin := b.CheckWinner() == player
		b.UndoMove(i)
		if isWin {
			return i, true
		}

		b.ApplyMove(i, -player)
		isBlock := b.CheckWinner() != Empty
		b.UndoMove(i)
		if isBlock {
			blockingMoves = append(blockingMoves, i)
		}
	}

	if len(blockingMoves) > 0 {
		return blockingMoves[rand.N(len(blockingMoves))], true
	}

	return -1, false
}

var timeSpentNeighbour time.Duration
var timeSpent2xWinner time.Duration

var cachehit int
var cachemiss int

func (b *Board) BiasedRandomMoveCache(cache *sync.Map) int {
	if v, ok := cache.Load(b.Hash); ok {
		cachehit++
		moves := v.([]int)
		return moves[rand.N(len(moves))]
	}
	cachemiss++

	moves := b.biasedRandomMoves()
	cache.Store(b.Hash, moves)
	return moves[rand.N(len(moves))]

}

func (b *Board) BiasedRandomMove() int {
	moves := b.biasedRandomMoves()
	return moves[rand.N(len(moves))]
}

func (b *Board) biasedRandomMoves() []int {
	var near []int
	var tactical []int
	var empty []int

	for m, p := range b.Cells {
		if p != Empty {
			continue
		}

		empty = append(empty, m)

		t := time.Now()
		if b.hasNeighbor(m) {
			near = append(near, m)
		}
		timeSpentNeighbour += time.Since(t)

		t = time.Now()
		if b.DoesPlacingAnyStoneHereEndTheGame(m) {
			tactical = append(tactical, m)
		}
		timeSpent2xWinner += time.Since(t)

	}

	if len(tactical) > 0 {
		return tactical
	}

	if len(near) > 0 {
		return near
	}

	return empty
}

func (b *Board) Print() {
	fmt.Printf("%#v\n", b.Cells)
	for i, p := range b.Cells {
		fmt.Printf("[%s]", p.Mark())
		if (i+1)%b.N == 0 {
			fmt.Print("\n")
		}
	}

	fmt.Println(timeSpentNeighbour)
	fmt.Println(timeSpent2xWinner)
	fmt.Println(timeSpentNeighbour+timeSpent2xWinner, cachehit, cachemiss)
	timeSpent2xWinner = 0
	timeSpentNeighbour = 0

	fmt.Println("--------")
}

type Game struct {
	Board       *Board
	ZobristKeys [][]uint64
}

func New(N, K int) (*Game, error) {
	g := Game{
		ZobristKeys: zobrist.New(N),
	}

	board := g.newBoard(N, K)
	g.Board = &board

	return &g, nil
}

func LoadGame(data []byte) (*Game, error) {
	g := &Game{}
	err := json.Unmarshal(data, g)
	if err != nil {
		return nil, err
	}

	g.Board.game = g
	N := g.Board.N

	xList := make([]int, N*N)
	for i := range N * N {
		xList[i] = i % N
	}

	yList := make([]int, N*N)
	for i := range N * N {
		yList[i] = i / N
	}

	emptyCells := make([]int, 0, N*N)
	for i, v := range g.Board.Cells {
		if v != Empty {
			continue
		}

		emptyCells = append(emptyCells, i)
	}

	g.Board.xList = xList
	g.Board.yList = yList
	g.Board.emptyCells = emptyCells

	return g, nil
}

func (g *Game) Save() ([]byte, error) {
	j, err := json.MarshalIndent(g, "", "\t")
	if err != nil {
		panic(err)
	}

	return j, nil
}

func (g *Game) newBoard(N, K int) Board {
	xList := make([]int, N*N)
	for i := range N * N {
		xList[i] = i % N
	}

	yList := make([]int, N*N)
	for i := range N * N {
		yList[i] = i / N
	}

	emptyCells := make([]int, N*N)
	for i := range emptyCells {
		emptyCells[i] = i
	}

	return Board{
		N:          N,
		Cells:      make([]Player, N*N),
		K:          K,
		xList:      xList,
		yList:      yList,
		emptyCells: emptyCells,
		game:       g,
	}
}
