package ds

type Quadrille interface {
	Insert(string, Position, map[string]interface{})
	Delete(string) error
	Update(string, Position, map[string]interface{}) error
	UpdateLocation(string, Position) error
	UpdateData(string, map[string]interface{}) error
	GetNearbyLocations(Position, int, int) []QuadTreeNeighborResult
	Get(string) (QuadTreeLeaf, error)
	GetAllLocations() QuadTreeSnapshot
}
