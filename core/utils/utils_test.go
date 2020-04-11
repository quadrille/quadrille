package utils

import (
	"testing"
)

func TestDistanceOnEarth(t *testing.T) {
	expected := 4047
	distance := DistanceOnEarth(12.9660637, 77.7157481, 12.9958069, 77.6942081)
	//fmt.Println(distance)
	if int(distance) != expected {
		t.Errorf("Expected %d, Got %d", expected, int(distance))
	}
}

func TestEuclideanDistance(t *testing.T) {
	expected := 0.036723691892838015
	distance := EuclideanDistance(12.9660637, 77.7157481, 12.9958069, 77.6942081)
	if distance != expected {
		t.Errorf("Expected %f, Got %f", expected, distance)
	}
}
