// Copyright (c) 2022 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package yamc

import (
	"fmt"
	"strconv"
)

// FindChain is trying to get a value from different types of slices and maps.
// this are limited to the most used data structures, they needed to work with
// json and yaml content
func FindChain(in interface{}, byChain ...string) (data interface{}, err error) {
	// nil is for sure nothing thats needs being handled.
	// we throw an error. any call must be direct, or by an recursive call.
	// recursive calls have to make sure they use valid data.
	if in == nil {
		return nil, fmt.Errorf("nil can not being parsed")
	}

	switch v := in.(type) {
	case []interface{}:
		// the key for arrays is allways the numeric index submitted as string
		if index, err := strconv.Atoi(byChain[0]); err != nil {
			return nil, err
		} else {
			// just checking the possible index range
			if len(v) > index && index >= 0 {
				newChain := byChain[1:] // reduce chain slice
				data = v[index]         // getting data by index
				if len(newChain) == 0 { // if we no entires left in the slice, we have found the value
					if CanHandleValue(data) { // we need to verify we can handle the value. structs for example are not supported
						return data, nil
					} else {
						return nil, fmt.Errorf("[]interface{} unsupported value type %T", data)
					}
				} else {
					return FindChain(data, newChain...) // we have keys left in the chain slice
				}
			} else {
				return nil, fmt.Errorf("index %v is out of range. have  %d entries in %T. starting at 0 max index is %d", index, len(v), v, len(v)-1)
			}
		}
	case map[interface{}]interface{}: // here again close the same logic as for array. just data is get diffent
		if data, ok := v[byChain[0]]; ok {
			newChain := byChain[1:]
			if len(newChain) == 0 {
				if CanHandleValue(data) {
					return data, nil
				}
				return nil, fmt.Errorf("map[interface{}]interface{} unsupported value type %T", data)
			} else {
				return FindChain(data, newChain...)
			}
		} else {
			return nil, fmt.Errorf("the map do not contains the key %s in %T", byChain[0], data)
		}
	case map[string]interface{}: // last supported type
		if data, ok := v[byChain[0]]; ok {
			newChain := byChain[1:]
			if len(newChain) == 0 {
				if CanHandleValue(data) {
					return data, nil
				}
				return nil, fmt.Errorf("unsupported value type %T", data)
			} else {
				return FindChain(data, newChain...)
			}
		} else { // not found an entriy with the key
			return nil, fmt.Errorf("the map do not contains the key %s in %T", byChain[0], data)
		}
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}

// CanHandleValue just make sure we do not run in panic because of an dataype we nor able to handle
func CanHandleValue(value any) bool {
	switch value.(type) {
	case int, int16, int32, int64, float32, float64, string, bool, byte, map[string]interface{}:
		return true
	default:
		return false
	}

}
