package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/quadrille/quadrille/constants"
	httpClient "github.com/quadrille/quadrille/http/client"
	"github.com/quadrille/quadrille/opt"
	"os"
	"strings"
)

func completer(d prompt.Document) []prompt.Suggest {
	if strings.Contains(d.Text, " ") {
		return []prompt.Suggest{}
	}
	s := []prompt.Suggest{
		{Text: "get", Description: "Retrieves a location by id"},
		{Text: "insert", Description: "Creates a new location"},
		{Text: "update", Description: "Updates an existing location"},
		{Text: "updateloc", Description: "Updates an existing location with new lat,long"},
		{Text: "updatedata", Description: "Updates an existing location with new data"},
		{Text: "del", Description: "Deletes an existing location"},
		{Text: "neighbors", Description: "Get nearby locations"},
		{Text: "members", Description: "Lists all replica members"},
		{Text: "leader", Description: "Displays the leader address"},
		{Text: "isleader", Description: "Returns true if connected instance is a leader. False otherwise"},
		{Text: "join", Description: "Joins an existing cluster"},
		{Text: "removenode", Description: "Removed node from a cluster"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func showStartMsg(service opt.QuadrilleService) {
	resp, err := service.IsLeader()
	if err != nil {
		fmt.Println("Unable to connect to node")
		os.Exit(1)
	}
	fmt.Printf("Quadrille shell version %s\n", version)
	if resp == "true" {
		fmt.Println("You are connected to Leader node")
	} else {
		fmt.Println("You are connected to Non-Leader node")
	}
}

func main() {
	quadrilleService := prepareQuadrilleHttpService()
	showStartMsg(quadrilleService)
	history := make([]string, 0)
	for {
		txt := prompt.Input(
			"> ",
			completer,
			prompt.OptionTitle("quadcli"),
			prompt.OptionHistory(history),
			prompt.OptionAddKeyBind(
				prompt.KeyBind{
					Key: prompt.ControlC,
					Fn: func(buf *prompt.Buffer) {
						fmt.Println("bye")
						os.Exit(0)
					},
				}))
		txt = strings.TrimSpace(txt)
		if txt == "exit" {
			os.Exit(0)
		}
		respStr, err := httpClient.Executor(txt, quadrilleService)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(respStr)
			history = append(history, txt)
		}
	}
}

func prepareQuadrilleHttpService() opt.QuadrilleService {
	quadrilleHTTPHost := fmt.Sprintf("localhost:%s", constants.DefaultHTTPPort)
	if len(os.Args) > 1 {
		quadrilleHTTPHost = os.Args[1]
	}
	quadrilleService := httpClient.New(quadrilleHTTPHost)
	return quadrilleService
}
