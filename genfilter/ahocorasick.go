package main

import (
	"encoding/gob"
	"os"

	"github.com/anknown/darts"
	"github.com/pkg/errors"
)

const (
	FailState = -1
	RootState = 1
)

type Machine struct {
	Trie    *godarts.DoubleArrayTrie
	Failure []int
	Output  map[int]struct{}
}

func buildMachine(paths [][]rune) (*Machine, error) {
	var d godarts.Darts
	dat, llt, err := d.Build(paths)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build double array")
	}

	output := make(map[int]struct{}, len(d.Output))
	for state := range d.Output {
		output[state] = struct{}{}
	}

	failure := make([]int, len(dat.Base))
	for _, c := range llt.Root.Children {
		failure[c.Base] = godarts.ROOT_NODE_BASE
	}

	m := &Machine{
		Trie:    dat,
		Failure: failure,
		Output:  output,
	}

	queue := make([]*godarts.LinkedListTrieNode, len(llt.Root.Children))
	copy(queue, llt.Root.Children)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for _, n := range node.Children {
			if n.Base == godarts.END_NODE_BASE {
				continue
			}
			input := n.Code - godarts.ROOT_NODE_BASE
			outState := FailState
			for inState := node.Base; outState == FailState; {
				inState = m.Failure[inState]
				outState = m.g(inState, input)
			}
			if _, ok := m.Output[outState]; ok {
				m.Output[n.Base] = struct{}{}
			}
			m.Failure[n.Base] = outState
		}
		queue = append(queue, node.Children...)
	}
	return m, nil
}

func loadMachine(path string) (*Machine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open machine")
	}
	defer f.Close()

	var m Machine
	if err := gob.NewDecoder(f).Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode machine")
	}
	return &m, nil
}

func (m *Machine) g(inState int, input rune) int {
	if inState == FailState {
		return RootState
	}
	t := inState + int(input) + godarts.ROOT_NODE_BASE
	if t < len(m.Trie.Base) && inState == m.Trie.Check[t] {
		return m.Trie.Base[t]
	}
	if inState == RootState {
		return RootState
	}
	return FailState
}

func (m *Machine) Contains(r []rune) bool {
	state := RootState
	for _, c := range r {
		for {
			next := m.g(state, c)
			if next != FailState {
				state = next
				break
			}
			state = m.Failure[state]
		}
		if _, ok := m.Output[state]; ok {
			return true
		}
	}
	return false
}
