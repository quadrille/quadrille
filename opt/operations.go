package opt

import (
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/replication/store"
)

type OperationType string

const (
	GetLocation       = "get"
	DeleteLocation    = "del"
	ReplicaSetMembers = "members"
	IsLeader          = "isleader"
	Leader            = "leader"
	Insert            = "insert"
	Update            = "update"
	UpdateLocation    = "updateloc"
	UpdateData        = "updatedata"
	Join              = "join"
	Remove            = "removenode"
	Neighbors         = "neighbors"
	BulkWrite         = "bulkwrite"
)

type QuadrilleService interface {
	GetLocation(locationID string) (body string, err error)
	DeleteLocation(locationID string) (body string, err error)
	Insert(locationID string, location ds.Position, data map[string]interface{}) (body string, err error)
	Update(locationID string, location ds.Position, data map[string]interface{}) (body string, err error)
	UpdateLocation(locationID string, location ds.Position) (body string, err error)
	UpdateData(locationID string, data map[string]interface{}) (body string, err error)
	Neighbors(location ds.Position, radius, limit int) (body string, err error)
	IsLeader() (body string, err error)
	Leader() (body string, err error)
	Members() (body string, err error)
	AddNode(nodeID, addr string) (body string, err error)
	RemoveNode(nodeID string) (body string, err error)
	BulkWrite(commands []store.Command) (body string, err error)
}
