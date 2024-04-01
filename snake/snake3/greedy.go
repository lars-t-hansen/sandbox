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
	moves := gm.s.generateMoves(1)

	// If there are any moves, make the best one
	if len(moves) > 0 {
		// Move A is preferred over move B if, following the move, the distance from location loc(A)
		// to food is shorter than from loc(B) to food.  On ties, prefer A over B if A is the same
		// direction as the current direction.
		best := moves[0]
		db := distance(best.x, best.y, gm.s.xFood, gm.s.yFood)
		for _, m := range moves[1:] {
			dm := distance(m.x, m.y, gm.s.xFood, gm.s.yFood)
			improve := false
			if dm < db {
				improve = true
			} else if dm == db && m.direction[0] == gm.s.direction {
				improve = true
			}
			if improve {
				best = m
				db = dm
			}
		}
		gm.s.direction = best.direction[0]
	}
}
