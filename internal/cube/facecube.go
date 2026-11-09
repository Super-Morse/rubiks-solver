package cube

import (
	"errors"
	"fmt"
)

// Standard Kociemba Facelet positions for Corners: { U/D, Clockwise, Counter-Clockwise }
var cornerFacelet = [8][3]int{
	{8, 9, 20},   // URF
	{6, 18, 38},  // UFL
	{0, 36, 47},  // ULB
	{2, 45, 11},  // UBR
	{29, 26, 15}, // DFR
	{27, 44, 24}, // DLF
	{33, 53, 42}, // DBL
	{35, 17, 51}, // DRB
}

var cornerColor = [8][3]Color{
	{U, R, F},
	{U, F, L},
	{U, L, B},
	{U, B, R},
	{D, F, R},
	{D, L, F},
	{D, B, L},
	{D, R, B},
}

// Standard Kociemba Facelet positions for Edges
var edgeFacelet = [12][2]int{
	{5, 10},  // UR
	{7, 19},  // UF
	{3, 37},  // UL
	{1, 46},  // UB
	{32, 16}, // DR
	{28, 25}, // DF
	{30, 43}, // DL
	{34, 52}, // DB
	{23, 12}, // FR
	{21, 41}, // FL
	{50, 39}, // BL
	{48, 14}, // BR
}

var edgeColor = [12][2]Color{
	{U, R},
	{U, F},
	{U, L},
	{U, B},
	{D, R},
	{D, F},
	{D, L},
	{D, B},
	{F, R},
	{F, L},
	{B, L},
	{B, R},
}

var charToColor = map[byte]Color{
	'U': U,
	'R': R,
	'F': F,
	'D': D,
	'L': L,
	'B': B,
}

// FromString parses a 54-character Kociemba string into a Cube (54 facelets).
func FromString(s string) (*Cube, error) {
	if len(s) != 54 {
		return nil, fmt.Errorf("cube string must have exactly 54 characters, got %d", len(s))
	}

	var facelets [54]Color
	for i := 0; i < 54; i++ {
		c, ok := charToColor[s[i]]
		if !ok {
			return nil, fmt.Errorf("invalid color character '%c' at index %d", s[i], i)
		}
		facelets[i] = c
	}

	return &Cube{Facelets: facelets}, nil
}

// ToCubieCube converts 54 facelets into a CubieCube (corner/edge permutations and orientations).
func (c *Cube) ToCubieCube() (*CubieCube, error) {
	cc := &CubieCube{}

	// 1. Resolve 8 Corners
	for i := 0; i < 8; i++ {
		var ori byte
		for ori = 0; ori < 3; ori++ {
			fCol := c.Facelets[cornerFacelet[i][ori]]
			if fCol == U || fCol == D {
				break
			}
		}

		col1 := c.Facelets[cornerFacelet[i][(ori+1)%3]]
		col2 := c.Facelets[cornerFacelet[i][(ori+2)%3]]

		matched := false
		for cIdx := 0; cIdx < 8; cIdx++ {
			if col1 == cornerColor[cIdx][1] && col2 == cornerColor[cIdx][2] {
				cc.CP[i] = Corner(cIdx)
				cc.CO[i] = ori % 3
				matched = true
				break
			}
		}

		if !matched {
			return nil, errors.New("invalid corner piece: colors do not correspond to any physical corner")
		}
	}

	// 2. Resolve 12 Edges
	for i := 0; i < 12; i++ {
		f0 := c.Facelets[edgeFacelet[i][0]]
		f1 := c.Facelets[edgeFacelet[i][1]]

		matched := false
		for eIdx := 0; eIdx < 12; eIdx++ {
			if f0 == edgeColor[eIdx][0] && f1 == edgeColor[eIdx][1] {
				cc.EP[i] = Edge(eIdx)
				cc.EO[i] = 0
				matched = true
				break
			}
			if f0 == edgeColor[eIdx][1] && f1 == edgeColor[eIdx][0] {
				cc.EP[i] = Edge(eIdx)
				cc.EO[i] = 1
				matched = true
				break
			}
		}

		if !matched {
			return nil, errors.New("invalid edge piece: colors do not correspond to any physical edge")
		}
	}

	return cc, nil
}

// FromFaceletString parses a 54-character Kociemba string, converts it to a
// CubieCube and verifies that the result is a physically reachable cube state.
// Use FromString + ToCubieCube directly to skip the reachability checks.
func FromFaceletString(s string) (*CubieCube, error) {
	cube, err := FromString(s)
	if err != nil {
		return nil, err
	}
	if err := cube.Verify(); err != nil {
		return nil, err
	}
	cc, err := cube.ToCubieCube()
	if err != nil {
		return nil, err
	}
	if err := cc.Verify(); err != nil {
		return nil, err
	}
	return cc, nil
}
