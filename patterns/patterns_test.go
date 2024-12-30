package patterns

import (
	"testing"

	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

func TestABAC(t *testing.T) {
	NewRand(10)
	dir, _ := tensorfs.NewDir("test")
	empty := dir.Float32("empty", 6, 3, 3)
	a := dir.Float32("A", 6, 3, 3)
	b := dir.Float32("B", 6, 3, 3)
	nOn := NFromPct(0.3, 9)
	nDiff := NFromPct(0.4, nOn)
	// AddVocabPermutedBinary(m, "A", 6, 3, 3, 0.3, 0.4)
	PermutedBinaryMinDiff(a, nOn, 1, 0, nDiff)
	// fmt.Println(a)
	// AddVocabDrift(m, "B", 6, 0.2, "A", 0) // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	PermutedBinaryMinDiff(b, nOn, 1, 0, nDiff)
	// fmt.Println(b)
	ctx1 := dir.Float32("ctxt1")
	// AddVocabRepeat(m, "ctxt1", 6, "A", 0)
	ReplicateRows(ctx1, a.SubSpace(0), 6)
	// fmt.Println(ctx1)
	// VocabConcat(m, "AB-C", []string{"A", "B"})
	ab := dir.Float32("ab", 0, 3, 3)
	ab.AppendFrom(a)
	ab.AppendFrom(b)
	// fmt.Println(ab)
	// VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 6, 12}) // 3 cutoffs for 2 vocabs
	SplitRows(dir, ab, []string{"asp", "bsp"}, 6)
	// fmt.Println(dir.Float32("asp"), dir.Float32("bsp"))
	// VocabShuffle(m, []string{"B'"})
	bshuf := Shuffle(b)
	dir.Set("bshuf", bshuf)
	// fmt.Println(bshuf)
	// AddVocabClone(m, "B''", "B'")

	exempty := `empty [6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   0   0 
[0 1]     0   0   0 
[0 2]     0   0   0 
[1 0]     0   0   0 
[1 1]     0   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     0   0   0 
[2 2]     0   0   0 
[3 0]     0   0   0 
[3 1]     0   0   0 
[3 2]     0   0   0 
[4 0]     0   0   0 
[4 1]     0   0   0 
[4 2]     0   0   0 
[5 0]     0   0   0 
[5 1]     0   0   0 
[5 2]     0   0   0 
`
	assert.Equal(t, exempty, empty.String())

	exa := `A [6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     1   0   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     1   0   0 
[2 2]     1   1   0 
[3 0]     0   1   0 
[3 1]     0   0   0 
[3 2]     1   0   1 
[4 0]     1   0   1 
[4 1]     0   1   0 
[4 2]     0   0   0 
[5 0]     1   0   0 
[5 1]     0   0   1 
[5 2]     1   0   0 
`

	assert.Equal(t, exa, a.String())

	exb := `B [6 3 3]
[r r c] [0] [1] [2] 
[0 0]     1   0   0 
[0 1]     0   1   0 
[0 2]     1   0   0 
[1 0]     1   0   0 
[1 1]     1   1   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     0   1   1 
[2 2]     0   0   1 
[3 0]     1   1   0 
[3 1]     0   0   0 
[3 2]     0   1   0 
[4 0]     0   0   0 
[4 1]     1   1   0 
[4 2]     1   0   0 
[5 0]     0   0   0 
[5 1]     0   0   1 
[5 2]     1   1   0 
`

	// drift version:
	// `B [6 3 3]
	// [r r c] [0] [1] [2]
	// [0 0]     0   1   0
	// [0 1]     1   0   1
	// [0 2]     0   0   0
	// [1 0]     0   1   0
	// [1 1]     1   0   1
	// [1 2]     0   0   0
	// [2 0]     0   1   0
	// [2 1]     1   0   1
	// [2 2]     0   0   0
	// [3 0]     0   1   1
	// [3 1]     1   0   0
	// [3 2]     0   0   0
	// [4 0]     0   1   1
	// [4 1]     1   0   0
	// [4 2]     0   0   0
	// [5 0]     0   1   1
	// [5 1]     1   0   0
	// [5 2]     0   0   0
	// `

	assert.Equal(t, exb, b.String())

	exctxt := `ctxt1 [6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     0   1   0 
[1 1]     1   0   1 
[1 2]     0   0   0 
[2 0]     0   1   0 
[2 1]     1   0   1 
[2 2]     0   0   0 
[3 0]     0   1   0 
[3 1]     1   0   1 
[3 2]     0   0   0 
[4 0]     0   1   0 
[4 1]     1   0   1 
[4 2]     0   0   0 
[5 0]     0   1   0 
[5 1]     1   0   1 
[5 2]     0   0   0 
`

	assert.Equal(t, exctxt, ctx1.String())

	exabc := `ab [12 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     1   0   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     1   0   0 
[2 2]     1   1   0 
[3 0]     0   1   0 
[3 1]     0   0   0 
[3 2]     1   0   1 
[4 0]     1   0   1 
[4 1]     0   1   0 
[4 2]     0   0   0 
[5 0]     1   0   0 
[5 1]     0   0   1 
[5 2]     1   0   0 
[6 0]     1   0   0 
[6 1]     0   1   0 
[6 2]     1   0   0 
[7 0]     1   0   0 
[7 1]     1   1   0 
[7 2]     0   0   0 
[8 0]     0   0   0 
[8 1]     0   1   1 
[8 2]     0   0   1 
[9 0]     1   1   0 
[9 1]     0   0   0 
[9 2]     0   1   0 
[10 0]    0   0   0 
[10 1]    1   1   0 
[10 2]    1   0   0 
[11 0]    0   0   0 
[11 1]    0   0   1 
[11 2]    1   1   0 
`

	// fmt.Println(ab)
	assert.Equal(t, exabc, ab.String())

	//////// Mix

	mix := dir.Float32("mix")
	Mix(mix, 12, a, b, ctx1, ctx1, empty, b)
	mix.SetShapeSizes(6, 3, 2, 3, 3)
	// InitPats(dt, "TrainAB", "describe", "Input", "ECout", 6, 3, 2, 3, 3)
	// MixPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})

	// try shuffle
	// Shuffle(dt, []int{0, 1, 2, 3, 4, 5}, []string{"Input", "ECout"}, false)

	exmix := `mix [6 3 2 3 3]
[r r c r c] [0 0] [0 1] [0 2] [1 0] [1 1] [1 2] 
[0 0 0]         0     1     0     1     0     0 
[0 0 1]         1     0     1     0     1     0 
[0 0 2]         0     0     0     1     0     0 
[0 1 0]         0     1     0     0     1     0 
[0 1 1]         1     0     1     1     0     1 
[0 1 2]         0     0     0     0     0     0 
[0 2 0]         0     0     0     1     0     0 
[0 2 1]         0     0     0     0     1     0 
[0 2 2]         0     0     0     1     0     0 
[1 0 0]         1     0     1     1     0     0 
[1 0 1]         1     0     0     1     1     0 
[1 0 2]         0     0     0     0     0     0 
[1 1 0]         0     1     0     0     1     0 
[1 1 1]         1     0     1     1     0     1 
[1 1 2]         0     0     0     0     0     0 
[1 2 0]         0     0     0     1     0     0 
[1 2 1]         0     0     0     1     1     0 
[1 2 2]         0     0     0     0     0     0 
[2 0 0]         0     0     0     0     0     0 
[2 0 1]         1     0     0     0     1     1 
[2 0 2]         1     1     0     0     0     1 
[2 1 0]         0     1     0     0     1     0 
[2 1 1]         1     0     1     1     0     1 
[2 1 2]         0     0     0     0     0     0 
[2 2 0]         0     0     0     0     0     0 
[2 2 1]         0     0     0     0     1     1 
[2 2 2]         0     0     0     0     0     1 
[3 0 0]         0     1     0     1     1     0 
[3 0 1]         0     0     0     0     0     0 
[3 0 2]         1     0     1     0     1     0 
[3 1 0]         0     1     0     0     1     0 
[3 1 1]         1     0     1     1     0     1 
[3 1 2]         0     0     0     0     0     0 
[3 2 0]         0     0     0     1     1     0 
[3 2 1]         0     0     0     0     0     0 
[3 2 2]         0     0     0     0     1     0 
[4 0 0]         1     0     1     0     0     0 
[4 0 1]         0     1     0     1     1     0 
[4 0 2]         0     0     0     1     0     0 
[4 1 0]         0     1     0     0     1     0 
[4 1 1]         1     0     1     1     0     1 
[4 1 2]         0     0     0     0     0     0 
[4 2 0]         0     0     0     0     0     0 
[4 2 1]         0     0     0     1     1     0 
[4 2 2]         0     0     0     1     0     0 
[5 0 0]         1     0     0     0     0     0 
[5 0 1]         0     0     1     0     0     1 
[5 0 2]         1     0     0     1     1     0 
[5 1 0]         0     1     0     0     1     0 
[5 1 1]         1     0     1     1     0     1 
[5 1 2]         0     0     0     0     0     0 
[5 2 0]         0     0     0     0     0     0 
[5 2 1]         0     0     0     0     0     1 
[5 2 2]         0     0     0     1     1     0 
`

	// fmt.Println(mix)
	assert.Equal(t, exmix, mix.String())
}

func TestNameRows(t *testing.T) {
	dir, _ := tensorfs.NewDir("test")
	nr := dir.StringValue("Name", 12)
	NameRows(nr, "AB_", 2)
	// fmt.Println(nr)

	exnm := `Name [12] AB_00 AB_01 AB_02 AB_03 AB_04 AB_05 AB_06 AB_07 AB_08 AB_09 AB_10 
          AB_11 
`

	assert.Equal(t, exnm, nr.String())
}
