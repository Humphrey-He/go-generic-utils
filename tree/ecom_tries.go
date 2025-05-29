// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tree

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// -- 前缀树(Trie) --

// TrieNode 前缀树节点
type TrieNode struct {
	Children map[rune]*TrieNode
	IsEnd    bool
	Data     interface{}
}

// Trie 前缀树
// 适用于自动补全、搜索推荐等场景
type Trie struct {
	root *TrieNode
	mu   sync.RWMutex
	size int
}

// NewTrie 创建新的前缀树
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			Children: make(map[rune]*TrieNode),
		},
	}
}

// Insert 插入单词
func (t *Trie) Insert(word string, data interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, r := range word {
		if _, exists := node.Children[r]; !exists {
			node.Children[r] = &TrieNode{
				Children: make(map[rune]*TrieNode),
			}
		}
		node = node.Children[r]
	}

	// 如果是新单词，增加计数
	if !node.IsEnd {
		t.size++
	}

	node.IsEnd = true
	node.Data = data
}

// Search 查找完全匹配的单词
func (t *Trie) Search(word string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, r := range word {
		child, exists := node.Children[r]
		if !exists {
			return nil, false
		}
		node = child
	}

	return node.Data, node.IsEnd
}

// StartsWith 查询是否有以prefix开头的单词
func (t *Trie) StartsWith(prefix string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, r := range prefix {
		child, exists := node.Children[r]
		if !exists {
			return false
		}
		node = child
	}

	return true
}

// Autocomplete 获取所有以prefix开头的单词
func (t *Trie) Autocomplete(prefix string, limit int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 先找到前缀对应的节点
	node := t.root
	for _, r := range prefix {
		child, exists := node.Children[r]
		if !exists {
			return []string{}
		}
		node = child
	}

	// 收集所有以该节点为起点的单词
	result := make([]string, 0)
	t.collect(node, prefix, &result, limit)

	return result
}

// collect 递归收集以node为起点的所有单词
func (t *Trie) collect(node *TrieNode, prefix string, result *[]string, limit int) {
	if node.IsEnd {
		*result = append(*result, prefix)
	}

	// 如果已经达到限制，停止收集
	if limit > 0 && len(*result) >= limit {
		return
	}

	// 按字典序遍历子节点
	for r, child := range node.Children {
		t.collect(child, prefix+string(r), result, limit)

		// 如果已经达到限制，停止收集
		if limit > 0 && len(*result) >= limit {
			return
		}
	}
}

// Delete 删除单词
func (t *Trie) Delete(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.deleteHelper(t.root, word, 0)
}

// deleteHelper 递归删除单词
func (t *Trie) deleteHelper(node *TrieNode, word string, depth int) bool {
	// 已处理完整个单词
	if depth == len(word) {
		// 如果单词不存在，返回false
		if !node.IsEnd {
			return false
		}

		// 标记为非单词结尾
		node.IsEnd = false
		node.Data = nil
		t.size--

		// 如果节点没有子节点，可以删除
		return len(node.Children) == 0
	}

	// 获取当前字符
	r := rune(word[depth])
	child, exists := node.Children[r]

	// 如果字符不存在，单词不存在
	if !exists {
		return false
	}

	// 递归删除子节点
	shouldDeleteChild := t.deleteHelper(child, word, depth+1)

	// 如果应该删除子节点
	if shouldDeleteChild {
		delete(node.Children, r)

		// 如果当前节点不是单词结尾且没有其他子节点，可以删除
		return !node.IsEnd && len(node.Children) == 0
	}

	return false
}

// Size 返回Trie中单词的数量
func (t *Trie) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// Clear 清空Trie
func (t *Trie) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.root = &TrieNode{
		Children: make(map[rune]*TrieNode),
	}
	t.size = 0
}

// -- 缓存树 --

// CacheNode 缓存树节点
type CacheNode[T any] struct {
	Key       string
	Value     T
	ExpiresAt time.Time
	Children  map[string]*CacheNode[T]
}

// CacheTree 层次化缓存树
// 适用于需要层次结构的缓存场景，如商品分类缓存、地区缓存等
type CacheTree[T any] struct {
	root       *CacheNode[T]
	mu         sync.RWMutex
	defaultTTL time.Duration
}

// NewCacheTree 创建新的缓存树
func NewCacheTree[T any](defaultTTL time.Duration) *CacheTree[T] {
	return &CacheTree[T]{
		root: &CacheNode[T]{
			Key:      "root",
			Children: make(map[string]*CacheNode[T]),
		},
		defaultTTL: defaultTTL,
	}
}

// Put 将值放入缓存
func (ct *CacheTree[T]) Put(path string, value T) {
	ct.PutWithTTL(path, value, ct.defaultTTL)
}

// PutWithTTL 将值放入缓存，指定TTL
func (ct *CacheTree[T]) PutWithTTL(path string, value T, ttl time.Duration) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	parts := strings.Split(path, "/")
	node := ct.root

	// 遍历路径，创建或获取节点
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		if _, exists := node.Children[part]; !exists {
			node.Children[part] = &CacheNode[T]{
				Key:      part,
				Children: make(map[string]*CacheNode[T]),
			}
		}

		node = node.Children[part]
	}

	// 获取最后一个部分作为键
	lastPart := parts[len(parts)-1]
	if lastPart == "" {
		return
	}

	// 创建或更新节点
	expiresAt := time.Now().Add(ttl)
	if existing, exists := node.Children[lastPart]; exists {
		existing.Value = value
		existing.ExpiresAt = expiresAt
	} else {
		node.Children[lastPart] = &CacheNode[T]{
			Key:       lastPart,
			Value:     value,
			ExpiresAt: expiresAt,
			Children:  make(map[string]*CacheNode[T]),
		}
	}
}

// Get 从缓存获取值
func (ct *CacheTree[T]) Get(path string) (T, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	parts := strings.Split(path, "/")
	node := ct.root

	// 遍历路径查找节点
	for i, part := range parts {
		if part == "" {
			continue
		}

		child, exists := node.Children[part]
		if !exists {
			var zero T
			return zero, errors.New("ggu: 缓存项不存在")
		}

		// 如果是最后一部分，检查是否过期并返回值
		if i == len(parts)-1 {
			if !child.ExpiresAt.IsZero() && time.Now().After(child.ExpiresAt) {
				// 过期了，移除它
				delete(node.Children, part)
				var zero T
				return zero, ErrExpired
			}
			return child.Value, nil
		}

		node = child
	}

	var zero T
	return zero, errors.New("ggu: 无效的缓存路径")
}

// Delete 删除缓存项
func (ct *CacheTree[T]) Delete(path string) bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	parts := strings.Split(path, "/")
	node := ct.root

	// 遍历路径找到父节点
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		child, exists := node.Children[part]
		if !exists {
			return false
		}

		node = child
	}

	// 删除最后一个部分对应的节点
	lastPart := parts[len(parts)-1]
	if lastPart == "" {
		return false
	}

	if _, exists := node.Children[lastPart]; exists {
		delete(node.Children, lastPart)
		return true
	}

	return false
}

// GetChildren 获取指定路径下的所有子项
func (ct *CacheTree[T]) GetChildren(path string) map[string]T {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// 先找到指定路径的节点
	node := ct.findNode(path)
	if node == nil {
		return make(map[string]T)
	}

	// 收集子项
	result := make(map[string]T)
	now := time.Now()

	for key, child := range node.Children {
		// 跳过过期的项
		if !child.ExpiresAt.IsZero() && now.After(child.ExpiresAt) {
			continue
		}

		result[key] = child.Value
	}

	return result
}

// 查找指定路径的节点
func (ct *CacheTree[T]) findNode(path string) *CacheNode[T] {
	parts := strings.Split(path, "/")
	node := ct.root

	for _, part := range parts {
		if part == "" {
			continue
		}

		child, exists := node.Children[part]
		if !exists {
			return nil
		}

		node = child
	}

	return node
}

// Cleanup 清理过期的缓存项
func (ct *CacheTree[T]) Cleanup() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	removed := 0
	removed += ct.cleanupNode(ct.root)
	return removed
}

// 清理节点及其子节点
func (ct *CacheTree[T]) cleanupNode(node *CacheNode[T]) int {
	if node == nil {
		return 0
	}

	removed := 0
	now := time.Now()

	// 创建待删除键的列表
	keysToDelete := make([]string, 0)

	for key, child := range node.Children {
		// 检查是否过期
		if !child.ExpiresAt.IsZero() && now.After(child.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
			removed++
			continue
		}

		// 递归清理子节点
		removed += ct.cleanupNode(child)
	}

	// 删除过期的键
	for _, key := range keysToDelete {
		delete(node.Children, key)
	}

	return removed
}

// -- 搜索引擎 --

// SearchEngine 搜索引擎
// 使用Trie树实现高效的前缀搜索
type SearchEngine struct {
	prefixTrie     *Trie               // 前缀树，用于自动完成
	productMapping map[string][]string // 词 -> 商品ID映射
	termFrequency  map[string]int      // 词频统计
	mu             sync.RWMutex
}

// NewSearchEngine 创建新的搜索引擎
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		prefixTrie:     NewTrie(),
		productMapping: make(map[string][]string),
		termFrequency:  make(map[string]int),
	}
}

// IndexProduct 为商品创建搜索索引
func (se *SearchEngine) IndexProduct(productID string, terms []string) {
	se.mu.Lock()
	defer se.mu.Unlock()

	for _, term := range terms {
		if term == "" {
			continue
		}

		// 添加到前缀树
		se.prefixTrie.Insert(term, productID)

		// 更新词频
		se.termFrequency[term]++

		// 更新商品映射
		ids, exists := se.productMapping[term]
		if !exists {
			se.productMapping[term] = []string{productID}
		} else {
			// 检查是否已存在
			found := false
			for _, id := range ids {
				if id == productID {
					found = true
					break
				}
			}

			if !found {
				se.productMapping[term] = append(ids, productID)
			}
		}
	}
}

// Search 搜索商品
func (se *SearchEngine) Search(query string, limit int) []SearchResult {
	se.mu.RLock()
	defer se.mu.RUnlock()

	if query == "" || limit <= 0 {
		return []SearchResult{}
	}

	// 找到所有匹配的前缀
	matchingTerms := se.prefixTrie.Autocomplete(query, 10)

	// 为每个匹配项创建结果
	results := make([]SearchResult, 0, len(matchingTerms))

	for _, term := range matchingTerms {
		productIDs, exists := se.productMapping[term]
		if exists {
			results = append(results, SearchResult{
				Term:       term,
				ProductIDs: productIDs,
				Score:      se.termFrequency[term], // 简单地使用词频作为分数
			})
		}
	}

	// 按分数排序
	sortSearchResults(results)

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// AutoComplete 自动补全
func (se *SearchEngine) AutoComplete(prefix string, limit int) []string {
	se.mu.RLock()
	defer se.mu.RUnlock()

	if prefix == "" || limit <= 0 {
		return []string{}
	}

	return se.prefixTrie.Autocomplete(prefix, limit)
}

// GetTopSearchTerms 获取热门搜索词
func (se *SearchEngine) GetTopSearchTerms(limit int) []string {
	se.mu.RLock()
	defer se.mu.RUnlock()

	if limit <= 0 {
		return []string{}
	}

	// 创建词-频率对
	pairs := make([]TermFreq, 0, len(se.termFrequency))
	for term, freq := range se.termFrequency {
		pairs = append(pairs, TermFreq{term, freq})
	}

	// 按频率排序
	sortTermsByFrequency(pairs)

	// 提取前N个词
	result := make([]string, 0, limit)
	for i := 0; i < limit && i < len(pairs); i++ {
		result = append(result, pairs[i].Term)
	}

	return result
}
