// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
