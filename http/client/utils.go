package client

import (
	"encoding/json"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/replication/store"
	"strconv"
	"strings"
)

func getGeolocationFromCoordsStr(coords string) *ds.Position {
	latLong := strings.Split(coords, ",")
	latStr, longStr := latLong[0], latLong[1]
	lat, _ := strconv.ParseFloat(latStr, 64)
	long, _ := strconv.ParseFloat(longStr, 64)
	return ds.NewPosition(lat, long)
}

func prepareNeighborQueryArgs(cmdParts []string) (location ds.Position, radius int, limit int) {
	location = *getGeolocationFromCoordsStr(cmdParts[1])
	radius, _ = strconv.Atoi(cmdParts[2])
	if len(cmdParts) > 3 {
		limitTmp, err := strconv.Atoi(cmdParts[3])
		if err == nil {
			limit = limitTmp
			return
		}
	}
	limit = 10
	return
}

func prepareDataFromStr(cmdParts []string, expectedPosition int) (data map[string]interface{}) {
	if len(cmdParts) < expectedPosition+1 {
		return make(map[string]interface{})
	}
	json.Unmarshal([]byte(cmdParts[expectedPosition]), &data)
	return
}

func prepareBulkWriteOpsFromStr(bulkWriteStr string) (bulkWriteOps []store.Command) {
	json.Unmarshal([]byte(bulkWriteStr), &bulkWriteOps)
	return
}
