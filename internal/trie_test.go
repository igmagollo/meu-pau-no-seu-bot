package internal

import "testing"

func TestSuffixTrie(t *testing.T) {
	t.Run("test insert/search", func(t *testing.T) {
		trie := NewSuffixTrie()
		trie.Insert("meu pau na sua [mão]")
		trie.Insert("meu pau limpou seu [den]te")
		results := trie.Search("meu pau na sua mão")

		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0] != "meu pau na sua mão" {
			t.Errorf("expected 'meu pau na sua mão', got %s", results[0])
		}

		results = trie.Search("meu pau limpou seu dente")
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0] != "meu pau limpou seu dente" {
			t.Errorf("expected 'meu pau limpou seu dente', got %s", results[0])
		}

		results = trie.Search("comi pão")
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0] != "meu pau na sua mão" {
			t.Errorf("expected 'meu pau na sua mão', got %s", results[0])
		}

		results = trie.Search("realmente")
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0] != "meu pau limpou seu dente" {
			t.Errorf("expected 'meu pau limpou seu dente', got %s", results[0])
		}

		trie.Insert("meu pau é delin[quen]te")
		results = trie.Search("realmente")
		if len(results) != 2 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0] != "meu pau limpou seu dente" {
			t.Errorf("expected 'meu pau limpou seu dente', got %s", results[0])
		}
		if results[1] != "meu pau é delinquente" {
			t.Errorf("expected 'meu pau é delinquente', got %s", results[1])
		}
	})

	t.Run("test get sorted values", func(t *testing.T) {
		trie := NewSuffixTrie()
		trie.Insert("meu pau na sua [mão]")
		trie.Insert("meu pau limpou seu [den]te")
		trie.Insert("meu pau é delin[quen]te")
		trie.Insert("meu pau te fez ci[la]da")
		trie.Insert("meu pau te deu uma de[da]da")
		trie.Insert("meu pau te cu[tu]ca")
		results := trie.GetSortedValues()
		if len(results) != 6 {
			t.Errorf("expected 6 results, got %d", len(results))
		}
		if results[0] != "meu pau te cutuca" {
			t.Errorf("expected 'meu pau te cutuca', got %s", results[0])
		}
		if results[1] != "meu pau te deu uma dedada" {
			t.Errorf("expected 'meu pau te deu uma dedada', got %s", results[1])
		}
		if results[2] != "meu pau te fez cilada" {
			t.Errorf("expected 'meu pau te fez cilada', got %s", results[2])
		}
		if results[3] != "meu pau limpou seu dente" {
			t.Errorf("expected 'meu pau limpou seu dente', got %s", results[3])
		}
		if results[4] != "meu pau é delinquente" {
			t.Errorf("expected 'meu pau é delinquente', got %s", results[4])
		}
		if results[5] != "meu pau na sua mão" {
			t.Errorf("expected 'meu pau na sua mão', got %s", results[5])
		}
	})
}
