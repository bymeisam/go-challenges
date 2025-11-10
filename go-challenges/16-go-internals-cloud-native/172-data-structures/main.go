package main

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
)

// Challenge 172: Advanced Data Structures
// Bloom Filter, Consistent Hashing, Trie, Skip List, Lock-Free Structures

// ===== 1. Bloom Filter =====

type BloomFilter struct {
	bits      []uint64
	hashCount int
	size      uint64
	count     int64
	mu        sync.RWMutex
}

func NewBloomFilter(expectedElements int, falsePositiveRate float64) *BloomFilter {
	size := calculateBloomFilterSize(expectedElements, falsePositiveRate)
	hashCount := calculateHashFunctions(size, expectedElements)

	return &BloomFilter{
		bits:      make([]uint64, (size+63)/64),
		hashCount: hashCount,
		size:      size,
	}
}

func calculateBloomFilterSize(n int, p float64) uint64 {
	return uint64(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2))
}

func calculateHashFunctions(m uint64, n int) int {
	return int(float64(m) / float64(n) * math.Log(2))
}

func (bf *BloomFilter) Add(data []byte) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := 0; i < bf.hashCount; i++ {
		pos := bf.hash(data, i) % bf.size
		bf.setBit(pos)
	}

	atomic.AddInt64(&bf.count, 1)
}

func (bf *BloomFilter) Contains(data []byte) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for i := 0; i < bf.hashCount; i++ {
		pos := bf.hash(data, i) % bf.size
		if !bf.getBit(pos) {
			return false
		}
	}

	return true
}

func (bf *BloomFilter) hash(data []byte, seed int) uint64 {
	h := fnv.New64a()
	h.Write(data)
	h.Write([]byte{byte(seed)})
	return h.Sum64()
}

func (bf *BloomFilter) setBit(pos uint64) {
	idx := pos / 64
	offset := pos % 64
	bf.bits[idx] |= 1 << offset
}

func (bf *BloomFilter) getBit(pos uint64) bool {
	idx := pos / 64
	offset := pos % 64
	return (bf.bits[idx] & (1 << offset)) != 0
}

func (bf *BloomFilter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"elements_added": atomic.LoadInt64(&bf.count),
		"filter_size":    bf.size,
		"hash_functions": bf.hashCount,
	}
}

// ===== 2. Consistent Hashing =====

type ConsistentHash struct {
	ring            map[uint64]string
	nodes           []uint64
	virtualNodesPerNode int
	mu              sync.RWMutex
	lookups         int64
}

func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		ring:                make(map[uint64]string),
		nodes:               make([]uint64, 0),
		virtualNodesPerNode: virtualNodes,
	}
}

func (ch *ConsistentHash) AddNode(nodeName string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	for i := 0; i < ch.virtualNodesPerNode; i++ {
		key := ch.hashKey(nodeName, i)
		ch.ring[key] = nodeName
		ch.nodes = append(ch.nodes, key)
	}

	sort.Slice(ch.nodes, func(i, j int) bool {
		return ch.nodes[i] < ch.nodes[j]
	})
}

func (ch *ConsistentHash) RemoveNode(nodeName string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	for i := 0; i < ch.virtualNodesPerNode; i++ {
		key := ch.hashKey(nodeName, i)
		delete(ch.ring, key)
	}

	// Rebuild sorted nodes
	ch.nodes = make([]uint64, 0)
	for k := range ch.ring {
		ch.nodes = append(ch.nodes, k)
	}
	sort.Slice(ch.nodes, func(i, j int) bool {
		return ch.nodes[i] < ch.nodes[j]
	})
}

func (ch *ConsistentHash) GetNode(key string) string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.nodes) == 0 {
		return ""
	}

	atomic.AddInt64(&ch.lookups, 1)

	hash := ch.hashKey(key, 0)
	idx := sort.Search(len(ch.nodes), func(i int) bool {
		return ch.nodes[i] >= hash
	})

	if idx == len(ch.nodes) {
		idx = 0
	}

	return ch.ring[ch.nodes[idx]]
}

func (ch *ConsistentHash) hashKey(key string, seed int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	h.Write([]byte{byte(seed)})
	return h.Sum64()
}

// ===== 3. Trie Data Structure =====

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	mu       sync.RWMutex
}

type Trie struct {
	root   *TrieNode
	mu     sync.RWMutex
	count  int64
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

func (t *Trie) Insert(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, ch := range word {
		node.mu.Lock()
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		next := node.children[ch]
		node.mu.Unlock()
		node = next
	}

	node.mu.Lock()
	node.isEnd = true
	node.mu.Unlock()

	atomic.AddInt64(&t.count, 1)
}

func (t *Trie) Search(word string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range word {
		node.mu.RLock()
		next, exists := node.children[ch]
		node.mu.RUnlock()

		if !exists {
			return false
		}
		node = next
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	return node.isEnd
}

func (t *Trie) StartsWith(prefix string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range prefix {
		node.mu.RLock()
		next, exists := node.children[ch]
		node.mu.RUnlock()

		if !exists {
			return false
		}
		node = next
	}

	return true
}

func (t *Trie) Autocomplete(prefix string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range prefix {
		node.mu.RLock()
		next, exists := node.children[ch]
		node.mu.RUnlock()

		if !exists {
			return []string{}
		}
		node = next
	}

	results := make([]string, 0)
	t.dfs(node, prefix, &results)
	return results
}

func (t *Trie) dfs(node *TrieNode, prefix string, results *[]string) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.isEnd {
		*results = append(*results, prefix)
	}

	for ch, child := range node.children {
		t.dfs(child, prefix+string(ch), results)
	}
}

// ===== 4. Skip List =====

type SkipListNode struct {
	value   interface{}
	key     string
	forward []*SkipListNode
	level   int
}

type SkipList struct {
	header *SkipListNode
	level  int
	size   int64
	mu     sync.RWMutex
}

func NewSkipList() *SkipList {
	return &SkipList{
		header: &SkipListNode{
			level:   16,
			forward: make([]*SkipListNode, 16),
		},
		level: 0,
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for rand.Float64() < 0.5 && level < 16 {
		level++
	}
	return level
}

func (sl *SkipList) Insert(key string, value interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	level := sl.randomLevel()
	if level > sl.level {
		sl.level = level
	}

	newNode := &SkipListNode{
		key:     key,
		value:   value,
		level:   level,
		forward: make([]*SkipListNode, level),
	}

	current := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}

		if i < level {
			newNode.forward[i] = current.forward[i]
			current.forward[i] = newNode
		}
	}

	atomic.AddInt64(&sl.size, 1)
}

func (sl *SkipList) Search(key string) (interface{}, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	current := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
	}

	current = current.forward[0]
	if current != nil && current.key == key {
		return current.value, true
	}

	return nil, false
}

func (sl *SkipList) Range(startKey, endKey string) []interface{} {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	results := make([]interface{}, 0)

	current := sl.header.forward[0]
	for current != nil && current.key <= endKey {
		if current.key >= startKey {
			results = append(results, current.value)
		}
		current = current.forward[0]
	}

	return results
}

// ===== 5. Lock-Free Counter =====

type LockFreeCounter struct {
	value int64
}

func NewLockFreeCounter() *LockFreeCounter {
	return &LockFreeCounter{value: 0}
}

func (lfc *LockFreeCounter) Increment() {
	atomic.AddInt64(&lfc.value, 1)
}

func (lfc *LockFreeCounter) Decrement() {
	atomic.AddInt64(&lfc.value, -1)
}

func (lfc *LockFreeCounter) Get() int64 {
	return atomic.LoadInt64(&lfc.value)
}

func (lfc *LockFreeCounter) Set(value int64) {
	atomic.StoreInt64(&lfc.value, value)
}

func (lfc *LockFreeCounter) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&lfc.value, old, new)
}

// ===== 6. Concurrent Map =====

type ConcurrentMap struct {
	shards    []*MapShard
	shardCount int
}

type MapShard struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func NewConcurrentMap(shardCount int) *ConcurrentMap {
	shards := make([]*MapShard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &MapShard{
			items: make(map[string]interface{}),
		}
	}

	return &ConcurrentMap{
		shards:     shards,
		shardCount: shardCount,
	}
}

func (cm *ConcurrentMap) getShard(key string) *MapShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() % uint32(cm.shardCount)
	return cm.shards[idx]
}

func (cm *ConcurrentMap) Set(key string, value interface{}) {
	shard := cm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = value
}

func (cm *ConcurrentMap) Get(key string) (interface{}, bool) {
	shard := cm.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	value, exists := shard.items[key]
	return value, exists
}

func (cm *ConcurrentMap) Delete(key string) {
	shard := cm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

func (cm *ConcurrentMap) Len() int {
	total := 0
	for _, shard := range cm.shards {
		shard.mu.RLock()
		total += len(shard.items)
		shard.mu.RUnlock()
	}
	return total
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Advanced Data Structures ===\n")

	// 1. Bloom Filter
	fmt.Println("1. Bloom Filter")
	bf := NewBloomFilter(1000, 0.01)

	items := []string{"apple", "banana", "cherry"}
	for _, item := range items {
		bf.Add([]byte(item))
	}

	fmt.Printf("Contains 'apple': %v\n", bf.Contains([]byte("apple")))
	fmt.Printf("Contains 'grape': %v\n", bf.Contains([]byte("grape")))
	fmt.Printf("Bloom Filter stats: %+v\n\n", bf.GetStats())

	// 2. Consistent Hashing
	fmt.Println("2. Consistent Hashing")
	ch := NewConsistentHash(3)

	ch.AddNode("server1")
	ch.AddNode("server2")
	ch.AddNode("server3")

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, key := range keys {
		node := ch.GetNode(key)
		fmt.Printf("Key '%s' -> %s\n", key, node)
	}
	fmt.Println()

	// 3. Trie
	fmt.Println("3. Trie Data Structure")
	trie := NewTrie()

	words := []string{"apple", "app", "application", "apply", "apricot"}
	for _, word := range words {
		trie.Insert(word)
	}

	fmt.Printf("Search 'apple': %v\n", trie.Search("apple"))
	fmt.Printf("Search 'app': %v\n", trie.Search("app"))
	fmt.Printf("StartsWith 'app': %v\n", trie.StartsWith("app"))

	autocomplete := trie.Autocomplete("app")
	fmt.Printf("Autocomplete 'app': %v\n\n", autocomplete)

	// 4. Skip List
	fmt.Println("4. Skip List")
	sl := NewSkipList()

	for i := 1; i <= 10; i++ {
		sl.Insert(fmt.Sprintf("key%d", i), i*10)
	}

	value, found := sl.Search("key5")
	fmt.Printf("Search 'key5': found=%v, value=%v\n", found, value)

	rangeResults := sl.Range("key3", "key7")
	fmt.Printf("Range [key3, key7]: %v\n\n", rangeResults)

	// 5. Lock-Free Counter
	fmt.Println("5. Lock-Free Counter")
	counter := NewLockFreeCounter()

	for i := 0; i < 100; i++ {
		counter.Increment()
	}

	fmt.Printf("Counter value: %d\n", counter.Get())

	swapped := counter.CompareAndSwap(100, 200)
	fmt.Printf("CompareAndSwap(100, 200): %v, value=%d\n\n", swapped, counter.Get())

	// 6. Concurrent Map
	fmt.Println("6. Concurrent Map")
	cmap := NewConcurrentMap(16)

	cmap.Set("name", "Alice")
	cmap.Set("age", 30)
	cmap.Set("city", "NYC")

	fmt.Printf("Get 'name': %v\n", cmap.Get("name"))
	fmt.Printf("Get 'age': %v\n", cmap.Get("age"))
	fmt.Printf("Map size: %d\n\n", cmap.Len())

	// 7. Features Summary
	fmt.Println("7. Advanced Data Structures Features")
	fmt.Println("  - Bloom filter with configurable false positive rate")
	fmt.Println("  - Consistent hashing with virtual nodes")
	fmt.Println("  - Trie with prefix matching and autocomplete")
	fmt.Println("  - Skip list with range queries")
	fmt.Println("  - Lock-free counter with CAS operations")
	fmt.Println("  - Sharded concurrent map")

	fmt.Println("\n=== Complete ===")
}
