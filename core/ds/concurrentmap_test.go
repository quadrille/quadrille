package ds

import (
	"fmt"
	"testing"
)

func TestConcurrentMap(t *testing.T) {
	m := NewMap()
	m.Set("go", &QuadTreeNode{})
	m.Set("node.js", &QuadTreeNode{})

	fmt.Println(m.GetAllKeyVal())
}
