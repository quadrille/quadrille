package ds

import (
	quadrilleError "github.com/quadrille/quadrille/core/errors"
	"sort"
	"sync"
)

type QuadTreeNode struct {
	boundingBox Rectangle         //Bounds of the current node
	children    *[4]*QuadTreeNode //Quadrants of the current node.  This is lazily initialized to conserve memory
	once        sync.Once         //Mutex to ensure single init of children
	isLeaf      bool
	leaves      *map[string]*QuadTreeLeaf //Leaves which contain the actual LocationID
	leavesMtx   sync.RWMutex              //Mutex to synchronize the leaves
	parent      *QuadTreeNode
}

type QuadTreeNeighborResult struct {
	Leaf     QuadTreeLeaf `json:"Leaf"`
	Distance float64      `json:"Distance"`
}

type byDistance []QuadTreeNeighborResult

func (d byDistance) Len() int {
	return len(d)
}

func (d byDistance) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d byDistance) Less(i, j int) bool {
	return d[i].Distance < d[j].Distance
}

func NewQuadTreeNeighborResult(leaf QuadTreeLeaf, distance float64) *QuadTreeNeighborResult {
	return &QuadTreeNeighborResult{Leaf: leaf, Distance: distance}
}

func NewQuadTreeNode(boundingBox Rectangle, children *[4]*QuadTreeNode, isLeaf bool, leaves *map[string]*QuadTreeLeaf, parent *QuadTreeNode) *QuadTreeNode {
	return &QuadTreeNode{boundingBox: boundingBox, children: children, isLeaf: isLeaf, leaves: leaves, parent: parent}
}

func (q *QuadTreeNode) findContainingChild(location GeoLocation) *QuadTreeNode {
	initChildren(q)
	for _, child := range q.children {
		box := child.boundingBox
		if isWithinBox(box, location) {
			return child
		}
	}
	return nil
}

//This function is used to lazily initialize children as it is only needed to be set for non-empty nodes
func initChildren(q *QuadTreeNode) {
	//The function passed to the below sync.Once is only executed once for the node q
	q.once.Do(func() {
		//Code to initialize children of node q
		var children [4]*QuadTreeNode
		sections := q.boundingBox.GetQuadrants()
		for i := 0; i < 4; i++ {
			children[i] = NewQuadTreeNode(sections[i], nil, false, nil, q)
		}
		q.children = &children
	})
}

func isWithinBox(box Rectangle, location GeoLocation) bool {
	minLong, minLat, maxLong, maxLat := box.Corner1().Long(), box.Corner1().Lat(), box.Corner2().Long(), box.Corner2().Lat()
	if minLong > maxLong {
		minLong, maxLong = maxLong, minLong
	}
	if minLat > maxLat {
		minLat, maxLat = maxLat, minLat
	}
	if location.Lat() >= minLat && location.Lat() <= maxLat && location.Long() >= minLong && location.Long() <= maxLong {
		return true
	}
	return false
}

type QuadTreeLeaf struct {
	Location   Position               `json:"location"`
	LocationID string                 `json:"locationID"`
	Data       map[string]interface{} `json:"data"`
}

func NewQuadTreeLeaf(location Position, locationID string, data map[string]interface{}) *QuadTreeLeaf {
	return &QuadTreeLeaf{Location: location, LocationID: locationID, Data: data}
}

func (q QuadTreeLeaf) GetLocation() Position {
	return q.Location
}

func (q QuadTreeLeaf) GetLocationID() string {
	return q.LocationID
}

type QuadTree struct {
	root          *QuadTreeNode
	height        int
	locationIndex *concurrentMap
}

func NewQuadTree(height int) *QuadTree {
	return &QuadTree{height: height,
		root: NewQuadTreeNode(
			NewRectangle(NewPosition(90, -180), NewPosition(-90, 180)),
			nil,
			false,
			nil,
			nil),
		locationIndex: NewMap(),
	}
}

func (q *QuadTree) Insert(locationID string, location Position, data map[string]interface{}) {
	q.insert(locationID, location, data, true)
}

func (q *QuadTree) insert(locationID string, location Position, data map[string]interface{}, safe bool) *QuadTreeNode {
	var insert func(int, *QuadTreeNode) *QuadTreeNode
	insert = func(depth int, cur *QuadTreeNode) *QuadTreeNode {
		if depth <= q.height {
			// We haven't reached target depth yet.
			// Keep searching child nodes recursively.
			return insert(depth+1, cur.findContainingChild(location))
		} else {
			//We are in Leaf. Add Location
			if safe {
				q.locationIndex.Lock(locationID)
				defer q.locationIndex.UnLock(locationID)
				cur.leavesMtx.Lock()
				defer cur.leavesMtx.Unlock()
			}
			if cur.leaves == nil {
				cur.leaves = &map[string]*QuadTreeLeaf{}
			}
			//*cur.leaves = append(*cur.leaves, *NewQuadTreeLeaf(Location, LocationID, ShardKey))
			(*cur.leaves)[locationID] = NewQuadTreeLeaf(location, locationID, data)
			q.locationIndex.SetUnsafe(locationID, cur)
			return cur
		}
	}
	node := insert(1, q.root)
	return node
}

func (q *QuadTree) Delete(locationID string) error {
	node := q.locationIndex.Get(locationID)
	if node == nil {
		return quadrilleError.NonExistingLocationDeleteAttempt
	}
	q.locationIndex.Lock(locationID)
	defer q.locationIndex.UnLock(locationID)
	node.leavesMtx.Lock()
	defer node.leavesMtx.Unlock()
	//Checking again as this might have changed due to concurrent code
	if node = q.locationIndex.GetUnsafe(locationID); node == nil {
		return quadrilleError.NonExistingLocationDeleteAttempt
	}

	delete(*node.leaves, locationID)
	q.locationIndex.DeleteUnsafe(locationID)
	return nil
}

func (q *QuadTree) UpdateLocation(locationID string, location Position) error {
	node := q.locationIndex.Get(locationID)
	if node == nil {
		return quadrilleError.NonExistingLocationUpdateAttempt
	}
	q.locationIndex.Lock(locationID)
	defer q.locationIndex.UnLock(locationID)
	node.leavesMtx.Lock()
	defer node.leavesMtx.Unlock()
	//Checking again as this might have changed due to concurrent code
	if node = q.locationIndex.GetUnsafe(locationID); node == nil {
		return quadrilleError.NonExistingLocationUpdateAttempt
	}
	if isWithinBox(node.boundingBox, location) {
		(*node.leaves)[locationID].Location = location
	} else {
		data := (*node.leaves)[locationID].Data
		delete(*node.leaves, locationID)
		q.locationIndex.DeleteUnsafe(locationID)
		q.insert(locationID, location, data, false)
	}
	return nil
}

func (q *QuadTree) UpdateData(locationID string, data map[string]interface{}) error {
	node := q.locationIndex.Get(locationID)
	if node == nil {
		return quadrilleError.NonExistingLocationUpdateAttempt
	}
	q.locationIndex.Lock(locationID)
	defer q.locationIndex.UnLock(locationID)
	node.leavesMtx.Lock()
	defer node.leavesMtx.Unlock()
	//Checking again as this might have changed due to concurrent code
	if node = q.locationIndex.GetUnsafe(locationID); node == nil {
		return quadrilleError.NonExistingLocationUpdateAttempt
	}
	(*node.leaves)[locationID].Data = data
	return nil
}

func (q *QuadTree) Update(locationID string, location Position, data map[string]interface{}) error {
	_, err := q.update(locationID, location, data)
	return err
}

func (q *QuadTree) update(locationID string, location Position, data map[string]interface{}) (*QuadTreeNode, error) {
	node := q.locationIndex.Get(locationID)
	if node == nil {
		return nil, quadrilleError.NonExistingLocationUpdateAttempt
	}
	q.locationIndex.Lock(locationID)
	defer q.locationIndex.UnLock(locationID)
	node.leavesMtx.Lock()
	defer node.leavesMtx.Unlock()
	//Checking again as this might have changed due to concurrent code
	if node = q.locationIndex.GetUnsafe(locationID); node == nil {
		return nil, quadrilleError.NonExistingLocationUpdateAttempt
	}
	updatedNode := node
	if isWithinBox(node.boundingBox, location) {
		(*node.leaves)[locationID].Location = location
		(*node.leaves)[locationID].Data = data
	} else {
		delete(*node.leaves, locationID)
		q.locationIndex.DeleteUnsafe(locationID)
		updatedNode = q.insert(locationID, location, data, false)
	}
	return updatedNode, nil
}

func (q *QuadTree) Get(locationID string) (QuadTreeLeaf, error) {
	if node := q.locationIndex.Get(locationID); node == nil {
		return QuadTreeLeaf{}, quadrilleError.LocationNotFound
	}
	node := q.locationIndex.Get(locationID)
	node.leavesMtx.RLock()
	defer node.leavesMtx.RUnlock()
	leaves := node.leaves
	return *(*leaves)[locationID], nil
}

func filterLeafsByDistance(leaves map[string]*QuadTreeLeaf, location GeoLocation, distanceInMetres int) []QuadTreeNeighborResult {
	filteredLeaves := []QuadTreeNeighborResult{}
	for _, leaf := range leaves {
		distance := location.DistanceTo(leaf.GetLocation())
		if distance <= float64(distanceInMetres) {
			filteredLeaves = append(filteredLeaves, *NewQuadTreeNeighborResult(*leaf, distance))
		}
	}
	return filteredLeaves
}

func getNearbyChildLeaves(q QuadTreeNode, location GeoLocation, radiusInMetres int) []QuadTreeNeighborResult {
	leaves := []QuadTreeNeighborResult{}
	var addMatchingLeaves func(node QuadTreeNode)
	addMatchingLeaves = func(node QuadTreeNode) {
		if node.leaves != nil {
			leaves = append(leaves, filterLeafsByDistance(*node.leaves, location, radiusInMetres)...)
		} else if node.children != nil {
			for _, child := range node.children {
				if location.IntersectsRectangle(child.boundingBox, radiusInMetres) {
					addMatchingLeaves(*child)
				}
			}
		}
	}
	addMatchingLeaves(q)
	return leaves
}

func (q *QuadTreeNode) findNeighbourQuadMatches(location GeoLocation, radiusInMetres int) []QuadTreeNeighborResult {
	matchedLeaves := []QuadTreeNeighborResult{}
	prevNode, curNode := q, q.parent
	for true {
		//If not reached root
		if curNode != nil {
			childsExplored := 0
			for _, child := range curNode.children {
				if *child != *prevNode && location.IntersectsRectangle(child.boundingBox, radiusInMetres) {
					leaves := getNearbyChildLeaves(*child, location, radiusInMetres)
					if len(leaves) > 0 {
						matchedLeaves = append(matchedLeaves, leaves...)
					}
					childsExplored++
				}
			}
			if childsExplored == 0 {
				break
			}
		} else {
			break
		}
		prevNode = curNode
		curNode = curNode.parent
	}
	return matchedLeaves
}

func (q *QuadTree) GetNearbyLocations(location Position, radiusInMetres, limit int) []QuadTreeNeighborResult {
	matchedLeaves := []QuadTreeNeighborResult{}
	if q.root != nil {
		curNode := q.root
		for curNode.children != nil {
			curNode = curNode.findContainingChild(location)
		}
		if curNode.leaves != nil {
			matchedLeaves = append(matchedLeaves, filterLeafsByDistance(*curNode.leaves, location, radiusInMetres)...)
		}
		matchedLeaves = append(matchedLeaves, curNode.findNeighbourQuadMatches(location, radiusInMetres)...)
	}
	sort.Sort(byDistance(matchedLeaves))
	if len(matchedLeaves) > limit {
		return matchedLeaves[:limit]
	}
	return matchedLeaves
}

type QuadTreeSnapshot map[string]QuadTreeLeaf

func (q QuadTree) GetAllLocations() QuadTreeSnapshot {
	var snapShot = make(map[string]QuadTreeLeaf)
	for locationID, qNode := range q.locationIndex.GetAllKeyVal() {
		snapShot[locationID] = *(*qNode.leaves)[locationID]
	}
	return snapShot
}
