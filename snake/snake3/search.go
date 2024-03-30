// A mover that searches a little.
//
// Generate all possible move sequences of length up to N (N very short, say, <= 3) from the current
// position, then starting in the final position of each sequence play a simple and deterministic
// strategy for K moves (K larger, say, 100) or until we're stuck or food is found.  Moves are
// sorted by desirabilty after the outcome of the simulation (food is great, close to food is better
// than far away from food, few moves is better than many moves, dying is bad) and the best is
// picked.
//
// This does not probe further after finding food, so a morsel at the bottom of a tunnel with no
// exit will trick it.
//
// This does not save a plan, since the moves made after the initial ones are deterministic.
//
// Currently only N=1 is supported.

package main

import "fmt"

const (
	initialMoves = 1			// N
	depth = 200					// K
)

type searchMover struct /* implements mover */ {
	s *Snake
}

func newSearchMover(s *Snake) mover {
	if initialMoves != 1 {
		panic("N != 1 not supported")
	}
	return &searchMover{s}
}

func (_ *searchMover) name() string {
	return fmt.Sprintf("Search(%d,%d)", initialMoves, depth)
}

type simUi struct {
	dead, fed bool
}

func (s *simUi) clear() (width, height int) {
	panic("clear() should not be called")
}

func (s *simUi) drawAt(x, y int, c rune) {}

func (s *simUi) notifyDead() {
	s.dead = true
}

func (s *simUi) notifyNewScore() {
	s.fed = true
}

func (sm *searchMover) autoMove() {
	type probe struct {
		direction uint8
		moves     int
		ui        *simUi
		s         *Snake
		mover     mover
	}

	// Generate some legal initial positions.

	probers := make([]*probe, 0)
	if initialMoves != 1 {
		panic("initialMoves")
	}
	for _, m := range sm.s.generateMoves(1) {
		ui := new(simUi)
		s2 := sm.s.clone(ui)
		s2.direction = m.direction[0]
		s2.move()
		probers = append(probers, &probe{m.direction[0], 1, ui, s2, newGreedyMover(s2) /*newLocalMover(s2, false)*/})
	}

	// Evaluate those positions.
	//
	// From each legal initial position, automove until dead or fed or exhausted.  Then prioritize:
	// - moves that find food
	// - otherwise, moves that bring us closer to food
	// - and on ties, shorter move sequences over longer
	//
	// For hopeless positions (all initial moves lead to death, following the normal strategy) we
	// choose the longest move sequence in the hope that some better plan will be found along the
	// way.  Not sure yet if that matters.
	//
	// It's possible that minimizing # of moves is more important than getting close to food, but
	// they are probably closely related anyway.

	var best *probe
	var bestDead *probe
	var bestAte bool
	var xFood = sm.s.xFood
	var yFood = sm.s.yFood
	for _, p := range probers {
		for remaining := depth; remaining > 0 && !p.ui.dead && !p.ui.fed; remaining-- {
			p.mover.autoMove()
			p.s.move()
			p.moves++
		}

		if p.ui.fed {
			if !bestAte {
				best = p
			} else {
				// Pick the one that gets us closer, and optimize for # of moves on ties
				bestDist := distance(best.s.xHead, best.s.yHead, xFood, yFood)
				dist := distance(p.s.xHead, p.s.yHead, xFood, yFood)
				if dist < bestDist {
					best = p
				} else if dist == bestDist && p.moves < best.moves {
					best = p
				}
			}
			bestAte = true
		}

		if !bestAte {
			if p.ui.dead {
				if bestDead == nil {
					bestDead = p
				} else if p.moves > bestDead.moves {
					bestDead = p
				}
			} else if best == nil {
				best = p
			} else {
				// Pick the one that gets us closer, and optimize for # of moves on ties
				bestDist := distance(best.s.xHead, best.s.yHead, xFood, yFood)
				dist := distance(p.s.xHead, p.s.yHead, xFood, yFood)
				if dist < bestDist {
					best = p
				} else if dist == bestDist && p.moves < best.moves {
					best = p
				}
			}
		}
	}

	// If no sensible move then pick the most honorable death.
	if best == nil {
		best = bestDead
	}

	// Make the move.  If we're stuck at the bottom of a cul-de-sac there will be no initial moves
	// and hence no final moves to make, just keep going (and die).
	if best != nil {
		sm.s.direction = best.direction
	}
}
