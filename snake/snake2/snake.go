// Simplified snake, easier for experimenting with autoplay.
//
// start with -a to autoplay
// hjkl to steer, r to reset, q to quit
//
// one speed only
// one level only
// board max 80x24, min 5x5
// initial length 5
// when moving, grows by 5 if not eating within 1.5x the manhattan distance between the head and the food at
//   the time the food is placed, gets no point, this repeats
// when eating, grows by 5 and gets 1 point

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"os"
	"os/user"
	"path"
	"slices"
	"time"
)

const (
	dOpen   = ' '
	dHoriz  = '-'
	dVert   = '|'
	dCorner = '+'
	dBody   = '#'
	dFood   = '*'
)

var (
	s        tcell.Screen
	defStyle tcell.Style
)

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
	up = body | (iota << dirShift)
	down
	left
	right
)

var (
	// These are indexed by high bits of up/down/left/right
	xDelta     = []int{0, 0, -1, 1}
	yDelta     = []int{-1, 1, 0, 0}
	oppositeOf = []uint8{down, up, right, left}
)

var (
	width, height int
	board         []uint8 // width * height
	xHead, yHead  int
	xTail, yTail  int
	xFood, yFood  int
	deadline      int
	savedDeadline int
	grow          int
	direction     uint8
	speed         int
	score         int
	dead          bool
	growAmount    int
)

type keyrec struct {
	key  rune
	next *keyrec
}

var keys *keyrec

type Score struct {
	Name  string `json:"name"`
	Date  string `json:"date"`
	Score int    `json:"score"`
}

var (
	scores    []Score
	scoreFile = path.Join(os.Getenv("HOME"), ".snake2")
)

type mover interface {
	autoMove()
	name() string
}

var automove mover

func clearState(w, h int) {
	width, height = w, h
	board = make([]uint8, width*height)
	xHead, yHead = width/2, height/2
	xTail, yTail = xHead, yHead
	xFood, yFood = 0, 0
	growAmount = 5
	grow = growAmount
	savedDeadline = width * height
	deadline = savedDeadline
	direction = right
	speed = 8
	score = 0
	dead = false
	keys = nil
	scores = nil
}

func main() {
	var autoplay bool
	flag.BoolVar(&autoplay, "a", false, "Autoplay \"local\" strategy")
	flag.Parse()
	initScreen()
	defer s.Fini()

	if autoplay {
		automove = newLocalMover()
	}

	resetGame()
	evChan := make(chan tcell.Event, 100)
	quitChan := make(chan struct{}, 1)
	go s.ChannelEvents(evChan, quitChan)
	ticker := time.NewTicker((1 * time.Second) / time.Duration(speed))
EvLoop:
	for {
		s.Show()
		select {
		case <-ticker.C:
			if !dead {
				tick()
			}
		case ev := <-evChan:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				resetGame()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyRune {
					switch ev.Rune() {
					case 'q':
						break EvLoop
					case 'r':
						s.Beep()
						resetGame()
					case 'h', 'j', 'k', 'l':
						if automove == nil {
							keys = &keyrec{ev.Rune(), keys}
						}
					}
				}
			}
		}
	}
	close(quitChan)
	ticker.Stop()
}

func tick() {
	if automove != nil {
		automove.autoMove()
	}
	if keys != nil {
		for ; keys != nil; keys = keys.next {
			next := direction
			switch keys.key {
			case 'h':
				next = left
			case 'j':
				next = down
			case 'k':
				next = up
			case 'l':
				next = right
			}
			if next != oppositeOf[direction>>dirShift] {
				direction = next
			}
			moveSnake()
		}
	} else {
		moveSnake()
	}
}

func moveSnake() {
	if direction != up && direction != down && direction != left && direction != right {
		panic(fmt.Sprintf("Bad direction %x", direction))
	}

	if dead {
		return
	}

	setAt(xHead, yHead, direction, dBody)
	xHead += xDelta[direction>>dirShift]
	yHead += yDelta[direction>>dirShift]
	nextKind := at(xHead, yHead) & kindMask
	setAt(xHead, yHead, body, dBody)

	if nextKind == wall || nextKind == body {
		s.Beep()
		dead = true
		recordResult()
		return
	}

	if nextKind == food {
		grow += growAmount
		score++
		showScore()
		placeFood()
	} else {
		deadline--
		if deadline == 0 {
			deadline = savedDeadline
			grow = growAmount
		} else if grow == 0 {
			t := at(xTail, yTail)
			setAt(xTail, yTail, open, dOpen)
			xTail += xDelta[t>>dirShift]
			yTail += yDelta[t>>dirShift]
		} else {
			grow--
		}
	}
}

func showScore() {
	auto := ""
	if automove != nil {
		auto = fmt.Sprintf("(%s) ", automove.name())
	}
	msg(fmt.Sprintf(" %sScore: %d ", auto, score))
}

func msg(m string) {
	for i, c := range m {
		s.SetContent(10+i, 0, c, nil, defStyle)
	}
}

func recordResult() {
	bytes, err := os.ReadFile(scoreFile)
	if err == nil {
		err = json.Unmarshal(bytes, &scores)
		if err != nil {
			scores = make([]Score, 0)
		}
	}
	name := ""
	if automove != nil {
		name = automove.name()
	} else {
		u, err := user.Current()
		if err == nil {
			name = u.Username
		}
	}
	scores = append(scores, Score{
		Name:  name,
		Date:  time.Now().Format("Jan 2 2006"),
		Score: score,
	})
	slices.SortFunc(scores, func(a, b Score) int {
		return b.Score - a.Score
	})
	if len(scores) > 10 {
		scores = scores[0:10]
	}
	bytes, err = json.Marshal(&scores)
	if err == nil {
		_ = os.WriteFile(scoreFile, bytes, 0666)
	}
}

func resetGame() {
	s.Clear()
	w, h := s.Size()
	if w > 80 {
		w = 80
	}
	if h > 24 {
		h = 24
	}
	if w < 5 || h < 5 {
		panic("Board too small")
	}
	clearState(w, h)

	for x := range width {
		setAt(x, 0, wall, dHoriz)
		setAt(x, height-1, wall, dHoriz)
	}
	for y := range height {
		setAt(0, y, wall, dVert)
		setAt(width-1, y, wall, dVert)
	}
	setAt(0, 0, wall, dCorner)
	setAt(width-1, 0, wall, dCorner)
	setAt(0, height-1, wall, dCorner)
	setAt(width-1, height-1, wall, dCorner)
	setAt(xHead, yHead, body, dBody)

	showScore()
	placeFood()
}

func placeFood() {
	for {
		xFood = rand.Intn(width)
		yFood = rand.Intn(height)
		if at(xFood, yFood) == open {
			setAt(xFood, yFood, food, dFood)
			savedDeadline = (3 * (abs(xHead-xFood) + abs(yHead-yFood))) / 2
			deadline = savedDeadline
			break
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func at(x, y int) uint8 {
	return board[y*width+x]
}

func setAt(x, y int, what uint8, how rune) {
	board[y*width+x] = what
	drawAt(x, y, how)
}

func initScreen() {
	var err error
	if s, err = tcell.NewScreen(); err != nil {
		panic(err)
	}
	if err = s.Init(); err != nil {
		panic(err)
	}
	defStyle = tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault)
	s.SetStyle(defStyle)
}

func drawAt(x, y int, c rune) {
	s.SetContent(x, y, c, nil, defStyle)
}
