package util

import (
	"encoding/json"
	"fmt"
	"math"
)

// 指定小數點位數
func GetFloat64WithDP(v interface{}, dp int) float64 {
	if dp < 0 {
		dp = 0
	}
	val := GetFloat64(v)
	if dp == 0 {
		return val
	}
	return val / math.Pow10(int(dp))
}

func GetFloat64(v interface{}) float64 {
	switch v.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", v))
	case nil:
		return 0.0
	case uint32:
		return float64(v.(uint32))
	case uint16:
		return float64(v.(uint16))
	case int16:
		return float64(v.(int16))
	case int:
		return float64(v.(int))
	case json.Number:
		jn, _ := v.(json.Number)
		result, err := jn.Float64()
		if err != nil {
			return 0.0
		}
		return result
	case float64:
		return v.(float64)
	}
}
