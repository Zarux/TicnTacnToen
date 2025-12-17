package zobrist

import "math/rand/v2"

func New(n int) [][]uint64 {
	zobrist := make([][]uint64, n*n)
	for i := range n * n {
		zobrist[i] = make([]uint64, 2) // index by player
		zobrist[i][0] = rand.Uint64()
		zobrist[i][1] = rand.Uint64()
	}

	return zobrist
}
