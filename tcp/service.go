package tcp

import (
	"encoding/json"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/opt"
	"github.com/quadrille/quadrille/replication/store"
)

type quadrilleTCPClient struct {
	store store.Store
}

func NewQuadrilleService(store store.Store) opt.QuadrilleService {
	return &quadrilleTCPClient{store: store}
}

func transformResponse(responseObj interface{}, e error) (body string, err error) {
	if e == nil {
		bodyByte, err := json.Marshal(responseObj)
		if err == nil {
			body = string(bodyByte)
		}
	}
	return body, e
}

func getResponseObjectFromQuadtreeLeaf(leaf ds.QuadTreeLeaf) map[string]interface{} {
	return map[string]interface{}{"data": leaf.GetLocationID(), "lat": leaf.GetLocation().Lat(), "lon": leaf.GetLocation().Long()}
}

func (q quadrilleTCPClient) GetLocation(locationID string) (body string, err error) {
	leaf, err := q.store.Get(locationID)
	if err == nil {
		return transformResponse(getResponseObjectFromQuadtreeLeaf(leaf), err)
	}
	return
}

func (q quadrilleTCPClient) DeleteLocation(locationID string) (body string, err error) {
	err = q.store.Delete(locationID)
	return
}

func (q quadrilleTCPClient) Insert(locationID string, position ds.Position, data map[string]interface{}) (body string, err error) {
	err = q.store.Insert(locationID, position, data)
	return
}

func (q quadrilleTCPClient) Update(locationID string, position ds.Position, data map[string]interface{}) (body string, err error) {
	err = q.store.Update(locationID, position, data)
	return
}

func (q quadrilleTCPClient) UpdateLocation(locationID string, position ds.Position) (body string, err error) {
	err = q.store.UpdateLocation(locationID, position)
	return
}

func (q quadrilleTCPClient) UpdateData(locationID string, data map[string]interface{}) (body string, err error) {
	err = q.store.UpdateData(locationID, data)
	return
}

func (q quadrilleTCPClient) Neighbors(location ds.Position, radius, limit int) (body string, err error) {
	neighbors := q.store.GetNeighbors(location, radius, limit)
	neighborsTmp := make([]map[string]interface{}, 0)
	for _, neighbor := range neighbors {
		neighborResponse := getResponseObjectFromQuadtreeLeaf(neighbor.Leaf)
		neighborResponse["distance"] = neighbor.Distance
		neighborsTmp = append(neighborsTmp, neighborResponse)
	}
	return transformResponse(neighborsTmp, nil)
}

func (q quadrilleTCPClient) IsLeader() (body string, err error) {
	return transformResponse(q.store.IsLeader(), nil)
}

func (q quadrilleTCPClient) Leader() (body string, err error) {
	return transformResponse(q.store.GetLeader(), nil)
}

func (q quadrilleTCPClient) Members() (body string, err error) {
	return transformResponse(q.store.Nodes())
}

func (q quadrilleTCPClient) AddNode(nodeID, addr string) (body string, err error) {
	err = q.store.Join(nodeID, addr)
	return
}

func (q quadrilleTCPClient) RemoveNode(nodeID string) (body string, err error) {
	err = q.store.Remove(nodeID)
	return
}
