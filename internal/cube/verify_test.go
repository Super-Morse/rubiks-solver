package cube

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

// --- helpers ---------------------------------------------------------------

// toFacelets is the inverse of ToCubieCube: it paints the 54 stickers of a
// CubieCube using the same facelet tables. It lets tests generate valid
// states from move sequences and corrupt them at the piece level.
func toFacelets(cc *CubieCube) string {
	var f [54]Color
	for face, idx := range centerFacelet {
		f[idx] = Color(face)
	}
	for i := 0; i < 8; i++ {
		piece, ori := cc.CP[i], int(cc.CO[i])
		for n := 0; n < 3; n++ {
			f[cornerFacelet[i][(n+ori)%3]] = cornerColor[piece][n]
		}
	}
	for i := 0; i < 12; i++ {
		piece, ori := cc.EP[i], int(cc.EO[i])
		for n := 0; n < 2; n++ {
			f[edgeFacelet[i][(n+ori)%2]] = edgeColor[piece][n]
		}
	}
	var sb strings.Builder
	for _, c := range f {
		sb.WriteByte("URFDLB"[c])
	}
	return sb.String()
}

func randomCubie(rng *rand.Rand, n int) *CubieCube {
	cc := NewSolvedCubieCube()
	for i := 0; i < n; i++ {
		cc.ApplyMove(Move(rng.Intn(18)))
	}
	return cc
}

func expectErr(t *testing.T, err, want error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %v, got nil", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

// --- helper self-check -----------------------------------------------------

func TestToFacelets_RoundTrip(t *testing.T) {
	if got := toFacelets(NewSolvedCubieCube()); got != solvedFacelets {
		t.Fatalf("solved cube painted as %s", got)
	}
	rng := rand.New(rand.NewSource(1))
	for k := 0; k < 200; k++ {
		cc := randomCubie(rng, 30)
		back, err := mustParseRaw(t, toFacelets(cc))
		if err != nil {
			t.Fatalf("round trip parse failed: %v", err)
		}
		if *back != *cc {
			t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", back, cc)
		}
	}
}

// --- valid states ----------------------------------------------------------

func TestVerify_SolvedCube(t *testing.T) {
	if err := NewSolvedCubieCube().Verify(); err != nil {
		t.Fatalf("solved cubie cube rejected: %v", err)
	}
	if err := NewSolvedCube().Verify(); err != nil {
		t.Fatalf("solved facelet cube rejected: %v", err)
	}
	if _, err := FromFaceletString(solvedFacelets); err != nil {
		t.Fatalf("solved string rejected: %v", err)
	}
}

func TestVerify_EverySingleMove(t *testing.T) {
	for m := Move(0); m < 18; m++ {
		cc := NewSolvedCubieCube()
		cc.ApplyMove(m)
		if err := cc.Verify(); err != nil {
			t.Errorf("move %d: %v", m, err)
		}
		if _, err := FromFaceletString(toFacelets(cc)); err != nil {
			t.Errorf("move %d via facelets: %v", m, err)
		}
	}
}

func TestVerify_RandomScramblesAreValid(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for k := 0; k < 1000; k++ {
		cc := randomCubie(rng, 1+rng.Intn(40))
		if err := cc.Verify(); err != nil {
			t.Fatalf("scramble %d rejected: %v", k, err)
		}
		if _, err := FromFaceletString(toFacelets(cc)); err != nil {
			t.Fatalf("scramble %d rejected via facelets: %v", k, err)
		}
	}
}

func TestVerify_KnownRealState(t *testing.T) {
	// State after a single U turn, written out by hand.
	if _, err := FromFaceletString("UUUUUUUUUBBBRRRRRRRRRFFFFFFDDDDDDDDDFFFLLLLLLLLLBBBBBB"); err != nil {
		t.Fatalf("U-turn state rejected: %v", err)
	}
}

// Compensating defects must cancel out: these states are reachable.
func TestVerify_CompensatingDefectsAreValid(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for k := 0; k < 200; k++ {
		base := randomCubie(rng, 25)

		// Two corners twisted in opposite directions.
		c := *base
		i, j := rng.Intn(8), rng.Intn(8)
		for j == i {
			j = rng.Intn(8)
		}
		c.CO[i] = (c.CO[i] + 1) % 3
		c.CO[j] = (c.CO[j] + 2) % 3
		if err := c.Verify(); err != nil {
			t.Fatalf("two opposite twists rejected: %v", err)
		}

		// Three corners twisted the same way.
		c = *base
		for _, p := range rng.Perm(8)[:3] {
			c.CO[p] = (c.CO[p] + 1) % 3
		}
		if err := c.Verify(); err != nil {
			t.Fatalf("three equal twists rejected: %v", err)
		}

		// Two edges flipped.
		c = *base
		e := rng.Perm(12)
		c.EO[e[0]] ^= 1
		c.EO[e[1]] ^= 1
		if err := c.Verify(); err != nil {
			t.Fatalf("two flips rejected: %v", err)
		}

		// One corner swap plus one edge swap (parities cancel).
		c = *base
		ci := rng.Perm(8)
		c.CP[ci[0]], c.CP[ci[1]] = c.CP[ci[1]], c.CP[ci[0]]
		c.EP[e[0]], c.EP[e[1]] = c.EP[e[1]], c.EP[e[0]]
		if err := c.Verify(); err != nil {
			t.Fatalf("corner+edge swap rejected: %v", err)
		}

		// A 3-cycle of corners (even permutation).
		c = *base
		c.CP[ci[0]], c.CP[ci[1]], c.CP[ci[2]] = c.CP[ci[1]], c.CP[ci[2]], c.CP[ci[0]]
		if err := c.Verify(); err != nil {
			t.Fatalf("corner 3-cycle rejected: %v", err)
		}
	}
}

// --- invalid states: piece level ------------------------------------------

func TestVerify_SingleCornerTwist(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	for k := 0; k < 200; k++ {
		cc := randomCubie(rng, 25)
		i := rng.Intn(8)
		cc.CO[i] = (cc.CO[i] + byte(1+rng.Intn(2))) % 3
		expectErr(t, cc.Verify(), ErrCornerTwist)
		_, err := FromFaceletString(toFacelets(cc))
		expectErr(t, err, ErrCornerTwist)
	}
}

func TestVerify_SingleEdgeFlip(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	for k := 0; k < 200; k++ {
		cc := randomCubie(rng, 25)
		cc.EO[rng.Intn(12)] ^= 1
		expectErr(t, cc.Verify(), ErrEdgeFlip)
		_, err := FromFaceletString(toFacelets(cc))
		expectErr(t, err, ErrEdgeFlip)
	}
}

func TestVerify_ThreeEdgeFlips(t *testing.T) {
	cc := NewSolvedCubieCube()
	cc.EO[0], cc.EO[5], cc.EO[11] = 1, 1, 1
	expectErr(t, cc.Verify(), ErrEdgeFlip)
}

func TestVerify_TwoCornersSameTwist(t *testing.T) {
	cc := NewSolvedCubieCube()
	cc.CO[0], cc.CO[1] = 1, 1 // total 2, not divisible by 3
	expectErr(t, cc.Verify(), ErrCornerTwist)
}

func TestVerify_SingleCornerSwap(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	for k := 0; k < 200; k++ {
		cc := randomCubie(rng, 25)
		p := rng.Perm(8)
		cc.CP[p[0]], cc.CP[p[1]] = cc.CP[p[1]], cc.CP[p[0]]
		expectErr(t, cc.Verify(), ErrParity)
		_, err := FromFaceletString(toFacelets(cc))
		expectErr(t, err, ErrParity)
	}
}

func TestVerify_SingleEdgeSwap(t *testing.T) {
	rng := rand.New(rand.NewSource(14))
	for k := 0; k < 200; k++ {
		cc := randomCubie(rng, 25)
		p := rng.Perm(12)
		cc.EP[p[0]], cc.EP[p[1]] = cc.EP[p[1]], cc.EP[p[0]]
		expectErr(t, cc.Verify(), ErrParity)
		_, err := FromFaceletString(toFacelets(cc))
		expectErr(t, err, ErrParity)
	}
}

func TestVerify_DuplicateCorner(t *testing.T) {
	cc := NewSolvedCubieCube()
	cc.CP[DRB] = URF
	expectErr(t, cc.Verify(), ErrDuplicateCorner)
}

func TestVerify_DuplicateEdge(t *testing.T) {
	cc := NewSolvedCubieCube()
	cc.EP[BR] = UR
	expectErr(t, cc.Verify(), ErrDuplicateEdge)
}

func TestVerify_OutOfRangeValues(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*CubieCube)
		want error
	}{
		{"corner id", func(c *CubieCube) { c.CP[0] = 8 }, ErrDuplicateCorner},
		{"edge id", func(c *CubieCube) { c.EP[0] = 12 }, ErrDuplicateEdge},
		// 3 would pass the mod-3 sum check if not range-checked.
		{"corner ori", func(c *CubieCube) { c.CO[0] = 3 }, ErrCornerTwist},
		// 2 would pass the parity sum check if not range-checked.
		{"edge ori", func(c *CubieCube) { c.EO[0] = 2 }, ErrEdgeFlip},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := NewSolvedCubieCube()
			tc.mut(cc)
			expectErr(t, cc.Verify(), tc.want)
		})
	}
}

// --- invalid states: facelet level ----------------------------------------

func TestVerify_DuplicatePieceFromScan(t *testing.T) {
	// Repaint the UBR corner as a second URF corner (U,R,F). Each piece is a
	// real piece, so ToCubieCube succeeds; only the verifier catches it.
	// Color counts break too, so check the cubie-level error directly.
	// UBR slot facelets {2, 45, 11} painted U,R,F = URF corner, orientation 0.
	b := []byte(solvedFacelets)
	b[2], b[45], b[11] = 'U', 'R', 'F'
	cc, err := mustParseRaw(t, string(b))
	if err != nil {
		t.Fatalf("could not build duplicate-corner scan: %v", err)
	}
	if cc.CP[UBR] != URF {
		t.Fatalf("expected URF in UBR slot, got %v", cc.CP[UBR])
	}
	expectErr(t, cc.Verify(), ErrDuplicateCorner)

	_, err = FromFaceletString(string(b))
	if err == nil {
		t.Fatal("duplicate corner scan accepted")
	}
}

func TestVerify_ColorCount(t *testing.T) {
	// Misread one F sticker as R (classic red/orange confusion).
	b := []byte(solvedFacelets)
	b[18] = 'R'
	_, err := FromFaceletString(string(b))
	expectErr(t, err, ErrColorCount)

	c, _ := FromString(string(b))
	expectErr(t, c.Verify(), ErrColorCount)
}

func TestVerify_InvalidColorValue(t *testing.T) {
	// FromString cannot produce this, but a hand-built Cube can.
	c := NewSolvedCube()
	c.Facelets[0] = Color(6)
	expectErr(t, c.Verify(), ErrColorCount)
}

func TestVerify_Centers(t *testing.T) {
	// Swap two center stickers: counts stay 9 each, centers are wrong.
	b := []byte(solvedFacelets)
	b[4], b[13] = b[13], b[4]
	_, err := FromFaceletString(string(b))
	expectErr(t, err, ErrCenters)
}

func TestVerify_WholeFaceRotatedIsNotAnError(t *testing.T) {
	// Rotating a face's stickers as a block is exactly a face turn: valid.
	cc := NewSolvedCubieCube()
	cc.ApplyMove(MoveF)
	if _, err := FromFaceletString(toFacelets(cc)); err != nil {
		t.Fatalf("F turn rejected: %v", err)
	}
}

func TestVerify_StickerSwapOnOnePiece(t *testing.T) {
	// Peeling and re-sticking two stickers of one corner reverses its cyclic
	// order, making it a mirror piece that does not exist.
	b := []byte(solvedFacelets)
	b[9], b[20] = b[20], b[9] // URF: U,R,F -> U,F,R
	_, err := FromFaceletString(string(b))
	if err == nil {
		t.Fatal("mirrored corner accepted")
	}
}

func TestVerify_ErrorsAreDistinct(t *testing.T) {
	all := []error{ErrColorCount, ErrCenters, ErrDuplicateCorner, ErrDuplicateEdge, ErrCornerTwist, ErrEdgeFlip, ErrParity}
	for i := range all {
		for j := range all {
			if i != j && errors.Is(all[i], all[j]) {
				t.Errorf("%v matches %v", all[i], all[j])
			}
		}
	}
}

// --- parity helpers -------------------------------------------------------

func TestParityHelpers(t *testing.T) {
	id := NewSolvedCubieCube()
	if cornerParity(id.CP) != 0 || edgeParity(id.EP) != 0 {
		t.Fatal("identity must be even")
	}
	c := id.CP
	c[0], c[1] = c[1], c[0]
	if cornerParity(c) != 1 {
		t.Error("transposition must be odd")
	}
	c[2], c[3], c[4] = c[3], c[4], c[2]
	if cornerParity(c) != 1 {
		t.Error("transposition + 3-cycle must stay odd")
	}
	e := id.EP
	e[0], e[1], e[2], e[3] = e[1], e[2], e[3], e[0] // 4-cycle is odd
	if edgeParity(e) != 1 {
		t.Error("4-cycle must be odd")
	}
	// Each quarter turn is a 4-cycle on corners and on edges: parities equal.
	for m := Move(0); m < 18; m++ {
		cc := NewSolvedCubieCube()
		cc.ApplyMove(m)
		if cornerParity(cc.CP) != edgeParity(cc.EP) {
			t.Errorf("move %d: parities differ", m)
		}
	}
}
