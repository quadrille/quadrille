package ds

import (
	"github.com/quadrille/quadrille/core/errors"
	"testing"
)

import (
	"reflect"
)

var q = NewQuadTree(16)

func init() {
	q.Insert("loc00001", *NewPosition(12.9660637, 77.7157481), map[string]interface{}{})
	q.Insert("loc00002", *NewPosition(12.9649603, 77.7164898), map[string]interface{}{})
}

func TestQuadTree_Get(t *testing.T) {
	locationID := "loc00001"
	leaf, err := q.Get(locationID)
	if err != nil {
		t.Fatalf("Get(%s) Not Found", locationID)
	}
	if leaf.LocationID != locationID || leaf.Location.Lat() != 12.9660637 || leaf.Location.Long() != 77.7157481 {
		t.Fatalf("Get result unexpected")
	}

	_, err = q.Get("loc1234")

	if err == nil {
		t.Fatalf("Get(loc00001) expecting error but got nil")
	}
	if !reflect.DeepEqual(err, errors.LocationNotFound) {
		t.Fatalf("Expecting LocationNotFound but got other error")
	}
}

func TestQuadTree_GetNearbyLocations(t *testing.T) {
	neighbors := q.GetNearbyLocations(*NewPosition(12.9639716, 77.7120424), 1000, 10)
	expectedNeighborCount := 2
	if len(neighbors) != expectedNeighborCount {
		t.Fatalf("Expected %d neighbors, got %d", expectedNeighborCount, len(neighbors))
	}
}

func TestQuadTree_Delete(t *testing.T) {
	err := q.Delete("loc00001")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err.Error())
	}

	err = q.Delete("loc123")
	if !reflect.DeepEqual(err, errors.NonExistingLocationDeleteAttempt) {
		t.Fatalf("Expected LocationNotFound, got %v", err)
	}
}
