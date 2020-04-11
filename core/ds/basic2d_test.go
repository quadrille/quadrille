package ds

import (
	"testing"
)

func TestBasic2D(t *testing.T) {
	location1 := NewPosition(90, 180)
	location2 := NewPosition(0, 0)

	distanceInMetres := location1.DistanceTo(location2)
	expected := 10007543
	if int(distanceInMetres) != expected {
		t.Fatalf("DistanceTo: Expected:%d, Got:%d", expected, int(distanceInMetres))
	}
	location3 := NewPosition(45, -90)
	location4 := NewPosition(-45, 90)
	rect := NewRectangle(location3, location4)
	nearestCorner := rect.GetNearestCorner(location1)
	expectedLat, expectedLong := 45.0, 90.0
	if nearestCorner.Lat() != expectedLat || nearestCorner.Long() != expectedLong {
		t.Fatalf("GetNearestCorner: Expected %f,%f, Got %f,%f", expectedLat, expectedLong, nearestCorner.Lat(), nearestCorner.Long())
	}

	intersects := location1.IntersectsRectangle(rect, (expected/2)+1)
	if !intersects {
		t.Fatalf("IntersectsRectangle: expected:%v, got:%v", true, intersects)
	}
}
