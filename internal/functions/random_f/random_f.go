package random_f

import "math/rand/v2"

func RandomF64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
