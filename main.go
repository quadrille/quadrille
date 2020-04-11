package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/quadrille/quadrille/constants"
	httpd "github.com/quadrille/quadrille/http"
	"github.com/quadrille/quadrille/replication/store"
	"github.com/quadrille/quadrille/tcp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

// Command line defaults

// Command line parameters
var httpPort string
var raftPort string
var tcpPort string
var joinAddr string
var nodeID string
var bindToHost bool
var bindIP string
var dbPath string

func init() {
	flag.StringVar(&httpPort, "h", constants.DefaultHTTPPort, "Set the HTTP port")
	flag.StringVar(&raftPort, "r", constants.DefaultRaftPort, "Set Raft port")
	flag.StringVar(&tcpPort, "t", constants.DefaultTCPPort, "Set TCP port")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID")
	flag.BoolVar(&bindToHost, "bindToHost", false, "Bind to Hostname")
	flag.StringVar(&bindIP, "bindIP", "", "Bind IP")
	flag.StringVar(&dbPath, "dbPath", "", "DB Data Path")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

const logo = `
   ____                  _      _ _ _      _____  ____  
  / __ \                | |    (_) | |    |  __ \|  _ \ 
 | |  | |_   _  __ _  __| |_ __ _| | | ___| |  | | |_) | The distributed, fault-tolerant
 | |  | | | | |/ _' |/ _' | '__| | | |/ _ \ |  | |  _ <  Quadtree database
 | |__| | |_| | (_| | (_| | |  | | | |  __/ |__| | |_) |
  \___\_\\__,_|\__,_|\__,_|_|  |_|_|_|\___|_____/|____/ 
                                                        
                                                        
`

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()

	nodeID := ensureAndGetNodeID(nodeID, joinAddr)
	raftDir := prepareDataDir(dbPath, nodeID)
	hostname := getHostName(bindToHost, bindIP)

	httpAddr, raftAddr, tcpAddr := getListenerAddresses(hostname, httpPort, raftPort, tcpPort)
	s := prepareAndOpenRaftStore(raftDir, raftAddr, nodeID)
	startHTTPListener(httpAddr, s)
	startTCPListener(tcpAddr, s)

	//If join was specified, make the join request.
	if joinAddr != "" {
		joinLeader(raftAddr, nodeID)
	}

	printStartMessages(logo, raftDir)
	waitForCtrlC()
	log.Println("Quadrille exiting")
}

func printStartMessages(logo, dataDir string) {
	log.Println(logo)
	log.Println("Quadrille started successfully")
	log.Printf("Data Dir: %s\n", dataDir)
}

func joinLeader(raftAddr string, nodeID string) {
	if err := join(joinAddr, raftAddr, nodeID); err != nil {
		log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
	}
}

func startTCPListener(tcpAddr string, s store.Store) {
	//Start TCP listener
	tcpServer := tcp.New(tcpAddr, s)
	if err := tcpServer.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}
}

func startHTTPListener(httpAddr string, s store.Store) {
	fmt.Printf("Starting http service at %s\n", httpAddr)
	//Start HTTP listener
	h := httpd.New(httpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}
}

func prepareAndOpenRaftStore(raftDir string, raftAddr string, nodeID string) store.Store {
	s := store.New(raftDir, raftAddr)
	//Open Raft storage
	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}
	return s
}

func getListenerAddresses(hostname, httpPort, raftPort, tcpPort string) (httpAddr, raftAddr, tcpAddr string) {
	httpAddr = hostname + ":" + httpPort
	raftAddr = hostname + ":" + raftPort
	tcpAddr = hostname + ":" + tcpPort
	raftAddr = strings.Replace(raftAddr, "0.0.0.0", "localhost", -1)
	return
}

func waitForCtrlC() {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
}

func ensureAndGetNodeID(nodeID, joinAddr string) string {
	if joinAddr != "" && nodeID == "" {
		log.Fatalln("All cluster members must have a unique NodeID. Please specify an NodeID using -id flag")
	}
	if nodeID == "" {
		nodeID = "node0"
	}
	return nodeID
}

func getHostName(bindToHost bool, bindIP string) string {
	hostname := "0.0.0.0"
	if bindToHost {
		host, err := os.Hostname()
		if err != nil {
			log.Fatalln("failed to bind to host")
		}
		hostname = host
	} else if bindIP != "" {
		hostname = bindIP
	}
	return hostname
}

func prepareDataDir(raftDir, nodeID string) string {
	//Ensure Raft storage exists.
	if raftDir == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln("error getting user's home directory")
		}
		raftDir = fmt.Sprintf("%s/quadrille-data-%s", userHomeDir, nodeID)
	}
	os.MkdirAll(raftDir, 0700)
	return raftDir
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
