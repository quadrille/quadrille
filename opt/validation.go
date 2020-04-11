package opt

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type cmdValidator func([]string) error

var validatorMap = map[string]cmdValidator{
}

func init() {
	validatorMap[GetLocation] = validateGet
	validatorMap[Insert] = validateInsertOrUpdate
	validatorMap[Update] = validateInsertOrUpdate
	validatorMap[UpdateLocation] = validateUpdateLocation
	validatorMap[UpdateData] = validateUpdateData
	validatorMap[DeleteLocation] = validateDel
	validatorMap[Neighbors] = validateNeighbors
	validatorMap[Join] = validateAddNode
}

func validateDel(cmdParts []string) error {
	if len(cmdParts) < 2 {
		return errors.New("del needs a location_id")
	}
	return nil
}

func validateAddNode(cmdParts []string) error {
	if len(cmdParts) < 2 {
		return errors.New("addnode needs address of the node to be added. Example `addnode quadhost2:5677 node2`")
	}
	//var re = regexp.MustCompile(`(?m)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):\d{1,5}`)
	//var str = cmdParts[1]
	//isValidAddr := false
	//for _, _ = range re.FindAllString(str, -1) {
	//	isValidAddr = true
	//}
	//if !isValidAddr {
	//	return errors.New("please enter a valid address to add. Format `addnode IP:port`")
	//}

	if len(cmdParts) < 3 {
		return errors.New("addnode needs node name")
	}
	return nil
}

func isValidCoords(coords string) bool {
	latLong := strings.Split(coords, ",")
	if len(latLong) != 2 {
		return false
	}
	latStr, longStr := latLong[0], latLong[1]
	lat, latErr := strconv.ParseFloat(latStr, 64)
	long, longErr := strconv.ParseFloat(longStr, 64)
	return !(latErr != nil || longErr != nil || lat < -90 || lat > 90 || long < -180 || long > 180)
}

func validateNeighbors(cmdParts []string) error {
	if len(cmdParts) < 2 {
		return errors.New("neighbors needs a lat,lon")
	}
	if !isValidCoords(cmdParts[1]) {
		return InvalidLatLon
	}

	if len(cmdParts) < 3 {
		return errors.New("neighbors needs a radius")
	}

	radius, err := strconv.Atoi(cmdParts[2])
	if err != nil || radius == 0 {
		return errors.New("radius should be a positive integer")
	}
	return nil
}

func validateInsertOrUpdate(cmdParts []string) error {
	if len(cmdParts) < 3 {
		return errors.New("operation needs a location_id and lat,long")
	}
	if !isValidCoords(cmdParts[2]) {
		return InvalidLatLon
	}
	if len(cmdParts) >= 4 && !isDataValid(cmdParts[3]) {
		return InvalidData
	}
	return nil
}

func validateUpdateLocation(cmdParts []string) error {
	if len(cmdParts) < 3 {
		return errors.New("updateloc needs a location_id and lat,long")
	}
	if !isValidCoords(cmdParts[2]) {
		return InvalidLatLon
	}
	return nil
}

func validateUpdateData(cmdParts []string) error {
	if len(cmdParts) < 3 {
		return errors.New("updateloc needs a location_id and lat,long")
	}
	if !isDataValid(cmdParts[2]) {
		return InvalidData
	}
	return nil
}

func isDataValid(dataStr string) bool {
	var dataMap map[string]interface{}
	err := json.Unmarshal([]byte(dataStr), &dataMap)
	if err != nil {
		return false
	}
	return true
}

func validateGet(cmdParts []string) error {
	if len(cmdParts) < 2 {
		return errors.New("get needs a location_id ")
	}
	return nil
}

func NewValidator(cmd string) cmdValidator {
	validator, ok := validatorMap[cmd]
	if !ok {
		return func(parts []string) error {
			return nil
		}
	}
	return validator
}
