package http

import (
	"encoding/json"
	"errors"
	"github.com/quadrille/quadrille/core/ds"
	"net/http"
	"strings"
)

func prepareUpdateArgs(r *http.Request) (latExists, lonExists, dataExists bool, locationID string, position *ds.Position, data map[string]interface{}, err error) {
	var leaf map[string]interface{}
	if err = json.NewDecoder(r.Body).Decode(&leaf); err != nil {
		err = InvalidBodyErr
		return
	}
	locationID, err = getLocationID(r)
	if err != nil {
		return
	}

	_, latExists = leaf["lat"]
	_, lonExists = leaf["lon"]

	if latExists && lonExists {
		lat, errTmp := getFloatAttrFromBody(leaf, "lat")
		if errTmp != nil {
			err = errTmp
			return
		}
		lon, errTmp := getFloatAttrFromBody(leaf, "lon")
		if errTmp != nil {
			err = errTmp
			return
		}
		position = ds.NewPosition(lat, lon)
	}
	dataTmp, dataExistsTmp := leaf["data"]
	dataExists = dataExistsTmp
	if dataExistsTmp {
		dataTmp, ok := dataTmp.(map[string]interface{})
		if ok {
			data = dataTmp
		} else {
			err = InvalidDataErr
			return
		}
	}
	return
}

func prepareInsertArgs(r *http.Request) (locationID string, position *ds.Position, data map[string]interface{}, err error) {
	var leaf map[string]interface{}
	if err = json.NewDecoder(r.Body).Decode(&leaf); err != nil {
		err = InvalidBodyErr
		return
	}
	locationID, err = getLocationID(r)
	if err != nil {
		return
	}
	lat, err := getFloatAttrFromBody(leaf, "lat")
	lon, err := getFloatAttrFromBody(leaf, "lon")
	position = ds.NewPosition(lat, lon)
	dataTmp, exists := leaf["data"]
	if exists {
		dataTmp, ok := dataTmp.(map[string]interface{})
		if ok {
			data = dataTmp
		} else {
			err = InvalidDataErr
			return
		}
	}
	return
}

func prepareGetNeighborsArg(r *http.Request) (lat, lon float64, radius, limit int, err error) {
	queryParamMap := r.URL.Query()
	lat, err = getFloatParamFromQueryString(queryParamMap, "lat")
	if err != nil {
		return
	}
	lon, err = getFloatParamFromQueryString(queryParamMap, "lon")
	if err != nil {
		return
	}
	radius, err = getIntParamFromQueryString(queryParamMap, "radius")
	if err != nil {
		return
	}
	limit, err = getIntParamFromQueryString(queryParamMap, "limit")
	if err != nil {
		err = nil
		limit = 10
	}
	return
}

func prepareJoinArgs(r *http.Request) (nodeID, remoteAddr string, err error) {
	m := map[string]string{}
	if err = json.NewDecoder(r.Body).Decode(&m); err != nil {
		err = InvalidBodyErr
		return
	}

	nodeIDTmp, ok := m["id"]
	if !ok {
		err = errors.New("missing id")
		return
	}
	nodeID = nodeIDTmp
	
	remoteAddr, ok = m["addr"]
	if !ok {
		err = errors.New("missing addr")
		return
	}
	return
}

func getLocationID(r *http.Request) (string, error) {
	urlParts := strings.Split(r.URL.Path, "/")
	if len(urlParts) < 3 || urlParts[2] == "" {
		return "", errors.New("location_id expected in URL")
	}
	locationID := strings.TrimSpace(urlParts[2])
	return locationID, nil
}
