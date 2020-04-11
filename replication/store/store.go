package store

import (
	"encoding/json"
	"fmt"
	raftBadger "github.com/bbva/raft-badger"
	"github.com/hashicorp/raft"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/tcp/utils"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"time"
)

const quadTreeHeight = 16

type OperationType string

const (
	OperationInsert         OperationType = "insert"
	OperationDelete                       = "delete"
	OperationUpdate                       = "update"
	OperationUpdateLocation               = "updateloc"
	OperationUpdateData                   = "updatedata"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op         string                 `json:"op,omitempty"`
	LocationID string                 `json:"key,omitempty"`
	Lat        float64                `json:"lat,omitempty"`
	Long       float64                `json:"long,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// Store is the interface Raft-backed key-value stores must implement.
type Store interface {
	Open(enableSingle bool, localID string) error

	// Get returns the value for the given key.
	Get(key string) (ds.QuadTreeLeaf, error)

	// Set sets the value for the given key, via distributed consensus.
	Insert(locationID string, position ds.GeoLocation, data map[string]interface{}) error

	Update(locationID string, position ds.GeoLocation, data map[string]interface{}) error

	UpdateLocation(locationID string, position ds.GeoLocation) error

	UpdateData(locationID string, data map[string]interface{}) error

	// Delete removes the given key, via distributed consensus.
	Delete(key string) error

	// Join joins the node, identitifed by nodeID and reachable at addr, to the cluster.
	Join(nodeID string, addr string) error
	GetLeader() raft.ServerAddress
	GetNeighbors(ds.Position, int, int) []ds.QuadTreeNeighborResult
	Remove(nodeId string) error
	Nodes() ([]*Server, error)
	IsLeader() bool
}

type store struct {
	raftDir  string
	raftBind string

	q ds.Quadrille // The core data structure for Quadrille. As it is concurrency-safe, it is not required to synchronize the operations

	raft   *raft.Raft // The consensus mechanism
	logger *log.Logger
}

// New returns a new Store.
func New(raftDir, raftBind string) Store {
	return &store{
		q:        ds.NewQuadTree(quadTreeHeight),
		logger:   log.New(os.Stderr, "[store] ", log.LstdFlags),
		raftDir:  raftDir,
		raftBind: raftBind,
	}
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (s *store) Open(enableSingle bool, localID string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", s.raftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.raftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	badgerStore, err := raftBadger.New(raftBadger.Options{
		Path: s.raftDir,
	})

	logStore := badgerStore
	stableStore := badgerStore

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

// Get returns the data for the given location_id.
func (s *store) Get(locationID string) (ds.QuadTreeLeaf, error) {
	return s.q.Get(locationID)
}

//Returns nearby locations.
func (s *store) GetNeighbors(position ds.Position, radius, limit int) []ds.QuadTreeNeighborResult {
	return s.q.GetNearbyLocations(position, radius, limit)
}

// Set sets the data for the given location_id.
func (s *store) Insert(locationID string, location ds.GeoLocation, data map[string]interface{}) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	//log.Println("Inside Set")
	c := &command{
		Op:         string(OperationInsert),
		LocationID: locationID,
		Lat:        location.Lat(),
		Long:       location.Long(),
		Data:       data,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (s *store) Update(locationID string, location ds.GeoLocation, data map[string]interface{}) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	//log.Println("Inside Set")
	c := &command{
		Op:         string(OperationUpdate),
		LocationID: locationID,
		Lat:        location.Lat(),
		Long:       location.Long(),
		Data:       data,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (s *store) UpdateLocation(locationID string, location ds.GeoLocation) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	//log.Println("Inside Set")
	c := &command{
		Op:         string(OperationUpdateLocation),
		LocationID: locationID,
		Lat:        location.Lat(),
		Long:       location.Long(),
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (s *store) UpdateData(locationID string, data map[string]interface{}) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	//log.Println("Inside Set")
	c := &command{
		Op:         string(OperationUpdateData),
		LocationID: locationID,
		Data:       data,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Delete deletes the given location.
func (s *store) Delete(locationID string) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}

	if _, err := s.q.Get(locationID); err != nil {
		return NonExistentLocationDeleteError
	}
	c := &command{
		Op:         string(OperationDelete),
		LocationID: locationID,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *store) Join(nodeID, addr string) error {
	s.logger.Printf("received join request for remote node %s at %s", nodeID, addr)
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	isAvailable, err := utils.IsServiceAvailable(addr)
	if err != nil {
		return err
	}
	if !isAvailable {
		return AddressNotReachableError
	}
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's shardID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the shardID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				s.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	s.logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

// Remove removes a node from the store, specified by shardID.
func (s *store) Remove(id string) error {
	s.logger.Printf("received request to remove node %s", id)
	if err := s.remove(id); err != nil {
		s.logger.Printf("failed to remove node %s: %s", id, err.Error())
		return err
	}

	s.logger.Printf("node %s removed successfully", id)
	return nil
}

// remove removes the node, with the given shardID, from the cluster.
func (s *store) remove(id string) error {
	if s.raft.State() != raft.Leader {
		return NonLeaderNodeError
	}
	future := s.raft.RemoveServer(raft.ServerID(id), 0, 0)
	if err := future.Error(); err != nil {
		return fmt.Errorf("error removing existing node %s: %s", id, err)
	}
	return nil
}

// GetLeader returns the address of the cluster leader
func (s *store) GetLeader() raft.ServerAddress {
	return s.raft.Leader()
}

// IsLeader is used to determine if the current node is cluster leader
func (s *store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// Nodes returns the slice of nodes in the cluster, sorted by shardID ascending.
func (s *store) Nodes() ([]*Server, error) {
	f := s.raft.GetConfiguration()
	if f.Error() != nil {
		return nil, f.Error()
	}

	rs := f.Configuration().Servers
	servers := make([]*Server, len(rs))
	for i := range rs {
		servers[i] = &Server{
			ID:   string(rs[i].ID),
			Addr: string(rs[i].Address),
		}
	}

	sort.Sort(Servers(servers))
	return servers, nil
}

type fsmGenericResponse struct {
	error error
}

type fsm store

// Apply applies a Raft log entry to the Quadrille store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch OperationType(c.Op) {
	case OperationInsert:
		return f.applyInsert(c.LocationID, *ds.NewPosition(c.Lat, c.Long), c.Data)
	case OperationDelete:
		return f.applyDelete(c.LocationID)
	case OperationUpdate:
		return f.applyUpdate(c.LocationID, *ds.NewPosition(c.Lat, c.Long), c.Data)
	case OperationUpdateLocation:
		return f.applyUpdateLocation(c.LocationID, *ds.NewPosition(c.Lat, c.Long))
	case OperationUpdateData:
		return f.applyUpdateData(c.LocationID, c.Data)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
	return nil
}

// Snapshot returns a snapshot of the Quadrille store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	// Clone the map.
	o := make(map[string]ds.QuadTreeLeaf)
	for k, v := range f.q.GetAllLocations() {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

// Restore stores the Quadrille store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	o := make(map[string]ds.QuadTreeLeaf)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		fmt.Println(err)
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	qTmp := ds.NewQuadTree(16)
	for locationID, leaf := range o {
		qTmp.Insert(locationID, leaf.GetLocation(), leaf.Data)
	}
	f.q = qTmp
	return nil
}

func (f *fsm) applyInsert(locationId string, location ds.Position, data map[string]interface{}) interface{} {
	f.q.Insert(locationId, location, data)
	return nil
}

func (f *fsm) applyDelete(key string) error {
	return f.q.Delete(key)
}

func (f *fsm) applyUpdate(locationId string, location ds.Position, data map[string]interface{}) error {
	return f.q.Update(locationId, location, data)
}

func (f *fsm) applyUpdateLocation(locationId string, location ds.Position) error {
	return f.q.UpdateLocation(locationId, location)
}

func (f *fsm) applyUpdateData(locationId string, data map[string]interface{}) error {
	return f.q.UpdateData(locationId, data)
}

type fsmSnapshot struct {
	store map[string]ds.QuadTreeLeaf
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
