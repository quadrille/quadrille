package client

import (
	"encoding/json"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/replication/store"
	"testing"
)

type QuadrilleMockService struct {
}

func (q QuadrilleMockService) GetLocation(locationID string) (body string, err error) {
	b, err := json.Marshal(map[string]interface{}{
		"lat":  12,
		"long": 77,
	})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (q QuadrilleMockService) DeleteLocation(locationID string) (body string, err error) {
	return "", store.ErrNonExistentLocationDelete
}

func (q QuadrilleMockService) Insert(locationID string, location ds.Position, data map[string]interface{}) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) Update(locationID string, location ds.Position, data map[string]interface{}) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) UpdateLocation(locationID string, location ds.Position) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) UpdateData(locationID string, data map[string]interface{}) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) Neighbors(location ds.Position, radius, limit int) (body string, err error) {
	panic("implement me")
}

func (q QuadrilleMockService) IsLeader() (body string, err error) {
	return "true", nil
}

func (q QuadrilleMockService) Leader() (body string, err error) {
	return ":5677", nil
}

func (q QuadrilleMockService) Members() (body string, err error) {
	return ":5677", nil
}

func (q QuadrilleMockService) AddNode(nodeID, addr string) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) RemoveNode(nodeID string) (body string, err error) {
	return "", nil
}

func (q QuadrilleMockService) BulkWrite(commands []store.Command) (body string, err error) {
	return "", nil
}

var quadrilleMockService = QuadrilleMockService{}

func TestExecutor(t *testing.T) {
	getLocationCmd := "get loc001"
	delLocationCmd := "del"
	insertLocationCmd := "insert loc002"
	neighborsCmd := "neighbors 12,77"
	getLeaderCmd := "leader"
	isLeader := "isleader"

	responseStr, _ := Executor(getLocationCmd, quadrilleMockService)
	expectedResp := `{"lat":12,"long":77}`
	if responseStr != expectedResp {
		t.Fatalf("Expected: %s, got: %s", expectedResp, responseStr)
	}

	_, err := Executor(delLocationCmd, quadrilleMockService)
	expectedErrTxt := "del needs a location_id"
	if err == nil || err.Error() != expectedErrTxt {
		t.Fatalf("Expected: %s, got: %s", expectedErrTxt, err)
	}

	_, err = Executor(insertLocationCmd, quadrilleMockService)
	expectedErrTxt = "operation needs a location_id and lat,long"
	if err == nil || err.Error() != expectedErrTxt {
		t.Fatalf("Expected: %s, got: %s", expectedErrTxt, err)
	}

	_, err = Executor(neighborsCmd, quadrilleMockService)
	expectedErrTxt = "neighbors needs a radius"
	if err == nil || err.Error() != expectedErrTxt {
		t.Fatalf("Expected: %s, got: %s", expectedErrTxt, err)
	}

	responseStr, err = Executor(getLeaderCmd, quadrilleMockService)
	expectedResp = ":5677"
	if responseStr != expectedResp {
		t.Fatalf("Expected: %s, got: %s", expectedResp, err)
	}

	responseStr, err = Executor(isLeader, quadrilleMockService)
	expectedResp = "true"
	if responseStr != expectedResp {
		t.Fatalf("Expected: %s, got: %s", expectedResp, err)
	}
}
