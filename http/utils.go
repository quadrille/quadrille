package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func getFloatParamFromQueryString(queryParamMap url.Values, paramName string) (float64, error) {
	if valArr, ok := queryParamMap[paramName]; ok {
		valStr := valArr[0]
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be a valid float", paramName)
		}
		return val, nil
	}
	return 0, fmt.Errorf("missing %s", paramName)
}

func getIntParamFromQueryString(queryParamMap url.Values, paramName string) (int, error) {
	if valArr, ok := queryParamMap[paramName]; ok {
		valStr := valArr[0]
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return 0, fmt.Errorf("%s must be a valid int", paramName)
		}
		return val, nil
	}
	return 0, fmt.Errorf("missing %s", paramName)
}

func setContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func getFloatAttrFromBody(body map[string]interface{}, attrName string) (float64, error) {
	if val, ok := body[attrName]; ok {
		val, ok := val.(float64)
		if !ok {
			return 0, fmt.Errorf("%s must be a valid float", attrName)
		}
		return val, nil
	}
	return 0, fmt.Errorf("missing %s", attrName)
}

func respondWithErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}