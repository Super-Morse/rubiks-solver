package cube

import (
	"errors"
	"fmt"
)

// Sentinel errors returned (wrapped) by the physical-validity checks.
// Use errors.Is to distinguish them; each one corresponds to a distinct
// way a scanned state can fail to be a real, reachable cube.
var (
	// ErrColorCount: some color does not appear exactly 9 times.
	ErrColorCount = errors.New("wrong number of stickers of a color")
	// ErrCenters: center stickers are not U,R,F,D,L,B on their own faces.
	ErrCenters = errors.New("center stickers do not match their faces")
	// ErrDuplicateCorner: a corner piece appears more than once (another is missing).
	ErrDuplicateCorner = errors.New("corner piece appears more than once")
	// ErrDuplicateEdge: an edge piece appears more than once (another is missing).
	ErrDuplicateEdge = errors.New("edge piece appears more than once")
	// ErrCornerTwist: total corner twist is not divisible by 3 (a single corner is twisted in place).
	ErrCornerTwist = errors.New("a corner is twisted")
	// ErrEdgeFlip: total edge flip is odd (a single edge is flipped in place).
	ErrEdgeFlip = errors.New("an edge is flipped")
	// ErrParity: corner and edge permutation parities differ (two pieces are swapped).
	ErrParity = errors.New("two pieces are swapped (parity error)")
)

// centerFacelet holds the index of the center sticker of each face, in Color order.
var centerFacelet = [6]int{4, 13, 22, 31, 40, 49}

// Verify checks facelet-level validity: every color appears exactly 9 times
// and every center sticker carries its own face's color.
func (c *Cube) Verify() error {
	var count [6]int
	for _, col := range c.Facelets {
		if int(col) >= len(count) {
			return fmt.Errorf("%w: invalid color value %d", ErrColorCount, col)
		}
		count[col]++
	}
	for col, n := range count {
		if n != 9 {
			return fmt.Errorf("%w: %c appears %d times, expected 9", ErrColorCount, "URFDLB"[col], n)
		}
	}
	for face, idx := range centerFacelet {
		if c.Facelets[idx] != Color(face) {
			return fmt.Errorf("%w: center of face %c is %c", ErrCenters, "URFDLB"[face], "URFDLB"[c.Facelets[idx]])
		}
	}
	return nil
}

// Verify checks that the cubie state is reachable from the solved cube by
// legal face turns. The conditions below are necessary and sufficient:
//  1. CP and EP are permutations (each piece exactly once),
//  2. orientation values are in range,
//  3. sum(CO) % 3 == 0,
//  4. sum(EO) % 2 == 0,
//  5. parity(CP) == parity(EP).
func (cc *CubieCube) Verify() error {
	var seenC [8]bool
	for i, p := range cc.CP {
		if int(p) >= 8 {
			return fmt.Errorf("%w: invalid corner %d at position %d", ErrDuplicateCorner, p, i)
		}
		if seenC[p] {
			return fmt.Errorf("%w: corner %d at position %d", ErrDuplicateCorner, p, i)
		}
		seenC[p] = true
	}

	var seenE [12]bool
	for i, p := range cc.EP {
		if int(p) >= 12 {
			return fmt.Errorf("%w: invalid edge %d at position %d", ErrDuplicateEdge, p, i)
		}
		if seenE[p] {
			return fmt.Errorf("%w: edge %d at position %d", ErrDuplicateEdge, p, i)
		}
		seenE[p] = true
	}

	twist := 0
	for i, o := range cc.CO {
		if o > 2 {
			return fmt.Errorf("%w: invalid orientation %d at corner position %d", ErrCornerTwist, o, i)
		}
		twist += int(o)
	}
	if twist%3 != 0 {
		return fmt.Errorf("%w: total twist %d is not divisible by 3", ErrCornerTwist, twist)
	}

	flip := 0
	for i, o := range cc.EO {
		if o > 1 {
			return fmt.Errorf("%w: invalid orientation %d at edge position %d", ErrEdgeFlip, o, i)
		}
		flip += int(o)
	}
	if flip%2 != 0 {
		return fmt.Errorf("%w: total flip %d is odd", ErrEdgeFlip, flip)
	}

	if cornerParity(cc.CP) != edgeParity(cc.EP) {
		return ErrParity
	}
	return nil
}

// cornerParity returns 0 for an even permutation, 1 for odd (inversion count mod 2).
func cornerParity(p [8]Corner) int {
	inv := 0
	for i := 0; i < len(p); i++ {
		for j := i + 1; j < len(p); j++ {
			if p[i] > p[j] {
				inv++
			}
		}
	}
	return inv % 2
}

// edgeParity returns 0 for an even permutation, 1 for odd (inversion count mod 2).
func edgeParity(p [12]Edge) int {
	inv := 0
	for i := 0; i < len(p); i++ {
		for j := i + 1; j < len(p); j++ {
			if p[i] > p[j] {
				inv++
			}
		}
	}
	return inv % 2
}
