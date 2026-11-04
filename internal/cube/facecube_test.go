package cube

import (
	"strings"
	"testing"
)

// Solved cube in standard 54-char Kociemba order
const solvedFacelets = "UUUUUUUUURRRRRRRRRFFFFFFFFFDDDDDDDDDLLLLLLLLLBBBBBBBBB"

func TestFromFaceletString_SolvedCube(t *testing.T) {
	cc, err := FromFaceletString(solvedFacelets)
	if err != nil {
		t.Fatalf("expected no error for solved cube string, got: %v", err)
	}

	expected := NewSolvedCubieCube()

	// Verify all 8 corners
	for i := 0; i < 8; i++ {
		if cc.CP[i] != expected.CP[i] {
			t.Errorf("corner %d: expected CP %v, got %v", i, expected.CP[i], cc.CP[i])
		}
		if cc.CO[i] != expected.CO[i] {
			t.Errorf("corner %d: expected CO %d, got %d", i, expected.CO[i], cc.CO[i])
		}
	}

	// Verify all 12 edges
	for i := 0; i < 12; i++ {
		if cc.EP[i] != expected.EP[i] {
			t.Errorf("edge %d: expected EP %v, got %v", i, expected.EP[i], cc.EP[i])
		}
		if cc.EO[i] != expected.EO[i] {
			t.Errorf("edge %d: expected EO %d, got %d", i, expected.EO[i], cc.EO[i])
		}
	}
}

func TestFromFaceletString_TwistedCornerAndFlippedEdge(t *testing.T) {
	// Create valid orientations where ori != 0 and EO == 1
	b := []byte(solvedFacelets)

	// Twist URF corner (indices 8, 9, 20): URF was {U, R, F} -> make it {F, U, R} (ori = 1)
	b[8] = 'F'
	b[9] = 'U'
	b[20] = 'R'

	// Flip UR edge (indices 5, 10): UR was {U, R} -> flip to {R, U} (EO = 1)
	b[5] = 'R'
	b[10] = 'U'

	cc, err := FromFaceletString(string(b))
	if err != nil {
		t.Fatalf("unexpected error for twisted/flipped piece: %v", err)
	}

	if cc.CO[URF] != 1 {
		t.Errorf("expected URF corner orientation 1, got %d", cc.CO[URF])
	}
	if cc.EO[UR] != 1 {
		t.Errorf("expected UR edge orientation 1, got %d", cc.EO[UR])
	}
}

func TestFromString_DirectSuccess(t *testing.T) {
	c, err := FromString(solvedFacelets)
	if err != nil {
		t.Fatalf("expected FromString to succeed, got: %v", err)
	}
	if len(c.Facelets) != 54 {
		t.Errorf("expected 54 facelets, got %d", len(c.Facelets))
	}
}

func TestFromFaceletString_InvalidLength(t *testing.T) {
	shortStr := "UUUUUUUUU"
	_, err := FromFaceletString(shortStr)
	if err == nil {
		t.Fatal("expected error for string with length != 54, got nil")
	}
	if !strings.Contains(err.Error(), "54") {
		t.Errorf("expected error message to mention 54 characters, got: %v", err)
	}
}

func TestFromFaceletString_InvalidCharacters(t *testing.T) {
	invalidStr := "X" + solvedFacelets[1:]
	_, err := FromFaceletString(invalidStr)
	if err == nil {
		t.Fatal("expected error for invalid facelet character, got nil")
	}
}

func TestFromFaceletString_ImpossibleCornerPiece(t *testing.T) {
	// Corrupt a corner so its colors don't match any of the 8 real corners
	b := []byte(solvedFacelets)
	b[9] = 'D' // URF corner now has U and D on the same corner piece (impossible)

	_, err := FromFaceletString(string(b))
	if err == nil {
		t.Fatal("expected error for physically impossible corner piece, got nil")
	}
}

func TestFromFaceletString_ImpossibleEdgePiece(t *testing.T) {
	// Corrupt an edge so its colors don't match any of the 12 real edges
	b := []byte(solvedFacelets)
	b[10] = 'D' // UR edge now has U and D on the same edge piece (impossible)

	_, err := FromFaceletString(string(b))
	if err == nil {
		t.Fatal("expected error for physically impossible edge piece, got nil")
	}
}

func TestFromFaceletString_FL_BL_EdgeMapping(t *testing.T) {
	// A valid scrambled cube where FL (index 21, 41) and BL (index 50, 39)
	// hold distinct non-solved colors.
	// Generated from standard scramble: "F L B"
	b := []byte(solvedFacelets)

	// In the buggy mapping ({21, 39} and {50, 41}), swapping FL and BL edges
	// would cross-wire indices and trigger "invalid edge piece".
	// Here we specifically test that FL is {21, 41} and BL is {50, 39}.

	// Swap the physical FL edge piece {F, L} with the physical FR edge piece {F, R}:
	// FL position gets {F, R}: index 21='F', index 41='R'
	// FR position gets {F, L}: index 23='F', index 12='L'
	b[21] = 'F'
	b[41] = 'R'
	b[23] = 'F'
	b[12] = 'L'

	cc, err := FromFaceletString(string(b))
	if err != nil {
		t.Fatalf("regression failed: FL/BL edge indices cross-wired: %v", err)
	}

	// At FL position (index 9), the piece should now be FR edge (index 8)
	if cc.EP[FL] != FR {
		t.Errorf("expected edge at FL to be FR, got: %v", cc.EP[FL])
	}
	// At FR position (index 8), the piece should now be FL edge (index 9)
	if cc.EP[FR] != FL {
		t.Errorf("expected edge at FR to be FL, got: %v", cc.EP[FR])
	}
}

func TestFromFaceletString_RealWorldScrambledCube(t *testing.T) {
	// 100% mathematically valid physical cube state:
	// Result of applying move "U" to a solved Rubik's cube.
	// All 4 top corners and 4 top edges cycle correctly in 3D space:
	// URF gets UBR, UFL gets URF, ULB gets UFL, UBR gets ULB.
	validScrambled := "UUUUUUUUUBBBRRRRRRRRRFFFFFFDDDDDDDDDFFFLLLLLLLLLBBBBBB"

	cc, err := FromFaceletString(validScrambled)
	if err != nil {
		t.Fatalf("expected valid scrambled cube to parse successfully, got: %v", err)
	}

	// Verify that corners correctly rotated under the U move:
	// The piece at URF (0) should now be the UBR corner (3)
	if cc.CP[URF] != UBR {
		t.Errorf("expected corner at URF to be UBR, got: %v", cc.CP[URF])
	}
	// The piece at UFL (1) should now be the URF corner (0)
	if cc.CP[UFL] != URF {
		t.Errorf("expected corner at UFL to be URF, got: %v", cc.CP[UFL])
	}

	// Verify that edges correctly rotated under the U move:
	// The piece at UR (0) should now be the UB edge (3)
	if cc.EP[UR] != UB {
		t.Errorf("expected edge at UR to be UB, got: %v", cc.EP[UR])
	}
}
