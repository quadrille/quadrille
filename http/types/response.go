package types

import "github.com/quadrille/quadrille/core/ds"

type NeighborResult struct {
	Latitude   float64
	Longitude  float64
	LocationID string
	Distance   float64
	Data       map[string]interface{}
}

func NewNeighborResult(r ds.QuadTreeNeighborResult) *NeighborResult {
	return &NeighborResult{
		Latitude:   r.Leaf.GetLocation().Lat(),
		Longitude:  r.Leaf.GetLocation().Long(),
		LocationID: r.Leaf.GetLocationID(),
		Distance:   r.Distance,
		Data:       r.Leaf.Data,
	}
}

func PrepareNeighborResults(neighbors []ds.QuadTreeNeighborResult) []NeighborResult {
	results := make([]NeighborResult, 0)
	for _, result := range neighbors {
		results = append(results, *NewNeighborResult(result))
	}
	return results
}
