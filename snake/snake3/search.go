// A mover that searches a little.
//
// Basic approach: Generate all possible move sequences of length up to N (N very short, say, <= 3)
// from the current position, then starting in the final position of that sequence play the local
// strategy for K moves (K larger, say, 100) or until we're stuck or food is found.  We score:
// finding food is 1 point, getting stuck is -2 points, just farting around is 0 points.  Then pick
// the initial move sequence that has the highest payoff (or one of them, and we can stop once we've
// found one that yields > 0 points, because we won't know any better - we can't tell "good"
// sequences apart).  (Or can we?  Number of moves as a tie breaker?)
//
// There could be some penalty for "being in a tunnel" when we finish the K moves.  (Another tie
// breaker?) But there is no reason per se to have tunnel avoidance as part of the local strategy?
//
// The problem is what to do when we reach the food.  Ideally we keep playing until K is exhausted
// to make sure we don't get stuck after.  But this presupposes a goal to move towards - a new place
// where food is, or at least a new place that is our target.  We can't predict where the next food
// will be.  But if we're just trying to make sure that we don't get stuck then any place is fine as
// a target, just make sure we don't score a point when we get to the fake food.  So we can pick a
// random spot, or always have a fixed spot for it, or a predictable sequence of them.
//
// In general, the local strategy we use for this should avoid randomness, or the search result may
// easily become invalid when we actually execute moves.  We could fix that by storing a "plan" to
// follow but that's for later.

// It's possible K should be larger - manhattan distance to food * 1.5 would make a lot of sense.  But
// then we may be taking too long.

// It's possible that *after eating* we allow for a long search to make sure we don't get stuck.  And
// that we should be exploring paths partly DFS, partly BFS.  Consider: generate some initial moves I1, ..., In.
// For each, search 10 steps ahead (whatever).  This will cull some moves.  If we don't find food, we still
// have no more than n locations.  Now search another 10 steps from each of those.  And so on.  Then after finding food,
// probe deeply to make sure there's not a trap.

// If we find food we could record the moves and just perform them, esp if finding food triggers a
// guard against getting stuck....

package main

import "fmt"

const (
	initialMoves = 1			// N
	depth = 100					// K
)

type searchMover struct /* implements mover */ {
	s *Snake
}

func newSearchMover(s *Snake) mover {
	return &searchMover{s}
}

func (_ *searchMover) name() string {
	return fmt.Sprintf("Search(%d,%d)", initialMoves, depth)
}

func (sm *searchMover) autoMove() {
	newLocalMover(sm.s, true).autoMove()
}

