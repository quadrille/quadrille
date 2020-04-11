// Package httpd provides the HTTP server for accessing the distributed QuadTree store.
// It also provides the endpoint for other nodes to join an existing cluster.
package http

import (
	"encoding/json"
	"errors"
	"github.com/quadrille/quadrille/core/ds"
	"github.com/quadrille/quadrille/http/types"
	"github.com/quadrille/quadrille/replication/store"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Service provides HTTP service.
type Service struct {
	addr  string
	ln    net.Listener
	store store.Store
}

// New returns an uninitialized HTTP service.
func New(addr string, store store.Store) *Service {
	return &Service{
		addr:  addr,
		store: store,
	}
}

// Start starts the service.
func (s *Service) Start() error {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	http.Handle("/", s)

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

// Close closes the service.
func (s *Service) Close() {
	s.ln.Close()
	return
}

// ServeHTTP allows Service to serve HTTP requests.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.Method, r.URL.Path)
	if strings.HasPrefix(r.URL.Path, "/location/") {
		switch r.Method {
		case "GET":
			s.getLocation(w, r)
		case "POST":
			s.insert(w, r)
		case "PUT":
			s.update(w, r)
		case "DELETE":
			s.deleteLocation(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	} else if r.URL.Path == "/neighbors" {
		s.getNeighbors(w, r)
	} else if r.URL.Path == "/join" {
		s.handleJoin(w, r)
	} else if r.URL.Path == "/remove" {
		s.handleRemove(w, r)
	} else if r.URL.Path == "/leader" {
		s.getLeader(w, r)
	} else if r.URL.Path == "/members" {
		s.getMembers(w, r)
	} else if r.URL.Path == "/isleader" {
		s.isLeader(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	nodeID, remoteAddr, err := prepareJoinArgs(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	if err := s.store.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
}

func (s *Service) handleRemove(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Remove(nodeID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
}

func (s *Service) getLeader(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, string(s.store.GetLeader()))
}

func (s *Service) isLeader(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, strconv.FormatBool(s.store.IsLeader()))
}

func (s *Service) getMembers(w http.ResponseWriter, r *http.Request) {
	members, _ := s.store.Nodes()
	membersStr, _ := json.Marshal(members)
	io.WriteString(w, string(membersStr))
}

func (s *Service) getLocation(w http.ResponseWriter, r *http.Request) {
	locationID, err := getLocationID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	leaf, err := s.store.Get(locationID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	b, err := json.Marshal(map[string]interface{}{
		"lat":  leaf.GetLocation().Lat(),
		"long": leaf.GetLocation().Long(),
		"data": leaf.Data,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setContentTypeJSON(w)
	io.WriteString(w, string(b))
}

func (s *Service) deleteLocation(w http.ResponseWriter, r *http.Request) {
	locationID, err := getLocationID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.Delete(locationID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "ok")
}

func (s *Service) insert(w http.ResponseWriter, r *http.Request) {
	locationID, position, data, err := prepareInsertArgs(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err := s.store.Insert(locationID, position, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, "ok")
}

func (s *Service) update(w http.ResponseWriter, r *http.Request) {
	latExists, lonExists, dataExists, locationID, position, data, err := prepareUpdateArgs(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if latExists && lonExists && dataExists {
		err = s.store.Update(locationID, position, data)
	} else if latExists && lonExists {
		err = s.store.UpdateLocation(locationID, position)
	} else if dataExists {
		err = s.store.UpdateData(locationID, data)
	} else {
		err = errors.New("nothing to update")
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	io.WriteString(w, "ok")
}

func (s *Service) getNeighbors(w http.ResponseWriter, r *http.Request) {
	lat, lon, radius, limit, err := prepareGetNeighborsArg(r)
	if err != nil {
		respondWithErr(w, err)
		return
	}
	neighbors := s.store.GetNeighbors(*ds.NewPosition(lat, lon), radius, limit)
	neighborsStr, _ := json.Marshal(types.PrepareNeighborResults(neighbors))
	resp := string(neighborsStr)
	setContentTypeJSON(w)
	io.WriteString(w, resp)
}

// Addr returns the address on which the Service is listening
func (s *Service) Addr() net.Addr {
	return s.ln.Addr()
}
