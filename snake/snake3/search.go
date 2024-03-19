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
// where food is.  We can't predict that.  But if we're just trying to make sure that we don't get
// stuck then anyplace is fine for food, just make sure we don't get a point when we get to the fake
// food.  So pick a random spot.  Or always have a fixed spot for it, or a predictable sequence of them?
//
// In general, the local strategy we use for this must avoid randomness.  So the "pick a random
// direction" should probably follow a fixed schedule, because, why not?  What did randomness ever
// do for us, except help us half the time and hurt us the other half?

package main

func newSearchMover(s *Snake) mover {
	// FIXME
	return &localMover{s}
}
