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
	for _, m := range lm.s.generateSingleMoves() {
		s2 := lm.s.clone(new(simUi))
		s2.direction = m.direction
		s2.move()
		p := len(s2.generateSingleMoves())
		moves = append(moves, l2move{move: m, possible: p})
	}

	// Move A to loc(A) is preferred over move B to loc(B) if, following the move,
	//  - distance(loc(A), food) == 0, or else
	//  - loc(A) does not have a forced move but loc(B) does, or else
	//  - distance(loc(A), food) < distance(loc(B), food).
	//  - On ties, prefer A over B if A is the same direction as the current direction.
	for i := 0 ; i < len(moves)-1 ; i++ {
		for j := i+1 ; j < len(moves) ; j++ {
			mi := moves[i]
			mj := moves[j]
			di := distance(mi.x, mi.y, lm.s.xFood, lm.s.yFood)
			dj := distance(mj.x, mj.y, lm.s.xFood, lm.s.yFood)
			swap := false
			if dj == 0 {
				swap = true
			} else if mj.possible > 1 && mi.possible <= 1 {
				swap = true
			} else if dj < di {
				swap = true
			} else if dj == di && mj.direction == lm.s.direction {
				swap = true
			} /*else if (mj.possible > 1) == (mi.possible > 1) && mj.direction == lm.s.direction {
				swap = true
			}*/
			if swap {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}

	// And then pick the first one if there is one
	if len(moves) > 0 {
		lm.s.direction = moves[0].direction
	}
}
