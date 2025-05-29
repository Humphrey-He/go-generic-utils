// Copyright 2024 Humphrey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tree 提供了高性能树结构的实现，适用于电商平台后端开发
package tree

import (
	"errors"
	"sync"
)

// 常用错误定义
var (
	ErrKeyNotFound        = errors.New("ggu: 键不存在")
	ErrKeyExists          = errors.New("ggu: 键已存在")
	ErrEmptyTree          = errors.New("ggu: 树为空")
	ErrNilComparator      = errors.New("ggu: 比较器不能为nil")
	ErrInvalidRange       = errors.New("ggu: 无效的范围")
	ErrConcurrentModified = errors.New("ggu: 并发修改错误")
)

// Comparator 用于比较键的大小
// 返回值:
// -1: k1 < k2
//
//	0: k1 = k2
//	1: k1 > k2
type Comparator[K any] func(k1, k2 K) int

// avlNode AVL树节点
type avlNode[K any, V any] struct {
	Key      K
	Value    V
	Height   int
	Left     *avlNode[K, V]
	Right    *avlNode[K, V]
	Parent   *avlNode[K, V] // 父节点引用，用于高效迭代
	Modified int64          // 修改计数，用于并发检测
}

// AVLTree 是一个自平衡二叉搜索树
// AVL树在查找、插入和删除操作上都具有O(log n)的时间复杂度
// 对于频繁查询操作的电商场景(如商品目录、价格查询等)非常适合
type AVLTree[K any, V any] struct {
	root       *avlNode[K, V]
	size       int
	comparator Comparator[K]
	mu         sync.RWMutex // 用于并发安全操作
	modified   int64        // 修改计数，用于并发检测
}

// NewAVLTree 创建一个新的AVL树
func NewAVLTree[K any, V any](comparator Comparator[K]) (*AVLTree[K, V], error) {
	if comparator == nil {
		return nil, ErrNilComparator
	}
	return &AVLTree[K, V]{
		comparator: comparator,
	}, nil
}

// 获取节点高度
func height[K any, V any](node *avlNode[K, V]) int {
	if node == nil {
		return -1
	}
	return node.Height
}

// 计算平衡因子
func balanceFactor[K any, V any](node *avlNode[K, V]) int {
	if node == nil {
		return 0
	}
	return height(node.Left) - height(node.Right)
}

// 更新节点高度
func updateHeight[K any, V any](node *avlNode[K, V]) {
	leftHeight := height(node.Left)
	rightHeight := height(node.Right)
	if leftHeight > rightHeight {
		node.Height = leftHeight + 1
	} else {
		node.Height = rightHeight + 1
	}
}

// 右旋转操作
func rightRotate[K any, V any](y *avlNode[K, V]) *avlNode[K, V] {
	x := y.Left
	T2 := x.Right

	// 执行旋转
	x.Right = y
	y.Left = T2

	// 更新父节点引用
	x.Parent = y.Parent
	y.Parent = x
	if T2 != nil {
		T2.Parent = y
	}

	// 更新高度
	updateHeight(y)
	updateHeight(x)

	return x
}

// 左旋转操作
func leftRotate[K any, V any](x *avlNode[K, V]) *avlNode[K, V] {
	y := x.Right
	T2 := y.Left

	// 执行旋转
	y.Left = x
	x.Right = T2

	// 更新父节点引用
	y.Parent = x.Parent
	x.Parent = y
	if T2 != nil {
		T2.Parent = x
	}

	// 更新高度
	updateHeight(x)
	updateHeight(y)

	return y
}

// 平衡节点
func balance[K any, V any](node *avlNode[K, V]) *avlNode[K, V] {
	if node == nil {
		return nil
	}

	updateHeight(node)

	// 获取平衡因子
	balance := balanceFactor(node)

	// 左子树高 - 左左情况
	if balance > 1 && balanceFactor(node.Left) >= 0 {
		return rightRotate(node)
	}

	// 左子树高 - 左右情况
	if balance > 1 && balanceFactor(node.Left) < 0 {
		node.Left = leftRotate(node.Left)
		return rightRotate(node)
	}

	// 右子树高 - 右右情况
	if balance < -1 && balanceFactor(node.Right) <= 0 {
		return leftRotate(node)
	}

	// 右子树高 - 右左情况
	if balance < -1 && balanceFactor(node.Right) > 0 {
		node.Right = rightRotate(node.Right)
		return leftRotate(node)
	}

	// 已经平衡
	return node
}

// Put 插入或更新键值对
func (t *AVLTree[K, V]) Put(key K, value V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.modified++
	t.root = t.put(t.root, nil, key, value)
}

// put 插入新节点的内部递归方法
func (t *AVLTree[K, V]) put(node *avlNode[K, V], parent *avlNode[K, V], key K, value V) *avlNode[K, V] {
	// 创建新节点
	if node == nil {
		t.size++
		return &avlNode[K, V]{
			Key:      key,
			Value:    value,
			Height:   0,
			Parent:   parent,
			Modified: t.modified,
		}
	}

	// 比较键，决定插入左子树还是右子树
	cmp := t.comparator(key, node.Key)
	if cmp < 0 {
		// 键小于当前节点，插入左子树
		node.Left = t.put(node.Left, node, key, value)
	} else if cmp > 0 {
		// 键大于当前节点，插入右子树
		node.Right = t.put(node.Right, node, key, value)
	} else {
		// 键已存在，更新值
		node.Value = value
		node.Modified = t.modified
		return node
	}

	// 更新当前节点信息并平衡
	node.Modified = t.modified
	return balance(node)
}

// Get 获取指定键的值
func (t *AVLTree[K, V]) Get(key K) (V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.findNode(key)
	if node == nil {
		var zero V
		return zero, ErrKeyNotFound
	}
	return node.Value, nil
}

// findNode 查找具有特定键的节点
func (t *AVLTree[K, V]) findNode(key K) *avlNode[K, V] {
	current := t.root
	for current != nil {
		cmp := t.comparator(key, current.Key)
		if cmp < 0 {
			current = current.Left
		} else if cmp > 0 {
			current = current.Right
		} else {
			return current
		}
	}
	return nil
}

// Contains 检查树中是否包含特定键
func (t *AVLTree[K, V]) Contains(key K) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.findNode(key) != nil
}

// Remove 删除指定键的节点
func (t *AVLTree[K, V]) Remove(key K) (V, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.findNode(key)
	if node == nil {
		var zero V
		return zero, ErrKeyNotFound
	}

	value := node.Value
	t.modified++
	t.root = t.remove(t.root, key)
	return value, nil
}

// remove 删除节点的内部递归方法
func (t *AVLTree[K, V]) remove(node *avlNode[K, V], key K) *avlNode[K, V] {
	if node == nil {
		return nil
	}

	// 查找要删除的节点
	cmp := t.comparator(key, node.Key)
	if cmp < 0 {
		// 在左子树中继续查找
		node.Left = t.remove(node.Left, key)
		if node.Left != nil {
			node.Left.Parent = node
		}
	} else if cmp > 0 {
		// 在右子树中继续查找
		node.Right = t.remove(node.Right, key)
		if node.Right != nil {
			node.Right.Parent = node
		}
	} else {
		// 找到要删除的节点
		t.size--

		// 情况1: 叶子节点或只有一个子节点
		if node.Left == nil {
			return node.Right
		} else if node.Right == nil {
			return node.Left
		}

		// 情况2: 有两个子节点
		// 找到右子树中的最小节点(后继)
		successor := t.findMin(node.Right)

		// 使用后继节点的键值替代当前节点
		node.Key = successor.Key
		node.Value = successor.Value
		node.Modified = t.modified

		// 删除后继节点
		node.Right = t.remove(node.Right, successor.Key)
		if node.Right != nil {
			node.Right.Parent = node
		}
	}

	// 更新高度并重新平衡
	node.Modified = t.modified
	return balance(node)
}

// findMin 找到以node为根的子树中的最小节点
func (t *AVLTree[K, V]) findMin(node *avlNode[K, V]) *avlNode[K, V] {
	current := node
	for current != nil && current.Left != nil {
		current = current.Left
	}
	return current
}

// findMax 找到以node为根的子树中的最大节点
func (t *AVLTree[K, V]) findMax(node *avlNode[K, V]) *avlNode[K, V] {
	current := node
	for current != nil && current.Right != nil {
		current = current.Right
	}
	return current
}

// Size 返回树中的节点数量
func (t *AVLTree[K, V]) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// IsEmpty 检查树是否为空
func (t *AVLTree[K, V]) IsEmpty() bool {
	return t.Size() == 0
}

// Clear 清空树
func (t *AVLTree[K, V]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = nil
	t.size = 0
	t.modified++
}

// Keys 返回所有键的有序切片
func (t *AVLTree[K, V]) Keys() []K {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]K, 0, t.size)
	t.inOrderTraversal(t.root, func(k K, v V) bool {
		result = append(result, k)
		return true
	})
	return result
}

// Values 返回与键对应的所有值的切片
func (t *AVLTree[K, V]) Values() []V {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]V, 0, t.size)
	t.inOrderTraversal(t.root, func(k K, v V) bool {
		result = append(result, v)
		return true
	})
	return result
}

// KeyValues 返回所有键值对的有序切片
func (t *AVLTree[K, V]) KeyValues() ([]K, []V) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	keys := make([]K, 0, t.size)
	values := make([]V, 0, t.size)

	t.inOrderTraversal(t.root, func(k K, v V) bool {
		keys = append(keys, k)
		values = append(values, v)
		return true
	})

	return keys, values
}

// ForEach 对树中的每个节点按顺序执行指定函数
// 如果函数返回false，则停止遍历
func (t *AVLTree[K, V]) ForEach(fn func(key K, value V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.inOrderTraversal(t.root, fn)
}

// inOrderTraversal 中序遍历树
func (t *AVLTree[K, V]) inOrderTraversal(node *avlNode[K, V], fn func(key K, value V) bool) bool {
	if node == nil {
		return true
	}

	// 先遍历左子树
	if !t.inOrderTraversal(node.Left, fn) {
		return false
	}

	// 处理当前节点
	if !fn(node.Key, node.Value) {
		return false
	}

	// 再遍历右子树
	return t.inOrderTraversal(node.Right, fn)
}

// Iterator 返回一个迭代器，用于遍历树中的元素
func (t *AVLTree[K, V]) Iterator() *Iterator[K, V] {
	t.mu.RLock()
	defer t.mu.RUnlock()

	iter := &Iterator[K, V]{
		tree:     t,
		modified: t.modified,
	}

	// 初始化为最小节点
	if t.root != nil {
		iter.current = t.findMin(t.root)
	}

	return iter
}

// Iterator 是树的迭代器
type Iterator[K any, V any] struct {
	tree     *AVLTree[K, V]
	current  *avlNode[K, V]
	modified int64
}

// HasNext 返回迭代器是否有下一个元素
func (it *Iterator[K, V]) HasNext() bool {
	return it.current != nil
}

// Next 返回下一个键值对，并将迭代器向前移动
func (it *Iterator[K, V]) Next() (K, V, error) {
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()

	// 检查并发修改
	if it.modified != it.tree.modified {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, ErrConcurrentModified
	}

	if it.current == nil {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, ErrKeyNotFound
	}

	key, value := it.current.Key, it.current.Value
	it.advance()
	return key, value, nil
}

// advance 将迭代器向前移动到下一个节点
func (it *Iterator[K, V]) advance() {
	if it.current == nil {
		return
	}

	// 如果有右子树，则下一个节点是右子树中的最小节点
	if it.current.Right != nil {
		it.current = it.tree.findMin(it.current.Right)
		return
	}

	// 否则，找到第一个祖先节点，其左子树包含当前节点
	parent := it.current.Parent
	for parent != nil && it.current == parent.Right {
		it.current = parent
		parent = parent.Parent
	}
	it.current = parent
}

// FindRange 查找键在指定范围内的所有值
// fromKey: 起始键(包含)
// toKey: 结束键(不包含)
func (t *AVLTree[K, V]) FindRange(fromKey, toKey K) ([]K, []V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 验证范围
	if t.comparator(fromKey, toKey) >= 0 {
		return nil, nil, ErrInvalidRange
	}

	keys := make([]K, 0)
	values := make([]V, 0)

	t.inOrderTraversal(t.root, func(k K, v V) bool {
		if t.comparator(k, fromKey) >= 0 && t.comparator(k, toKey) < 0 {
			keys = append(keys, k)
			values = append(values, v)
		}
		// 如果已经超过了上限，可以提前终止遍历
		return t.comparator(k, toKey) < 0
	})

	return keys, values, nil
}

// Min 返回树中的最小键及其对应的值
func (t *AVLTree[K, V]) Min() (K, V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, ErrEmptyTree
	}

	min := t.findMin(t.root)
	return min.Key, min.Value, nil
}

// Max 返回树中的最大键及其对应的值
func (t *AVLTree[K, V]) Max() (K, V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, ErrEmptyTree
	}

	max := t.findMax(t.root)
	return max.Key, max.Value, nil
}

// Height 返回树的高度
func (t *AVLTree[K, V]) Height() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return height(t.root) + 1
}

// --- 电商场景特定方法 ---

// GetOrDefault 获取键对应的值，如果键不存在则返回默认值
// 这对于获取商品属性、用户配置等场景很有用
func (t *AVLTree[K, V]) GetOrDefault(key K, defaultValue V) V {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.findNode(key)
	if node == nil {
		return defaultValue
	}
	return node.Value
}

// PutIfAbsent 仅当键不存在时插入值
// 返回值表示是否进行了插入操作
// 适用于防止重复创建订单、用户等场景
func (t *AVLTree[K, V]) PutIfAbsent(key K, value V) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.findNode(key) != nil {
		return false
	}

	t.modified++
	t.root = t.put(t.root, nil, key, value)
	return true
}

// ComputeIfPresent 当键存在时，使用指定函数计算新值
// 如果函数返回的新值不为nil，则更新值；否则删除该键
// 适用于购物车数量调整、库存变更等场景
func (t *AVLTree[K, V]) ComputeIfPresent(key K, remappingFunction func(key K, oldValue V) (V, bool)) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.findNode(key)
	if node == nil {
		var zero V
		return zero, false
	}

	newValue, keep := remappingFunction(key, node.Value)
	if !keep {
		// 删除节点
		t.modified++
		t.root = t.remove(t.root, key)
		return newValue, true
	}

	// 更新节点值
	node.Value = newValue
	node.Modified = t.modified
	t.modified++
	return newValue, true
}

// BatchInsert 批量插入键值对
// 适用于批量导入商品、批量更新价格等场景
func (t *AVLTree[K, V]) BatchInsert(pairs []struct {
	Key   K
	Value V
}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.modified++
	for _, pair := range pairs {
		t.root = t.put(t.root, nil, pair.Key, pair.Value)
	}
}

// FindPrefix 查找所有以指定前缀开头的键值对
// 这需要键类型支持获取前缀的能力，例如字符串
// 适用于搜索补全、商品分类查询等场景
func (t *AVLTree[K, V]) FindPrefix(prefix K, isPrefixOf func(prefix, key K) bool) ([]K, []V) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	keys := make([]K, 0)
	values := make([]V, 0)

	t.inOrderTraversal(t.root, func(k K, v V) bool {
		if isPrefixOf(prefix, k) {
			keys = append(keys, k)
			values = append(values, v)
		}
		return true
	})

	return keys, values
}

// 创建字符串键的比较器
func StringComparator(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// 创建整数键的比较器
func IntComparator(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// 创建浮点数键的比较器
func Float64Comparator(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
