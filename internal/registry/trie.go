package registry

import (
	"encoding/gob"
	"fmt"
	"io"
)

type TrieNode struct {
	Children map[byte]*TrieNode
	Records  []Record
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		Records:  nil,
		Children: make(map[byte]*TrieNode),
	}
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Root: NewTrieNode(),
	}
}

func (t *Trie) Insert(record Record) {
	node := t.Root
	for _, b := range record.Assignment {
		if _, found := node.Children[b]; !found {
			node.Children[b] = NewTrieNode()
		}
		node = node.Children[b]
	}
	node.Records = append(node.Records, record)
}

func (t *Trie) InsertMany(records []Record) {
	for _, record := range records {
		t.Insert(record)
	}
}

func (t *Trie) Lookup(prefix []byte) []Record {
	node := t.Root
	for _, b := range prefix {
		if nextNode, found := node.Children[b]; found {
			node = nextNode
		} else {
			return nil
		}
	}

	return node.Records
}

func (t *Trie) LongestPrefixMatch(prefix []byte) []Record {
	node := t.Root
	var longestMatchNode *TrieNode

	for _, b := range prefix {
		if _, exists := node.Children[b]; !exists {
			break
		}
		node = node.Children[b]
		if len(node.Records) > 0 {
			longestMatchNode = node
		}
	}
	if longestMatchNode != nil {
		return longestMatchNode.Records
	}

	return []Record{}
}

func (t *Trie) Traverse() []Record {
	var allRecords []Record
	if t.Root == nil {
		return allRecords
	}

	stack := []*TrieNode{t.Root}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		allRecords = append(allRecords, node.Records...)

		for _, child := range node.Children {
			stack = append(stack, child)
		}
	}

	return allRecords
}

func (t *Trie) EncodeGOB(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(t); err != nil {
		return fmt.Errorf("error encoding trie: %w", err)
	}

	return nil
}

func (t *Trie) DecodeGOB(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(t); err != nil {
		return fmt.Errorf("error decoding trie: %w", err)
	}

	return nil
}
