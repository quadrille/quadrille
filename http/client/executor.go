package client

import (
	"github.com/quadrille/quadrille/opt"
	"strings"
)

func executeCmd(cmdParts []string, service opt.QuadrilleService) (respBody string, err error) {
	switch opt.OperationType(cmdParts[0]) {
	case opt.GetLocation:
		return service.GetLocation(cmdParts[1])
	case opt.DeleteLocation:
		return service.DeleteLocation(cmdParts[1])
	case opt.Insert:
		return service.Insert(cmdParts[1], *getGeolocationFromCoordsStr(cmdParts[2]), prepareDataFromStr(cmdParts, 3))
	case opt.Update:
		return service.Update(cmdParts[1], *getGeolocationFromCoordsStr(cmdParts[2]), prepareDataFromStr(cmdParts, 3))
	case opt.UpdateLocation:
		return service.UpdateLocation(cmdParts[1], *getGeolocationFromCoordsStr(cmdParts[2]))
	case opt.UpdateData:
		return service.UpdateData(cmdParts[1], prepareDataFromStr(cmdParts, 2))
	case opt.Neighbors:
		return service.Neighbors(prepareNeighborQueryArgs(cmdParts))
	case opt.Join:
		return service.AddNode(cmdParts[1], cmdParts[2])
	case opt.Remove:
		return service.RemoveNode(cmdParts[1])
	case opt.Leader:
		return service.Leader()
	case opt.IsLeader:
		return service.IsLeader()
	case opt.ReplicaSetMembers:
		return service.Members()
	default:
		return "", UnrecognizedCommandError
	}
	return "", nil
}

func Executor(line string, service opt.QuadrilleService) (responseStr string, err error) {
	cmdParts := strings.Split(line, " ")
	validatorFunc := opt.NewValidator(cmdParts[0])
	validationErr := validatorFunc(cmdParts)
	if validationErr != nil {
		return "", validationErr
	}
	return executeCmd(cmdParts, service)
}
