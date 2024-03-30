// See main.go for instructions.

package main

import (
	"fmt"
	"math/rand"
	"slices"
)

// The snake's view of the UI: visual representation (a little leaky) and simple callback interface.

const (
	dOpen   = ' '
	dHoriz  = '-'
	dVert   = '|'
	dCorner = '+'
	dBody   = '#'
	dFood   = '*'
)

type Ui interface {
	// Reset and clear the screen and return its dimensions.
	clear() (width, height int)

	// Draw the rune at the location.
	drawAt(x, y int, c rune)

	// The snake has died.
	notifyDead()

	// The snake has eaten and its score has been updated.
	notifyNewScore()
}

// Game representation

const (
	open = iota // `open` must be zero
	food
	wall
	body
)

const (
	kindMask = 3
	dirShift = 2
)

const (
	// This order is encoded in the arrays below but is also used by other code, do not change it.
	// The encoding is also part of the public API.
	up = body | (iota << dirShift)
	down
	left
	right
)

var (
	// These "constants" are indexed by high bits of up/down/left/right.
	xDelta     = []int{0, 0, -1, 1}
	yDelta     = []int{-1, 1, 0, 0}
	oppositeOf = []uint8{down, up, right, left}
)

// The UI logic creates a new snake with this newSnake(), then when it's ready to use it calls
// reset() on it.  Board locations can be read and written with at() and setAt().  Other variables
// should be considered read-only.
//
// To move the snake, set `direction` and call move().

type Snake struct {
	width, height int     // dimensions
	board         []uint8 // width * height: open, food, wall, up, down, left, right, body
	xHead, yHead  int     // where the head's at
	xTail, yTail  int     // where the tail's at
	xFood, yFood  int     // where the food's at
	deadline      int     // how long before we grow without eating
	savedDeadline int     // the initializer for `deadline`
	grow          int     // if non-zero, we don't move the tail
	direction     uint8   // direction to move in: up, down, left, right
	speed         int     // this many moves per second
	score         int     // current score, updated before calling ui.notifyNewScore
	dead          bool    // set to true once dead, before calling ui.notifyDead
	growAmount    int     // how many segments to grow by when growing
	ui            Ui      // ui callbacks
}

func newSnake(ui Ui) *Snake {
	var s Snake
	s.ui = ui
	return &s
}

func (s *Snake) clone(ui Ui) *Snake {
	var t Snake
	t = *s
	t.ui = ui
	t.board = slices.Clone(s.board)
	return &t
}

func (s *Snake) reset() {
	w, h := s.ui.clear()
	if w > 80 {
		w = 80
	}
	if h > 24 {
		h = 24
	}
	if w < 5 || h < 5 {
		panic("Board too small")
	}

	s.width, s.height = w, h
	s.board = make([]uint8, s.width*s.height)
	s.xHead, s.yHead = s.width/2, s.height/2
	s.xTail, s.yTail = s.xHead, s.yHead
	s.xFood, s.yFood = 0, 0
	s.growAmount = 5
	s.grow = s.growAmount
	s.savedDeadline = s.width * s.height
	s.deadline = s.savedDeadline
	s.direction = right
	s.speed = 8
	s.score = 0
	s.dead = false

	for x := 0 ; x < s.width ; x++ {
		s.setAt(x, 0, wall, dHoriz)
		s.setAt(x, s.height-1, wall, dHoriz)
	}
	for y := 0 ; y < s.height ; y++ {
		s.setAt(0, y, wall, dVert)
		s.setAt(s.width-1, y, wall, dVert)
	}
	s.setAt(0, 0, wall, dCorner)
	s.setAt(s.width-1, 0, wall, dCorner)
	s.setAt(0, s.height-1, wall, dCorner)
	s.setAt(s.width-1, s.height-1, wall, dCorner)
	s.setAt(s.xHead, s.yHead, body, dBody)

	s.ui.notifyNewScore()

	s.placeFood()
}

func (s *Snake) move() {
	if s.direction != up && s.direction != down && s.direction != left && s.direction != right {
		panic(fmt.Sprintf("Bad direction %x", s.direction))
	}

	if s.dead {
		return
	}

	s.setAt(s.xHead, s.yHead, s.direction, dBody)
	s.xHead += xDelta[s.direction>>dirShift]
	s.yHead += yDelta[s.direction>>dirShift]
	nextKind := s.at(s.xHead, s.yHead) & kindMask
	s.setAt(s.xHead, s.yHead, body, dBody)

	if nextKind == wall || nextKind == body {
		s.dead = true
		s.ui.notifyDead()
		return
	}

	if nextKind == food {
		s.grow += s.growAmount
		s.score++
		s.ui.notifyNewScore()
		s.placeFood()
	} else {
		s.deadline--
		if s.deadline == 0 {
			s.deadline = s.savedDeadline
			s.grow = s.growAmount
		} else if s.grow == 0 {
			t := s.at(s.xTail, s.yTail)
			s.setAt(s.xTail, s.yTail, open, dOpen)
			s.xTail += xDelta[t>>dirShift]
			s.yTail += yDelta[t>>dirShift]
		} else {
			s.grow--
		}
	}
}

func (s *Snake) at(x, y int) uint8 {
	return s.board[y*s.width+x]
}

func (s *Snake) setAt(x, y int, what uint8, how rune) {
	s.board[y*s.width+x] = what
	s.ui.drawAt(x, y, how)
}

// private
func (s *Snake) placeFood() {
	for {
		s.xFood = rand.Intn(s.width)
		s.yFood = rand.Intn(s.height)
		if s.at(s.xFood, s.yFood) == open {
			s.setAt(s.xFood, s.yFood, food, dFood)
			s.savedDeadline = (3 * distance(s.xHead, s.yHead, s.xFood, s.yFood)) / 2
			s.deadline = s.savedDeadline
			break
		}
	}
}

// Moves (zero or more) followed by the final position.

type move struct {
	direction []uint8
	x, y int
}

// Generate move sequences of up to n moves from the current position.  The snake is not updated.
// If a sequence takes us to food it is first in the list.  A sequence that ends in certain death is
// not included, nor its prefix.

func (s *Snake) generateMoves(n int) []move {
	if n != 1 {
		panic("N=1")
	}
	moves := make([]move, 0)
	// The guard is required because, if dead then the head could be on a segment that has open
	// space next to it.
	if !s.dead {
		for i := 0 ; i < 4 ; i++ {
			nx := s.xHead + xDelta[i]
			ny := s.yHead + yDelta[i]
			d := body | (uint8(i) << dirShift)
			m := move{[]uint8{d}, nx, ny}
			switch s.at(nx, ny) {
			case open:
				moves = append(moves, m)
			case food:
				moves = append([]move{m}, moves...)
			}
		}
	}
	return moves
}
