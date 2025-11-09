package cache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHash provides consistent hashing for distributed caching
type ConsistentHash struct {
	mu           sync.RWMutex
	hashRing     map[uint32]string
	sortedHashes []uint32
	virtualNodes int
	nodes        map[string]bool
}

// NewConsistentHash creates a new consistent hash ring
func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		hashRing:     make(map[uint32]string),
		sortedHashes: make([]uint32, 0),
		virtualNodes: virtualNodes,
		nodes:        make(map[string]bool),
	}
}

// Add adds a node to the hash ring
func (ch *ConsistentHash) Add(node string, weight int) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.nodes[node] {
		return // Node already exists
	}

	ch.nodes[node] = true

	// Add virtual nodes based on weight
	virtualNodeCount := ch.virtualNodes * weight
	for i := 0; i < virtualNodeCount; i++ {
		virtualNodeKey := fmt.Sprintf("%s:%d", node, i)
		hash := ch.hash(virtualNodeKey)
		ch.hashRing[hash] = node
		ch.sortedHashes = append(ch.sortedHashes, hash)
	}

	// Sort the hash ring
	sort.Slice(ch.sortedHashes, func(i, j int) bool {
		return ch.sortedHashes[i] < ch.sortedHashes[j]
	})
}

// Remove removes a node from the hash ring
func (ch *ConsistentHash) Remove(node string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.nodes[node] {
		return // Node doesn't exist
	}

	delete(ch.nodes, node)

	// Remove all virtual nodes for this physical node
	newSortedHashes := make([]uint32, 0, len(ch.sortedHashes))
	for _, hash := range ch.sortedHashes {
		if ch.hashRing[hash] != node {
			newSortedHashes = append(newSortedHashes, hash)
		} else {
			delete(ch.hashRing, hash)
		}
	}

	ch.sortedHashes = newSortedHashes
}

// Get returns the node responsible for the given key
func (ch *ConsistentHash) Get(key string) (string, bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.sortedHashes) == 0 {
		return "", false
	}

	hash := ch.hash(key)

	// Binary search for the first node with hash >= key hash
	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	// If no node found, wrap around to the first node
	if idx == len(ch.sortedHashes) {
		idx = 0
	}

	nodeHash := ch.sortedHashes[idx]
	node := ch.hashRing[nodeHash]

	return node, true
}

// GetMultiple returns multiple nodes for replication
func (ch *ConsistentHash) GetMultiple(key string, count int) []string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.sortedHashes) == 0 || count <= 0 {
		return nil
	}

	hash := ch.hash(key)
	nodes := make([]string, 0, count)
	uniqueNodes := make(map[string]bool)

	// Find starting position
	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	// Collect unique nodes
	for len(nodes) < count && len(uniqueNodes) < len(ch.nodes) {
		if idx >= len(ch.sortedHashes) {
			idx = 0 // Wrap around
		}

		nodeHash := ch.sortedHashes[idx]
		node := ch.hashRing[nodeHash]

		if !uniqueNodes[node] {
			nodes = append(nodes, node)
			uniqueNodes[node] = true
		}

		idx++
	}

	return nodes
}

// GetNodes returns all nodes in the hash ring
func (ch *ConsistentHash) GetNodes() []string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	nodes := make([]string, 0, len(ch.nodes))
	for node := range ch.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// Size returns the number of nodes in the hash ring
func (ch *ConsistentHash) Size() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	return len(ch.nodes)
}

// Distribution returns the distribution of keys across nodes
func (ch *ConsistentHash) Distribution() map[string]float64 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.sortedHashes) == 0 {
		return nil
	}

	distribution := make(map[string]float64)
	nodeRanges := make(map[string]uint64)

	// Calculate the range each node is responsible for
	for i, hash := range ch.sortedHashes {
		node := ch.hashRing[hash]

		var rangeSize uint64
		if i == 0 {
			// First node: from last hash to current + from 0 to current
			lastHash := ch.sortedHashes[len(ch.sortedHashes)-1]
			rangeSize = uint64(hash) + (uint64(^uint32(0)) - uint64(lastHash))
		} else {
			// Normal range: current - previous
			prevHash := ch.sortedHashes[i-1]
			rangeSize = uint64(hash) - uint64(prevHash)
		}

		nodeRanges[node] += rangeSize
	}

	// Convert to percentages
	totalRange := uint64(^uint32(0))
	for node, nodeRange := range nodeRanges {
		distribution[node] = float64(nodeRange) / float64(totalRange) * 100
	}

	return distribution
}

// hash generates a hash for the given key
func (ch *ConsistentHash) hash(key string) uint32 {
	h := sha256.Sum256([]byte(key))
	return uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
}

// RebalanceInfo provides information about data that needs to be moved when nodes change
type RebalanceInfo struct {
	FromNode string
	ToNode   string
	KeyRange KeyRange
	KeyCount int64
}

// KeyRange represents a range of keys
type KeyRange struct {
	Start uint32
	End   uint32
}

// GetRebalanceInfo returns information about what data needs to be moved when adding/removing nodes
func (ch *ConsistentHash) GetRebalanceInfo(oldRing *ConsistentHash) []RebalanceInfo {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if oldRing == nil {
		return nil
	}

	oldRing.mu.RLock()
	defer oldRing.mu.RUnlock()

	rebalanceInfo := make([]RebalanceInfo, 0)

	// Sample key space to determine what data needs to move
	samplePoints := 1000
	maxHash := uint64(^uint32(0))

	for i := 0; i < samplePoints; i++ {
		keyHash := uint32(uint64(i) * maxHash / uint64(samplePoints))
		_ = strconv.FormatUint(uint64(keyHash), 10) // key not used, suppress warning

		oldNode, oldExists := oldRing.getNodeForHash(keyHash)
		newNode, newExists := ch.getNodeForHash(keyHash)

		if oldExists && newExists && oldNode != newNode {
			// Data needs to move
			info := RebalanceInfo{
				FromNode: oldNode,
				ToNode:   newNode,
				KeyRange: KeyRange{
					Start: keyHash,
					End:   keyHash,
				},
				KeyCount: 1, // Estimated
			}
			rebalanceInfo = append(rebalanceInfo, info)
		}
	}

	return rebalanceInfo
}

// getNodeForHash returns the node for a given hash (internal method)
func (ch *ConsistentHash) getNodeForHash(hash uint32) (string, bool) {
	if len(ch.sortedHashes) == 0 {
		return "", false
	}

	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	if idx == len(ch.sortedHashes) {
		idx = 0
	}

	nodeHash := ch.sortedHashes[idx]
	node := ch.hashRing[nodeHash]

	return node, true
}
