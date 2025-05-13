package internal

import (
	"slices"
	"strings"
)

var vowels = []rune{
	'a', 'á', 'â', 'ã', 'à',
	'e', 'é', 'ê',
	'i', 'í',
	'o', 'ó', 'ô', 'õ',
	'u', 'ú',
}

type SuffixTrie struct {
	root *node
}

type node struct {
	key      rune
	children []*node
	values   []string
}

func NewSuffixTrie() *SuffixTrie {
	return &SuffixTrie{
		root: &node{
			children: []*node{},
		},
	}
}

func (t *SuffixTrie) Insert(phrase string) {
	reversedSuffix := t.getReversedSuffix(phrase)
	currentNode := t.root
	for i := range reversedSuffix {
		n, index := t.findChild(currentNode.children, reversedSuffix[i])
		if n != nil {
			currentNode = n
		} else {
			n = &node{
				key:      reversedSuffix[i],
				children: []*node{},
			}
			currentNode.children = slices.Insert(currentNode.children, index, n)
			currentNode = n
		}
	}
	phrase = strings.ReplaceAll(phrase, "[", "")
	phrase = strings.ReplaceAll(phrase, "]", "")

	if !slices.Contains(currentNode.values, phrase) {
		currentNode.values = append(currentNode.values, phrase)
	}
}

func (t *SuffixTrie) Search(phrase string) []string {
	i := len(phrase) - 1
	currentNode := t.root
	possibleValues := []string{}
	for i >= 0 && currentNode != nil && phrase[i] != ' ' {
		if currentNode.values != nil {
			possibleValues = append(possibleValues, currentNode.values...)
		}
		n, _ := t.findChild(currentNode.children, rune(phrase[i]))
		if n != nil {
			currentNode = n
		} else {
			currentNode = nil
		}
		i--
	}
	return possibleValues
}

func (t *SuffixTrie) GetSortedValues() []string {
	return t.inOrderTraversal(t.root)
}

func (t *SuffixTrie) inOrderTraversal(node *node) []string {
	values := []string{}
	if node == nil {
		return values
	}
	for _, child := range node.children {
		values = append(values, t.inOrderTraversal(child)...)
	}
	slices.Sort(node.values)
	values = append(values, node.values...)

	return values
}

// findChild do a binary search to find the child node with the given key
// if the key is not found, it returns the index where the key should be inserted
func (t *SuffixTrie) findChild(children []*node, key rune) (*node, int) {
	low := 0
	high := len(children) - 1
	for low <= high {
		mid := (low + high) / 2
		if children[mid].key == key {
			return children[mid], mid
		}
		if children[mid].key < key {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return nil, low
}

// we are expecting here a phrase with the stressed syllable marked with square brackets:
// - "meu pau na sua [mão]"
// - "meu pau limpou seu [den]te"
// we then return the last runes, reversed, until the last vowel of the stressed syllable
// for example:
// - "meu pau na sua [mão]" will return "oã"
// - "meu pau limpou seu [den]te" will return "etne"
func (t *SuffixTrie) getReversedSuffix(phrase string) []rune {
	i := len(phrase) - 1
	for phrase[i] != ']' {
		i--
	}
	for !slices.Contains(vowels, rune(phrase[i])) && phrase[i] != '[' {
		i--
	}
	runes := []rune(strings.ReplaceAll(phrase[i:], "]", ""))
	slices.Reverse(runes)

	return runes
}
