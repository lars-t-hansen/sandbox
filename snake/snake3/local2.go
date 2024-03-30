// This is a more rigorous reimplementation of the "local" strategy

package main

type local2Mover struct /* implements mover */ {
	s *Snake
}

func newLocal2Mover(s *Snake) *local2Mover {
	return &local2Mover{s}
}

func (_ *local2Mover) name() string {
	return "Local2"
}

type l2move struct {
	move
	possible int
}

func (lm *local2Mover) autoMove() {
	// Generate all legal single moves (which are directions from the head).
	moves := make([]l2move, 0)
	for _, m := range lm.s.generateMoves(1) {
		s2 := lm.s.clone(new(simUi))
		s2.direction = m.direction[0]
		s2.move()
		p := len(s2.generateMoves(1))
		moves = append(moves, l2move{move: m, possible: p})
	}

	// Move A to loc(A) is preferred over move B to loc(B) if, following the move,
	//  - distance(loc(A), food) == 0, or else
	//  - loc(A) does not have a forced move but loc(B) does, or else
	//  - loc(A) has a move but loc(B) does not, or else
	//  - distance(loc(A), food) < distance(loc(B), food), or else
	//  - distance(loc(A), food) == distance(loc(B), food) and A is the same direction as
	//    the current direction.
	for i := 0 ; i < len(moves)-1 ; i++ {
		for j := i+1 ; j < len(moves) ; j++ {
			mi := moves[i]
			mj := moves[j]
			di := distance(mi.x, mi.y, lm.s.xFood, lm.s.yFood)
			dj := distance(mj.x, mj.y, lm.s.xFood, lm.s.yFood)
			swap := false
			if di == 0 || dj == 0 {
				swap = dj == 0
			} else if (mi.possible > 1) != (mj.possible > 1) {
				swap = mj.possible > 1
			} else if (mi.possible > 0) != (mj.possible > 0) {
				swap = mj.possible > 0
			} else if di == dj {
				if mj.direction[0] == lm.s.direction {
					swap = true
				}
			} else if dj < di {
				swap = true
			}
			if swap {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}

	// And then pick the first one if there is one
	if len(moves) > 0 {
		lm.s.direction = moves[0].direction[0]
	}
}
