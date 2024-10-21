package random

import "math/rand/v2"

func RandomF64(minValue, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}
