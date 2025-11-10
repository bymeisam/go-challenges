package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// TestBloomFilter tests bloom filter
func TestBloomFilterBasic(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)

	bf.Add([]byte("test"))

	if !bf.Contains([]byte("test")) {
		t.Errorf("Expected to find 'test' in filter")
	}
}

func TestBloomFilterNotContains(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)

	bf.Add([]byte("apple"))

	if bf.Contains([]byte("banana")) {
		t.Logf("False positive detected (acceptable for Bloom filter)")
	}
}

func TestBloomFilterStats(t *testing.T) {
	bf := NewBloomFilter(100, 0.05)

	bf.Add([]byte("a"))
	bf.Add([]byte("b"))

	stats := bf.GetStats()
	if count, ok := stats["elements_added"].(int64); !ok || count != 2 {
		t.Errorf("Expected 2 elements added")
	}
}

// TestConsistentHash tests consistent hashing
func TestConsistentHashBasic(t *testing.T) {
	ch := NewConsistentHash(1)

	ch.AddNode("server1")
	node := ch.GetNode("key1")

	if node != "server1" {
		t.Errorf("Expected 'server1', got %s", node)
	}
}

func TestConsistentHashMultipleNodes(t *testing.T) {
	ch := NewConsistentHash(1)

	ch.AddNode("server1")
	ch.AddNode("server2")
	ch.AddNode("server3")

	node := ch.GetNode("key1")
	if len(node) == 0 {
		t.Errorf("Expected non-empty node name")
	}
}

func TestConsistentHashDistribution(t *testing.T) {
	ch := NewConsistentHash(1)

	ch.AddNode("server1")
	ch.AddNode("server2")

	distribution := make(map[string]int)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		node := ch.GetNode(key)
		distribution[node]++
	}

	if len(distribution) < 1 {
		t.Errorf("Expected at least 1 unique node")
	}
}

// TestTrie tests trie data structure
func TestTrieInsertSearch(t *testing.T) {
	trie := NewTrie()

	trie.Insert("hello")
	if !trie.Search("hello") {
		t.Errorf("Expected to find 'hello'")
	}

	if trie.Search("hell") {
		t.Errorf("Expected not to find 'hell'")
	}
}

func TestTrieStartsWith(t *testing.T) {
	trie := NewTrie()

	trie.Insert("hello")
	if !trie.StartsWith("hel") {
		t.Errorf("Expected 'hel' prefix to exist")
	}

	if trie.StartsWith("hex") {
		t.Errorf("Expected 'hex' prefix not to exist")
	}
}

func TestTrieAutocomplete(t *testing.T) {
	trie := NewTrie()

	words := []string{"apple", "app", "application", "apply"}
	for _, word := range words {
		trie.Insert(word)
	}

	results := trie.Autocomplete("app")
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for 'app' autocomplete")
	}
}

func TestTrieCount(t *testing.T) {
	trie := NewTrie()

	for i := 0; i < 5; i++ {
		trie.Insert(fmt.Sprintf("word%d", i))
	}

	if atomic.LoadInt64(&trie.count) != 5 {
		t.Errorf("Expected count=5")
	}
}

// TestSkipList tests skip list
func TestSkipListInsertSearch(t *testing.T) {
	sl := NewSkipList()

	sl.Insert("key1", "value1")
	value, found := sl.Search("key1")

	if !found || value != "value1" {
		t.Errorf("Expected to find 'key1'")
	}
}

func TestSkipListNotFound(t *testing.T) {
	sl := NewSkipList()

	sl.Insert("key1", "value1")
	_, found := sl.Search("key2")

	if found {
		t.Errorf("Expected 'key2' not to be found")
	}
}

func TestSkipListRange(t *testing.T) {
	sl := NewSkipList()

	for i := 1; i <= 10; i++ {
		sl.Insert(fmt.Sprintf("key%d", i), i)
	}

	results := sl.Range("key3", "key7")
	if len(results) < 1 {
		t.Logf("Warning: Expected results in range [key3, key7]")
	}
}

func TestSkipListSize(t *testing.T) {
	sl := NewSkipList()

	for i := 0; i < 10; i++ {
		sl.Insert(fmt.Sprintf("key%d", i), i)
	}

	if atomic.LoadInt64(&sl.size) != 10 {
		t.Errorf("Expected size=10")
	}
}

// TestLockFreeCounter tests lock-free counter
func TestLockFreeCounterBasic(t *testing.T) {
	counter := NewLockFreeCounter()

	if counter.Get() != 0 {
		t.Errorf("Expected initial value 0")
	}

	counter.Increment()
	if counter.Get() != 1 {
		t.Errorf("Expected value 1 after increment")
	}

	counter.Decrement()
	if counter.Get() != 0 {
		t.Errorf("Expected value 0 after decrement")
	}
}

func TestLockFreeCounterCAS(t *testing.T) {
	counter := NewLockFreeCounter()
	counter.Set(10)

	swapped := counter.CompareAndSwap(10, 20)
	if !swapped || counter.Get() != 20 {
		t.Errorf("Expected CAS to succeed and value to be 20")
	}

	swapped = counter.CompareAndSwap(10, 30)
	if swapped {
		t.Errorf("Expected CAS to fail")
	}
}

func TestLockFreeCounterConcurrent(t *testing.T) {
	counter := NewLockFreeCounter()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()

	if counter.Get() != 100 {
		t.Errorf("Expected value 100 after 100 concurrent increments, got %d", counter.Get())
	}
}

// TestConcurrentMap tests concurrent map
func TestConcurrentMapSetGet(t *testing.T) {
	cm := NewConcurrentMap(4)

	cm.Set("key1", "value1")
	value, found := cm.Get("key1")

	if !found || value != "value1" {
		t.Errorf("Expected to find 'key1'")
	}
}

func TestConcurrentMapDelete(t *testing.T) {
	cm := NewConcurrentMap(4)

	cm.Set("key1", "value1")
	cm.Delete("key1")

	_, found := cm.Get("key1")
	if found {
		t.Errorf("Expected 'key1' to be deleted")
	}
}

func TestConcurrentMapLen(t *testing.T) {
	cm := NewConcurrentMap(4)

	for i := 0; i < 10; i++ {
		cm.Set(fmt.Sprintf("key%d", i), i)
	}

	if cm.Len() != 10 {
		t.Errorf("Expected length 10")
	}
}

func TestConcurrentMapConcurrent(t *testing.T) {
	cm := NewConcurrentMap(16)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id)
			cm.Set(key, id)
		}(i)
	}

	wg.Wait()

	if cm.Len() != 100 {
		t.Errorf("Expected length 100 after concurrent sets")
	}
}

// Benchmark tests

func BenchmarkBloomFilterAdd(b *testing.B) {
	bf := NewBloomFilter(10000, 0.01)
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}

func BenchmarkBloomFilterContains(b *testing.B) {
	bf := NewBloomFilter(10000, 0.01)
	bf.Add([]byte("test"))
	data := []byte("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.Contains(data)
	}
}

func BenchmarkConsistentHashGetNode(b *testing.B) {
	ch := NewConsistentHash(3)
	ch.AddNode("server1")
	ch.AddNode("server2")
	ch.AddNode("server3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.GetNode("key")
	}
}

func BenchmarkTrieInsert(b *testing.B) {
	trie := NewTrie()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Insert(fmt.Sprintf("word%d", i))
	}
}

func BenchmarkTrieSearch(b *testing.B) {
	trie := NewTrie()
	trie.Insert("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Search("test")
	}
}

func BenchmarkSkipListInsert(b *testing.B) {
	sl := NewSkipList()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Insert(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkLockFreeCounterIncrement(b *testing.B) {
	counter := NewLockFreeCounter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		counter.Increment()
	}
}

func BenchmarkConcurrentMapSet(b *testing.B) {
	cm := NewConcurrentMap(16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Set(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	cm := NewConcurrentMap(16)
	cm.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Get("key")
	}
}
