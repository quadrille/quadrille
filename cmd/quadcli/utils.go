package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
	"runtime/debug"
)

//Patch for go-prompt issue
//https://github.com/c-bata/go-prompt/issues/59#issuecomment-376002177

type Exit int

func exit(_ *prompt.Buffer) {
	panic(Exit(0))
}

func handleExit() {
	switch v := recover().(type) {
	case nil:
		return
	case Exit:
		os.Exit(int(v))
	default:
		fmt.Println(v)
		fmt.Println(string(debug.Stack()))
	}
}