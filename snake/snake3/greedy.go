// The greedy snake always moves directly toward the food, consequences be damned.
// (Maybe "Hungry" would be better than "Greedy")

package main

type greedyMover struct /* implements mover */ {
	s *Snake
}

func newGreedyMover(s *Snake) *greedyMover {
	return &greedyMover{s}
}

func (_ *greedyMover) name() string {
	return "Greedy"
}

func (gm *greedyMover) autoMove() {
	// Generate all legal single moves (which are directions from the head).
	moves := gm.s.generateSingleMoves()

	// Move A is preferred over move B if, following the move, the distance from location loc(A) to
	// food is shorter than from loc(B) to food.  On ties, prefer A over B if A is the same
	// direction as the current direction.
	for i := 0 ; i < len(moves)-1 ; i++ {
		for j := i+1 ; j < len(moves) ; j++ {
			mi := moves[i]
			mj := moves[j]
			di := distance(mi.x, mi.y, gm.s.xFood, gm.s.yFood)
			dj := distance(mj.x, mj.y, gm.s.xFood, gm.s.yFood)
			swap := false
			if dj < di {
				swap = true
			} else if dj == di && mj.direction == gm.s.direction {
				swap = true
			}
			if swap {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}

	// And then pick the first one if there is one
	if len(moves) > 0 {
		gm.s.direction = moves[0].direction
	}
}
