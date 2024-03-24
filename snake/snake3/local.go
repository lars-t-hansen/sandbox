package main

type localMover struct /* implements mover */ {
	s *Snake
	clever bool
}

// If "clever" is true then some heuristics are used to avoid possibly bad situations, these are not
// desirable if the local mover is used for search, however.

func newLocalMover(s *Snake, clever bool) *localMover {
	return &localMover{s, clever}
}

func (_ *localMover) name() string {
	return "Local"
}

// Obvious first strategy is to "move towards the food but avoid bumping into things".

func (lm *localMover) autoMove() {
	xDelta := lm.s.xFood - lm.s.xHead
	yDelta := lm.s.yFood - lm.s.yHead

	// xDelta and yDelta will not both be zero because once we move onto the food the food moves
	// 1. reduce delta-y to food if possible
	// 2. otherwise reduce delta-x

	// These are absolute directions not relative
	var nextDirection uint8
	var secondary uint8
	if yDelta > 0 {
		nextDirection = down
		if xDelta < 0 {
			secondary = left
		} else if xDelta > 0 {
			secondary = right
		}
	} else if yDelta < 0 {
		nextDirection = up
		if xDelta < 0 {
			secondary = left
		} else if xDelta > 0 {
			secondary = right
		}
	} else if xDelta < 0 {
		nextDirection = left
		if yDelta < 0 {
			secondary = up
		} else if yDelta > 0 {
			secondary = down
		}
	} else if xDelta > 0 {
		nextDirection = right
		if yDelta < 0 {
			secondary = up
		} else if yDelta > 0 {
			secondary = down
		}
	} else {
		panic("Unexpected")
	}

	// Prefer to move in the direction we're moving in if that is sensible

	if secondary != 0 && secondary == lm.s.direction {
		nextDirection, secondary = secondary, nextDirection
	}

	if lm.clever {
		_ = lm.tryMoves(nextDirection, secondary, rNormal) || lm.tryMoves(nextDirection, secondary, rNone)
	} else {
		_ = lm.tryMoves(nextDirection, secondary, rNone)
	}
}

func (lm *localMover) tryMoves(nextDirection, secondary uint8, rules int) bool {
	return lm.tryMove(nextDirection, rules) ||
		(secondary != 0 && lm.tryMove(secondary, rules)) ||
		lm.tryMove(lm.s.direction, rules) ||
		lm.tryMoveAnywhere(rules)
}

const (
	rTunnel = 1
	rNormal = rTunnel
	rNone   = 0
)

func (lm *localMover) tryMove(d uint8, rules int) bool {
	// Eliminate illegal and bad moves

	if d == oppositeOf[lm.s.direction>>dirShift] {
		return false
	}

	xNext := lm.s.xHead + xDelta[d>>dirShift]
	yNext := lm.s.yHead + yDelta[d>>dirShift]

	if lm.blockedAt(xNext, yNext) {
		return false
	}

	// "Don't go into an alley you can't get out of".  In a local interpretation this means, don't
	// make a move inbetween two obstacles.

	if (rules & rTunnel) != 0 {
		if d == up || d == down {
			if lm.blockedAt(xNext-1, yNext) && lm.blockedAt(xNext+1, yNext) {
				return false
			}
		} else {
			if lm.blockedAt(xNext, yNext-1) && lm.blockedAt(xNext, yNext+1) {
				return false
			}
		}
	}

	lm.s.direction = d
	return true
}

func (lm *localMover) blockedAt(x, y int) bool {
	nextKind := lm.s.at(x, y) & kindMask
	return nextKind == wall || nextKind == body
}

// This uses the same schedule every time in order to be deterministic.  (snake1 and snake2 start at
// a random spot in 0..3.)

func (lm *localMover) tryMoveAnywhere(rules int) bool {
	for i := 0 ; i < 4 ; i++ {
		if lm.tryMove(oppositeOf[i], rules) {
			return true
		}
	}
	return false
}
