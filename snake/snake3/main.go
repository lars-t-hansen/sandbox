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
	"os"
	"os/user"
	"path"
	"slices"
	"time"
)

type keyrec struct {
	key  rune
	next *keyrec
}

type Score struct {
	Name  string `json:"name"`
	Date  string `json:"date"`
	Score int    `json:"score"`
}

type mover interface {
	autoMove()
	name() string
}

var (
	s         tcell.Screen
	defStyle  tcell.Style
	snake     *Snake
	keys      *keyrec
	scores    []Score
	scoreFile string
	automove  mover
)

type RealUi int

func (_ RealUi) clear() (width, height int) {
	s.Clear()
	return s.Size()
}

func (_ RealUi) drawAt(x, y int, c rune) {
	s.SetContent(x, y, c, nil, defStyle)
}

func (_ RealUi) notifyDead() {
	s.Beep()
	recordResult()
}

func (_ RealUi) notifyNewScore() {
	auto := ""
	if automove != nil {
		auto = fmt.Sprintf("(%s) ", automove.name())
	}
	msg(fmt.Sprintf(" %sScore: %d ", auto, snake.score))
}

func initScreen() {
	scoreFile = path.Join(os.Getenv("HOME"), ".snake3")
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
		Score: snake.score,
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

func tick() {
	if automove != nil {
		automove.autoMove()
	}
	if keys != nil {
		for ; keys != nil; keys = keys.next {
			next := snake.direction
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
			if next != oppositeOf[snake.direction>>dirShift] {
				snake.direction = next
			}
			snake.move()
		}
	} else {
		snake.move()
	}
}

func main() {
	var autoplay bool
	flag.BoolVar(&autoplay, "a", false, "Autoplay \"local\" strategy")
	flag.Parse()

	snake = newSnake(RealUi(0))
	initScreen()
	defer s.Fini()

	if autoplay {
		automove = newLocalMover(snake)
	}

	snake.reset()
	keys = nil
	scores = nil
	evChan := make(chan tcell.Event, 100)
	quitChan := make(chan struct{}, 1)
	go s.ChannelEvents(evChan, quitChan)
	ticker := time.NewTicker((1 * time.Second) / time.Duration(snake.speed))
EvLoop:
	for {
		s.Show()
		select {
		case <-ticker.C:
			tick()
		case ev := <-evChan:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				snake.reset()
				keys = nil
				scores = nil
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyRune {
					switch ev.Rune() {
					case 'q':
						break EvLoop
					case 'r':
						s.Beep()
						snake.reset()
						keys = nil
						scores = nil
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
