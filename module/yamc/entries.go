package yamc

import (
	"fmt"
	"strconv"
)

func FindChain(in interface{}, byChain ...string) (data interface{}, err error) {
	if in == nil {
		return nil, fmt.Errorf("nil can not being parsed")
	}
	switch v := in.(type) {
	case []interface{}:
		if index, err := strconv.Atoi(byChain[0]); err != nil {
			return nil, err
		} else {
			if len(v) >= index {
				newChain := byChain[1:]
				data = v[index]
				if len(newChain) == 0 {

					if CanHandleValue(data) {
						return data, nil
					} else {
						return nil, fmt.Errorf("[]interface{} unsupported value type %T", data)
					}
				} else {
					return FindChain(data, newChain...)
				}
			} else {
				return nil, fmt.Errorf("index %v is out of bounds for Array %s in %T", index, byChain[0], data)
			}
		}
	case map[interface{}]interface{}:
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
		} else { // not found an entriy with the key
			return nil, fmt.Errorf("the map do not contains the key %s in %T", byChain[0], data)
		}
	case map[string]interface{}:
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

func CanHandleValue(value any) bool {
	switch value.(type) {
	case int, int16, int32, int64, float32, float64, string, bool, byte, map[string]interface{}:
		return true
	default:
		return false
	}

}
