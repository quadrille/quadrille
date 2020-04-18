package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/http/types"
	"github.com/quadrille/quadrille/opt"
	"github.com/quadrille/quadrille/replication/store"
	"math"
	"strconv"
	"strings"
)

type quadrilleHTTPClient struct {
	host string
}

func New(quadrilleHTTPHost string) opt.QuadrilleService {
	return &quadrilleHTTPClient{host: "http://" + quadrilleHTTPHost}
}

func (q quadrilleHTTPClient) GetLocation(locationID string) (body string, err error) {
	body, _, err = Get(q.host + "/location/" + locationID).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) DeleteLocation(locationID string) (body string, err error) {
	body, _, err = Delete(q.host + "/location/" + locationID).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) Insert(locationID string, location ds.Position, data map[string]interface{}) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"lat": location.Lat(), "lon": location.Long(), "data": data})
	if err != nil {
		return
	}
	body, _, err = Post(q.host + "/location/" + locationID).SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) Update(locationID string, location ds.Position, data map[string]interface{}) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"lat": location.Lat(), "lon": location.Long(), "data": data})
	if err != nil {
		return
	}
	body, _, err = Put(q.host + "/location/" + locationID).SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) UpdateLocation(locationID string, location ds.Position) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"lat": location.Lat(), "lon": location.Long()})
	if err != nil {
		return
	}
	body, _, err = Put(q.host + "/location/" + locationID).SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) UpdateData(locationID string, data map[string]interface{}) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return
	}
	body, _, err = Put(q.host + "/location/" + locationID).SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) BulkWrite(commands []store.Command) (body string, err error) {
	return "", errors.New("operation not supported by client")
}

func (q quadrilleHTTPClient) Neighbors(location ds.Position, radius, limit int) (body string, err error) {
	body, _, err = Get(q.host + "/neighbors").SetQueryParams(
		map[string]string{
			"radius": strconv.Itoa(radius),
			"limit":  strconv.Itoa(limit),
			"lat":    fmt.Sprintf("%f", location.Lat()),
			"lon":    fmt.Sprintf("%f", location.Long()),
		}).SetTimeout(5000).Do()
	if err == nil {
		var results []types.NeighborResult
		parseErr := json.Unmarshal([]byte(body), &results)
		var sb strings.Builder
		if parseErr == nil {
			for i, result := range results {
				dataByte, _ := json.Marshal(result.Data)
				sb.WriteString(fmt.Sprintf("%s %f,%f %.0fm %s", result.LocationID, result.Latitude, result.Longitude, math.Round(result.Distance), string(dataByte)))
				if i != len(results)-1 {
					sb.WriteString("\n")
				}
			}
			body = sb.String()
		}
	}
	if body == "" {
		body = fmt.Sprintf("No match found within %dm of %f,%f", radius, location.Lat(), location.Long())
	}
	return
}

func (q quadrilleHTTPClient) IsLeader() (body string, err error) {
	body, _, err = Get(q.host + "/isleader").SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) Leader() (body string, err error) {
	body, _, err = Get(q.host + "/leader").SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) Members() (body string, err error) {
	body, _, err = Get(q.host + "/members").SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) AddNode(nodeID, addr string) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"addr": addr, "id": nodeID})
	if err != nil {
		return
	}
	body, _, err = Get(q.host + "/join").SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}

func (q quadrilleHTTPClient) RemoveNode(nodeID string) (body string, err error) {
	payload, err := json.Marshal(map[string]interface{}{"id": nodeID})
	if err != nil {
		return
	}
	body, _, err = Get(q.host + "/remove").SetPayload(string(payload)).SetTimeout(5000).Do()
	return
}
