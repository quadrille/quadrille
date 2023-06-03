package ds

import (
	"stathat.com/c/consistent"
	"strconv"
	"sync"
	"fmt"
)

type specialMap struct {
	m  map[string]*QuadTreeNode
	mu sync.RWMutex
}

func (m *specialMap) Lock() {
	m.mu.Lock()
}

func (m *specialMap) UnLock() {
	m.mu.Unlock()
}

func (m *specialMap) RLock() {
	m.mu.RLock()
}

func (m *specialMap) RUnLock() {
	m.mu.RUnlock()
}

func (m *specialMap) Set(key string, node *QuadTreeNode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.set(key, node)
}

func (m *specialMap) set(key string, node *QuadTreeNode) {
	m.m[key] = node
}

func (m *specialMap) Get(key string) *QuadTreeNode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.get(key)
}

func (m *specialMap) get(key string) *QuadTreeNode {
	return m.m[key]
}

func (m *specialMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delete(key)
}

func (m *specialMap) delete(key string) {
	delete(m.m, key)
}

const mapConcurrency = 32

type concurrentMap struct {
	maps [mapConcurrency]specialMap
	hash *consistent.Consistent
}

func (m *concurrentMap) getMapIndexFomKey(key string) (mapIndex int) {
	indexStr, err := m.hash.Get(key)
	if err == nil {
		mapIndex, _ = strconv.Atoi(indexStr)
	}
	return
}

func NewMap() *concurrentMap {
	m := concurrentMap{
		maps: [mapConcurrency]specialMap{},
		hash: consistent.New(),
	}
	for i := 0; i < mapConcurrency; i++ {
		m.hash.Add(fmt.Sprint(i))
		m.maps[i] = specialMap{m: map[string]*QuadTreeNode{}}
	}
	return &m
}

func (m *concurrentMap) getMapFromKey(key string) *specialMap {
	mapIndex := m.getMapIndexFomKey(key)
	return &m.maps[mapIndex]
}

func (m *concurrentMap) Set(key string, val *QuadTreeNode) {
	cMap := m.getMapFromKey(key)
	cMap.Set(key, val)
}

func (m *concurrentMap) SetUnsafe(key string, val *QuadTreeNode) {
	cMap := m.getMapFromKey(key)
	cMap.set(key, val)
}

func (m *concurrentMap) Get(key string) *QuadTreeNode {
	cMap := m.getMapFromKey(key)
	return cMap.Get(key)
}

func (m *concurrentMap) GetUnsafe(key string) *QuadTreeNode {
	cMap := m.getMapFromKey(key)
	return cMap.get(key)
}

func (m *concurrentMap) Delete(key string) {
	cMap := m.getMapFromKey(key)
	cMap.Delete(key)
}

func (m *concurrentMap) DeleteUnsafe(key string) {
	cMap := m.getMapFromKey(key)
	cMap.delete(key)
}

func (m *concurrentMap) Lock(key string) {
	//fmt.Printf("🔑Locking %d\n", m.getMapIndexFomKey(key))
	cMap := m.getMapFromKey(key)
	cMap.Lock()
}

func (m *concurrentMap) UnLock(key string) {
	//fmt.Printf("🔑UnLocking %d\n", m.getMapIndexFomKey(key))
	cMap := m.getMapFromKey(key)
	cMap.UnLock()
}

func (m *concurrentMap) RLock(key string) {
	//fmt.Printf("🔑RLocking %d\n", m.getMapIndexFomKey(key))
	cMap := m.getMapFromKey(key)
	cMap.RLock()
}

func (m *concurrentMap) RUnLock(key string) {
	//fmt.Printf("🔑RLocking %d\n", m.getMapIndexFomKey(key))
	cMap := m.getMapFromKey(key)
	cMap.RUnLock()
}

func (m *concurrentMap) GetAllKeyVal() map[string]*QuadTreeNode {
	allMap := map[string]*QuadTreeNode{}
	maps := m.maps
	for _, cMap := range maps {
		cMap.RLock()
		for k, v := range cMap.m {
			allMap[k] = v
		}
		cMap.RUnLock()
	}
	return allMap
}
