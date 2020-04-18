package tcp

import (
	"bufio"
	"fmt"
	"github.com/quadrille/quadrille/http/client"
	"github.com/quadrille/quadrille/opt"
	"github.com/quadrille/quadrille/replication/store"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type service struct {
	addr  string
	store store.Store
}

func New(addr string, store store.Store) *service {
	return &service{addr, store}
}

func (srv *service) Start() error {
	s, err := net.ResolveTCPAddr("tcp", srv.addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", s)
	if err != nil {
		return err
	}

	quadrilleTCPService := NewQuadrilleService(srv.store)

	go func() {
		defer ln.Close()
		for {
			c, err := ln.Accept()
			if err != nil {
				log.Println(err)
				os.Exit(100)
			}
			go handleConnection(c, quadrilleTCPService)
		}
	}()
	return nil
}

func handleConnection(c net.Conn, service opt.QuadrilleService) {
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection from %s closed\n", c.RemoteAddr())
			} else {
				log.Println("Error reading from connection", err)
			}
			break
		}
		//fmt.Println(netData)
		cmdLine := strings.TrimSpace(string(netData))
		cmdParts := strings.Split(cmdLine, "::")
		if len(cmdParts) < 2 {
			c.Write([]byte(fmt.Sprintf("%s::ERROR:%s\n", cmdParts[0], "quadrille protocol expects format queryid::command")))
		} else {
			go func() {
				//	fmt.Println(cmdParts[0])
				//c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], "{}")))
				////log.Println("Calling executor")
				respBody, err := client.Executor(cmdParts[1], service)
				//log.Println("Got executor response", respBody, err)
				if err != nil {
					c.Write([]byte(fmt.Sprintf("%s::ERROR:%s\n", cmdParts[0], err.Error())))
				} else {
					c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], respBody)))
				}
			}()
		}
	}
	time.Sleep(5 * time.Second)
	defer c.Close()
}
