package utils

import "math"

func toRadians(degree float64) float64 {
	return degree * math.Pi / 180
}

func DistanceOnEarth(lat1, long1, lat2, long2 float64) float64 {
	long1 = toRadians(long1)
	long2 = toRadians(long2)
	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)

	dLong := long2 - long1
	dLat := lat2 - lat1
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLong/2), 2)
	c := 2 * math.Asin(math.Sqrt(a))
	r := 6371.0
	return c * r * 1000
}

func EuclideanDistance(lat1, long1, lat2, long2 float64) float64 {
	a := math.Abs(long2 - long1)
	b := math.Abs(lat2 - lat1)
	return math.Sqrt(a*a + b*b)
}
