package ds

import "github.com/quadrille/quadrille/core/utils"

func EuclideanDistance(location1, location2 GeoLocation) float64 {
	return utils.EuclideanDistance(location1.Lat(), location1.Long(), location2.Lat(), location2.Long())
}

func DistanceOnEarth(location1, location2 GeoLocation) float64 {
	return utils.DistanceOnEarth(location1.Lat(), location1.Long(), location2.Lat(), location2.Long())
}
