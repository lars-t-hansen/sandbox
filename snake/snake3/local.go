package main

import (
	"math/rand"
)

type localMover struct {
	s *Snake
}

func newLocalMover(s *Snake) *localMover {
	return &localMover{s}
}

func (s *localMover) name() string {
	return "Local"
}

// Obvious first strategy is to "move towards the food but avoid bumping into things".

func (s *localMover) autoMove() {
	xDelta := s.s.xFood - s.s.xHead
	yDelta := s.s.yFood - s.s.yHead

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

	if secondary != 0 && secondary == s.s.direction {
		nextDirection, secondary = secondary, nextDirection
	}

	_ = s.tryMoves(nextDirection, secondary, rNormal) || s.tryMoves(nextDirection, secondary, rNone)
}

func (s *localMover) tryMoves(nextDirection, secondary uint8, rules int) bool {
	return s.tryMove(nextDirection, rules) ||
		(secondary != 0 && s.tryMove(secondary, rules)) ||
		s.tryMove(s.s.direction, rules) ||
		s.tryMoveRandom(rules)
}

const (
	rTunnel = 1
	rNormal = rTunnel
	rNone   = 0
)

func (s *localMover) tryMove(d uint8, rules int) bool {
	// Eliminate illegal and bad moves

	if d == oppositeOf[s.s.direction>>dirShift] {
		return false
	}

	xNext := s.s.xHead + xDelta[d>>dirShift]
	yNext := s.s.yHead + yDelta[d>>dirShift]

	if s.blockedAt(xNext, yNext) {
		return false
	}

	// "Don't go into an alley you can't get out of".  In a local interpretation this means, don't
	// make a move inbetween two obstacles.

	if (rules & rTunnel) != 0 {
		if d == up || d == down {
			if s.blockedAt(xNext-1, yNext) && s.blockedAt(xNext+1, yNext) {
				return false
			}
		} else {
			if s.blockedAt(xNext, yNext-1) && s.blockedAt(xNext, yNext+1) {
				return false
			}
		}
	}

	s.s.direction = d
	return true
}

func (s *localMover) blockedAt(x, y int) bool {
	nextKind := s.s.at(x, y) & kindMask
	return nextKind == wall || nextKind == body
}

func (s *localMover) tryMoveRandom(rules int) bool {
	k := rand.Intn(4)
	for i := range 4 {
		if s.tryMove(oppositeOf[(i+k)%4], rules) {
			return true
		}
	}
	return false
}
