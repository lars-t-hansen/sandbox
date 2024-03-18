// snake autoplay
//
// TODO:
//
// - simulate several (how many? max(height, width)? time limit?) steps ahead to make sure we don't
//   get into a situation we can't handle (crash, tunnel).  this is hard.  the canonical situation
//   we want to avoid is going into a cavity that we won't be able to escape.  the cavity can be
//   large, and escape is only possible if the tail comes around soon enough to remove a wall.
//   naive search of this space is exponential.  saving the state makes it tractable but blows up
//   space to something like O(square of size of cavity) + we may have to look *a lot* of steps
//   ahead.  Anyway, what are the decision points?  And what about random moves, do we remove those
//   (make those moves deterministic) or remember the plan and then move "according to plan" until
//   the plan is exhausted?
//
// - a micro-version of that is tractable but maybe not useful?
//
// - is perhaps "fast cavity detection" an interesting thing?  "If I turn left here, will I find
//   myself in a cavity"?  What's a cavity anyway?  Maybe the problem is really about "the number
//   of available squares to move to" and two concurrent flood fills will detect this?
//
// - Interesting: what if we don't know where the food is and must find it from the snake's head's
//   perspective?  Turns into a ray casting / vision problem.

package main

import (
	"math/rand"
)

// notifyTick is invoked when the head is about to move to a new square and can set the direction of
// that move.  It has access to the entire game state.

func notifyTick() {
	strategy1()
}

// Obvious first strategy is to "move towards the food but avoid bumping into things".

func strategy1() {
	xDelta := xFood - xHead
	yDelta := yFood - yHead

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

	if secondary != 0 && secondary == direction {
		nextDirection, secondary = secondary, nextDirection
	}

	_ = tryMoves(nextDirection, secondary, rNormal) || tryMoves(nextDirection, secondary, rNone)
}

func tryMoves(nextDirection, secondary uint8, rules int) bool {
	return tryMove(nextDirection, rules) ||
		(secondary != 0 && tryMove(secondary, rules)) ||
		tryMove(direction, rules) ||
		tryMoveRandom(rules)
}

const (
	rTunnel = 1
	rNormal = rTunnel
	rNone   = 0
)

func tryMove(d uint8, rules int) bool {
	// Eliminate illegal and bad moves

	if d == oppositeOf[direction>>dirShift] {
		return false
	}

	xNext := xHead + xDelta[d>>dirShift]
	yNext := yHead + yDelta[d>>dirShift]

	if blockedAt(xNext, yNext) {
		return false
	}

	// "Don't go into an alley you can't get out of".  In a local interpretation this means, don't
	// make a move inbetween two obstacles.

	if (rules & rTunnel) != 0 {
		if d == up || d == down {
			if blockedAt(xNext-1, yNext) && blockedAt(xNext+1, yNext) {
				return false
			}
		} else {
			if blockedAt(xNext, yNext-1) && blockedAt(xNext, yNext+1) {
				return false
			}
		}
	}

	direction = d
	return true
}

func blockedAt(x, y int) bool {
	nextKind := at(x, y) & kindMask
	return nextKind == wall || nextKind == body
}

func tryMoveRandom(rules int) bool {
	k := rand.Intn(4)
	for i := range 4 {
		if tryMove(oppositeOf[(i+k)%4], rules) {
			return true
		}
	}
	return false
}
