package ctxout

import "math"

type roundSerial struct {
	floatDiff  float64
	lastResult float64
	max        int
	resultSet  []int
}

func NewRoundSerial() *roundSerial {
	return &roundSerial{
		floatDiff: 0.0,
		max:       100,
		resultSet: []int{},
	}
}

func (rs *roundSerial) SetMax(max int) *roundSerial {
	rs.max = max
	return rs
}

func (rs *roundSerial) SetFloatDiff(floatDiff float64) *roundSerial {
	rs.floatDiff = floatDiff
	return rs
}

func (rs *roundSerial) GetFloatDiff() float64 {
	return rs.floatDiff
}

func (rs *roundSerial) Round(percentage int) int {
	var intValue int
	intValue, rs.lastResult, rs.floatDiff = RoundHelpWithOffest(percentage, rs.max, rs.floatDiff)
	rs.resultSet = append(rs.resultSet, intValue)
	return int(rs.lastResult)
}

func (rs *roundSerial) GetResultSet() []int {
	return rs.resultSet
}

func (rs *roundSerial) GetResult(index int) (int, bool) {
	if index < len(rs.resultSet) {
		return rs.resultSet[index], true
	}
	return 0, false
}

func (rs *roundSerial) GetLastResult() float64 {
	return rs.lastResult
}

func (rs *roundSerial) RoundArgs(args ...int) []int {
	results := make([]int, len(args))
	rs.Next()
	for i, arg := range args {
		results[i] = rs.Round(arg)
	}

	return results
}

// Next will reset the resultSet and the floatDiff
// this is useful if you want to start a new round
func (rs *roundSerial) Next() *roundSerial {
	rs.floatDiff = 0.0
	rs.resultSet = []int{}
	return rs
}

func RoundHelp(percentage int, max int) (intResult int, result float64, rest float64) {
	fpercent := (float64(max) * float64(percentage)) / 100
	result = math.RoundToEven(fpercent)
	rest = fpercent - result
	intResult = int(result)
	return
}

func RoundHelpWithOffest(percentage int, max int, offset float64) (intResult int, result float64, rest float64) {
	fpercent := (float64(max) * float64(percentage)) / 100
	result = math.RoundToEven(fpercent + offset)
	rest = fpercent - result
	intResult = int(result)
	return
}

func Round(percentage int, max int) int {
	_, result, _ := RoundHelp(percentage, max)
	return int(result)
}
