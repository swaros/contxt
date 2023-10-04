package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestRoundInSerial(t *testing.T) {
	for maxInt := 10; maxInt < 1000; maxInt++ {
		for workValue := 1; workValue < 100; workValue++ {

			left := workValue
			right := 100 - workValue

			// make sure that we have a 100% result
			if left+right != 100 {
				t.Errorf("TestRoundInSerial: left(%d) + right(%d) != 100", left, right)
			}

			// calculate the percentage
			pLeft, _, rest := ctxout.RoundHelp(left, maxInt)
			// you could also use the ctxout.RoundHelp again without using the rest value
			// just to see that is not working. The result is not 100% anymore
			pRight, _, _ := ctxout.RoundHelpWithOffest(right, maxInt, rest)

			// make sure that we have a 100% result
			if pLeft+pRight != maxInt {
				t.Errorf("TestRoundInSerial: pLeft(%d) + pRight(%d) != maxInt(%d)", pLeft, pRight, maxInt)
			}

		}
	}
}

func TestRoundSerial(t *testing.T) {
	for maxInt := 10; maxInt < 1000; maxInt++ {
		for workValue := 1; workValue < 100; workValue++ {

			left := workValue
			right := 100 - workValue

			// make sure that we have a 100% result
			if left+right != 100 {
				t.Errorf("TestRoundInSerial: left(%d) + right(%d) != 100", left, right)
			}

			serialRound := ctxout.NewRoundSerial().SetMax(maxInt).Next()
			pleft := serialRound.Round(left)
			pright := serialRound.Round(right)

			// make sure that we have a 100% result
			if pleft+pright != maxInt {
				t.Errorf("TestRoundInSerial: pleft(%d) + pright(%d) != maxInt(%d)", pleft, pright, maxInt)
			}

		}
	}
}

func TestRoundSerialAsArg(t *testing.T) {
	for maxInt := 10; maxInt < 1000; maxInt++ {
		for workValue := 1; workValue < 100; workValue++ {

			left := workValue
			right := 100 - workValue

			// make sure that we have a 100% result
			if left+right != 100 {
				t.Errorf("TestRoundInSerial: left(%d) + right(%d) != 100", left, right)
			}

			serialRound := ctxout.NewRoundSerial().SetMax(maxInt)
			res := serialRound.RoundArgs(left, right)
			if len(res) != 2 {
				t.Errorf("TestRoundInSerial: len(res) != 2")
			} else {

				fromResult, r1 := serialRound.GetResult(0)
				if !r1 {
					t.Errorf("TestRoundInSerial: !r1")
				} else {
					if fromResult != res[0] {
						t.Errorf("TestRoundInSerial: fromResult(%d) != res[0](%d)", fromResult, res[0])
					}
				}

				fromResult, r2 := serialRound.GetResult(1)
				if !r2 {
					t.Errorf("TestRoundInSerial: !r2")
				} else {
					if fromResult != res[1] {
						t.Errorf("TestRoundInSerial: fromResult(%d) != res[1](%d)", fromResult, res[1])
					}
				}

				// make sure that we have a 100% result
				if res[0]+res[1] != maxInt {
					t.Errorf("TestRoundInSerial: res[0](%d) + res[1](%d) != maxInt(%d)", res[0], res[1], maxInt)
				}
			}

		}
	}
}
