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
	"sync"
)

var (
	// ErrInvalidDegree B树最小度数无效
	ErrInvalidDegree = errors.New("ggu: B树最小度数必须>=2")
)

// BTreeNode B树节点
type bTreeNode[K comparable, V any] struct {
	keys     []K                // 键数组
	values   []V                // 值数组
	children []*bTreeNode[K, V] // 子节点数组
	leaf     bool               // 是否为叶子节点
}

// BTree B树实现
// B树特别适合大数据量、高并发的电商场景
// 例如：商品目录、库存管理、订单系统等
type BTree[K comparable, V any] struct {
	root       *bTreeNode[K, V] // 根节点
	degree     int              // 最小度数(每个非根节点至少有degree-1个键)
	comparator Comparator[K]    // 键比较器
	size       int              // 树中键的数量
	mu         sync.RWMutex     // 读写锁，用于并发访问控制
}

// NewBTree 创建一个新的B树
// degree: 最小度数，决定了节点中键的最小和最大数量
// 每个节点(除根节点外)必须至少有degree-1个键，最多有2*degree-1个键
func NewBTree[K comparable, V any](degree int, comparator Comparator[K]) (*BTree[K, V], error) {
	if degree < 2 {
		return nil, ErrInvalidDegree
	}

	if comparator == nil {
		return nil, ErrNilComparator
	}

	return &BTree[K, V]{
		root: &bTreeNode[K, V]{
			keys:     make([]K, 0),
			values:   make([]V, 0),
			children: make([]*bTreeNode[K, V], 0),
			leaf:     true,
		},
		degree:     degree,
		comparator: comparator,
		size:       0,
	}, nil
}

// Put 插入或更新键值对
func (t *BTree[K, V]) Put(key K, value V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	root := t.root
	maxKeys := 2*t.degree - 1

	// 如果根节点已满，需要分裂
	if len(root.keys) == maxKeys {
		newRoot := &bTreeNode[K, V]{
			keys:     make([]K, 0),
			values:   make([]V, 0),
			children: []*bTreeNode[K, V]{root},
			leaf:     false,
		}
		t.root = newRoot
		t.splitChild(newRoot, 0)
		t.insertNonFull(newRoot, key, value)
	} else {
		t.insertNonFull(root, key, value)
	}
}

// 分裂已满的子节点
func (t *BTree[K, V]) splitChild(parent *bTreeNode[K, V], index int) {
	degree := t.degree
	child := parent.children[index]

	// 创建新节点，将分裂节点的右半部分移入
	newChild := &bTreeNode[K, V]{
		keys:   make([]K, degree-1),
		values: make([]V, degree-1),
		leaf:   child.leaf,
	}

	// 复制原节点右半部分的键和值到新节点
	copy(newChild.keys, child.keys[degree:])
	copy(newChild.values, child.values[degree:])

	// 如果不是叶子节点，还需要移动子节点指针
	if !child.leaf {
		newChild.children = make([]*bTreeNode[K, V], degree)
		copy(newChild.children, child.children[degree:])
		// 截断原节点的子节点数组
		child.children = child.children[:degree]
	}

	// 将中间的键和值提升到父节点
	midKey := child.keys[degree-1]
	midValue := child.values[degree-1]

	// 调整原节点，移除已经分离出去的部分
	child.keys = child.keys[:degree-1]
	child.values = child.values[:degree-1]

	// 在父节点的适当位置插入中间键和指向新节点的指针
	parent.keys = append(parent.keys, midKey)       // 先添加到末尾
	parent.values = append(parent.values, midValue) // 先添加到末尾
	parent.children = append(parent.children, nil)  // 扩展子节点数组

	// 将末尾添加的键值对移动到正确位置
	i := len(parent.keys) - 1
	for i > index {
		parent.keys[i] = parent.keys[i-1]
		parent.values[i] = parent.values[i-1]
		parent.children[i+1] = parent.children[i]
		i--
	}

	parent.keys[index] = midKey
	parent.values[index] = midValue
	parent.children[index+1] = newChild
}

// 向非满节点插入键值对
func (t *BTree[K, V]) insertNonFull(node *bTreeNode[K, V], key K, value V) {
	i := len(node.keys) - 1

	// 如果是叶子节点，直接插入
	if node.leaf {
		// 找到插入位置
		for i >= 0 && t.comparator(key, node.keys[i]) < 0 {
			i--
		}

		// 检查是否为更新操作
		if i >= 0 && t.comparator(key, node.keys[i]) == 0 {
			node.values[i] = value // 更新值
			return
		}

		// 插入新键值对
		node.keys = append(node.keys, key)       // 先添加到末尾
		node.values = append(node.values, value) // 先添加到末尾

		// 移动到正确位置
		for j := len(node.keys) - 1; j > i+1; j-- {
			node.keys[j] = node.keys[j-1]
			node.values[j] = node.values[j-1]
		}
		node.keys[i+1] = key
		node.values[i+1] = value
		t.size++
	} else {
		// 非叶子节点，需要找到合适的子节点继续插入
		for i >= 0 && t.comparator(key, node.keys[i]) < 0 {
			i--
		}

		// 检查是否为更新操作
		if i >= 0 && t.comparator(key, node.keys[i]) == 0 {
			node.values[i] = value // 更新值
			return
		}

		childIndex := i + 1

		// 如果子节点已满，先分裂
		if len(node.children[childIndex].keys) == 2*t.degree-1 {
			t.splitChild(node, childIndex)

			// 分裂后可能改变了插入位置
			if t.comparator(key, node.keys[childIndex]) > 0 {
				childIndex++
			} else if t.comparator(key, node.keys[childIndex]) == 0 {
				node.values[childIndex] = value // 更新值
				return
			}
		}

		// 递归插入到合适的子节点
		t.insertNonFull(node.children[childIndex], key, value)
	}
}

// Get 获取指定键的值
func (t *BTree[K, V]) Get(key K) (V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.search(t.root, key)
}

// 在节点中搜索键
func (t *BTree[K, V]) search(node *bTreeNode[K, V], key K) (V, error) {
	i := 0
	// 查找键的位置
	for i < len(node.keys) && t.comparator(key, node.keys[i]) > 0 {
		i++
	}

	// 找到了键
	if i < len(node.keys) && t.comparator(key, node.keys[i]) == 0 {
		return node.values[i], nil
	}

	// 键不在当前节点，且当前节点是叶子节点，说明键不存在
	if node.leaf {
		var zero V
		return zero, ErrKeyNotFound
	}

	// 递归搜索合适的子节点
	return t.search(node.children[i], key)
}

// Contains 检查是否包含指定的键
func (t *BTree[K, V]) Contains(key K) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, err := t.search(t.root, key)
	return err == nil
}

// Remove 删除指定的键
func (t *BTree[K, V]) Remove(key K) (V, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.root == nil || len(t.root.keys) == 0 {
		var zero V
		return zero, ErrKeyNotFound
	}

	value, err := t.search(t.root, key)
	if err != nil {
		var zero V
		return zero, err
	}

	t.remove(t.root, key)

	// 如果根节点变空且不是叶子节点，更新根节点
	if len(t.root.keys) == 0 && !t.root.leaf {
		t.root = t.root.children[0]
	}

	t.size--
	return value, nil
}

// 删除节点中的键
func (t *BTree[K, V]) remove(node *bTreeNode[K, V], key K) {
	minKeys := t.degree - 1

	// 查找键的位置
	i := 0
	for i < len(node.keys) && t.comparator(key, node.keys[i]) > 0 {
		i++
	}

	// 如果找到键且在当前节点
	if i < len(node.keys) && t.comparator(key, node.keys[i]) == 0 {
		// 情况1: 叶子节点，直接删除
		if node.leaf {
			// 移除键和值
			node.keys = append(node.keys[:i], node.keys[i+1:]...)
			node.values = append(node.values[:i], node.values[i+1:]...)
			return
		}

		// 情况2: 内部节点

		// 2a: 如果左子节点有足够多的键，找前驱
		if len(node.children[i].keys) > minKeys {
			// 找到前驱
			pred, predVal := t.findPredecessor(node, i)
			// 用前驱替换当前键
			node.keys[i] = pred
			node.values[i] = predVal
			// 在左子节点中递归删除前驱
			t.remove(node.children[i], pred)
			return
		}

		// 2b: 如果右子节点有足够多的键，找后继
		if len(node.children[i+1].keys) > minKeys {
			// 找到后继
			succ, succVal := t.findSuccessor(node, i)
			// 用后继替换当前键
			node.keys[i] = succ
			node.values[i] = succVal
			// 在右子节点中递归删除后继
			t.remove(node.children[i+1], succ)
			return
		}

		// 2c: 左右子节点都没有足够多的键，合并节点
		t.mergeNodes(node, i)
		// 在合并后的节点中递归删除
		t.remove(node.children[i], key)
		return
	}

	// 键不在当前节点

	// 如果是叶子节点，键不存在
	if node.leaf {
		return
	}

	// 检查子节点是否有足够的键
	if len(node.children[i].keys) <= minKeys {
		t.fillChild(node, i)
	}

	// 如果最后一个子节点被合并，需要调整索引
	if i > len(node.keys) {
		i--
	}

	// 递归到合适的子节点删除
	t.remove(node.children[i], key)
}

// 找到节点中键的前驱
func (t *BTree[K, V]) findPredecessor(node *bTreeNode[K, V], index int) (K, V) {
	current := node.children[index]

	// 向右下方遍历到最右叶子节点
	for !current.leaf {
		current = current.children[len(current.children)-1]
	}

	// 返回最右叶子节点的最后一个键
	lastIndex := len(current.keys) - 1
	return current.keys[lastIndex], current.values[lastIndex]
}

// 找到节点中键的后继
func (t *BTree[K, V]) findSuccessor(node *bTreeNode[K, V], index int) (K, V) {
	current := node.children[index+1]

	// 向左下方遍历到最左叶子节点
	for !current.leaf {
		current = current.children[0]
	}

	// 返回最左叶子节点的第一个键
	return current.keys[0], current.values[0]
}

// 合并索引i和i+1的两个子节点
func (t *BTree[K, V]) mergeNodes(node *bTreeNode[K, V], index int) {
	child := node.children[index]
	sibling := node.children[index+1]

	// 将父节点中的键和值移到左子节点
	child.keys = append(child.keys, node.keys[index])
	child.values = append(child.values, node.values[index])

	// 将右子节点的所有键和值附加到左子节点
	child.keys = append(child.keys, sibling.keys...)
	child.values = append(child.values, sibling.values...)

	// 如果不是叶子节点，还需要移动子节点指针
	if !child.leaf {
		child.children = append(child.children, sibling.children...)
	}

	// 从父节点中删除键和右子节点
	node.keys = append(node.keys[:index], node.keys[index+1:]...)
	node.values = append(node.values[:index], node.values[index+1:]...)
	node.children = append(node.children[:index+1], node.children[index+2:]...)
}

// 确保子节点有足够多的键
func (t *BTree[K, V]) fillChild(node *bTreeNode[K, V], index int) {
	minKeys := t.degree - 1

	// 尝试从左兄弟节点借一个键
	if index > 0 && len(node.children[index-1].keys) > minKeys {
		t.borrowFromPrev(node, index)
		return
	}

	// 尝试从右兄弟节点借一个键
	if index < len(node.keys) && len(node.children[index+1].keys) > minKeys {
		t.borrowFromNext(node, index)
		return
	}

	// 如果无法借用，需要合并节点
	if index < len(node.keys) {
		t.mergeNodes(node, index)
	} else {
		t.mergeNodes(node, index-1)
	}
}

// 从前一个兄弟节点借一个键
func (t *BTree[K, V]) borrowFromPrev(node *bTreeNode[K, V], index int) {
	child := node.children[index]
	sibling := node.children[index-1]

	// 为从父节点下移的键腾出空间
	child.keys = append([]K{node.keys[index-1]}, child.keys...)
	child.values = append([]V{node.values[index-1]}, child.values...)

	// 如果不是叶子节点，需要移动子节点指针
	if !child.leaf {
		// 移动兄弟节点的最右子节点
		child.children = append([]*bTreeNode[K, V]{sibling.children[len(sibling.children)-1]}, child.children...)
		// 从兄弟节点中移除该子节点
		sibling.children = sibling.children[:len(sibling.children)-1]
	}

	// 将兄弟节点的最右键上移到父节点
	node.keys[index-1] = sibling.keys[len(sibling.keys)-1]
	node.values[index-1] = sibling.values[len(sibling.values)-1]

	// 从兄弟节点中移除该键
	sibling.keys = sibling.keys[:len(sibling.keys)-1]
	sibling.values = sibling.values[:len(sibling.values)-1]
}

// 从后一个兄弟节点借一个键
func (t *BTree[K, V]) borrowFromNext(node *bTreeNode[K, V], index int) {
	child := node.children[index]
	sibling := node.children[index+1]

	// 将父节点中的键下移到子节点
	child.keys = append(child.keys, node.keys[index])
	child.values = append(child.values, node.values[index])

	// 如果不是叶子节点，需要移动子节点指针
	if !child.leaf {
		// 移动兄弟节点的最左子节点
		child.children = append(child.children, sibling.children[0])
		// 从兄弟节点中移除该子节点
		sibling.children = sibling.children[1:]
	}

	// 将兄弟节点的最左键上移到父节点
	node.keys[index] = sibling.keys[0]
	node.values[index] = sibling.values[0]

	// 从兄弟节点中移除该键
	sibling.keys = sibling.keys[1:]
	sibling.values = sibling.values[1:]
}

// Size 返回树中键的数量
func (t *BTree[K, V]) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// IsEmpty 检查树是否为空
func (t *BTree[K, V]) IsEmpty() bool {
	return t.Size() == 0
}

// Clear 清空树
func (t *BTree[K, V]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.root = &bTreeNode[K, V]{
		keys:     make([]K, 0),
		values:   make([]V, 0),
		children: make([]*bTreeNode[K, V], 0),
		leaf:     true,
	}
	t.size = 0
}

// Keys 返回所有键的有序切片
func (t *BTree[K, V]) Keys() []K {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]K, 0, t.size)
	t.traverseInOrder(t.root, func(k K, v V) bool {
		result = append(result, k)
		return true
	})

	return result
}

// Values 返回与键对应的所有值的切片
func (t *BTree[K, V]) Values() []V {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]V, 0, t.size)
	t.traverseInOrder(t.root, func(k K, v V) bool {
		result = append(result, v)
		return true
	})

	return result
}

// KeyValues 返回所有键值对的有序切片
func (t *BTree[K, V]) KeyValues() ([]K, []V) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	keys := make([]K, 0, t.size)
	values := make([]V, 0, t.size)

	t.traverseInOrder(t.root, func(k K, v V) bool {
		keys = append(keys, k)
		values = append(values, v)
		return true
	})

	return keys, values
}

// ForEach 对树中的每个节点按顺序执行指定函数
func (t *BTree[K, V]) ForEach(fn func(key K, value V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.traverseInOrder(t.root, fn)
}

// 中序遍历
func (t *BTree[K, V]) traverseInOrder(node *bTreeNode[K, V], fn func(key K, value V) bool) bool {
	if node == nil {
		return true
	}

	for i := 0; i < len(node.keys); i++ {
		// 先遍历左子树
		if !node.leaf && !t.traverseInOrder(node.children[i], fn) {
			return false
		}

		// 处理当前键
		if !fn(node.keys[i], node.values[i]) {
			return false
		}
	}

	// 遍历最后一个子树
	if !node.leaf && !t.traverseInOrder(node.children[len(node.keys)], fn) {
		return false
	}

	return true
}

// --- 电商场景特定方法 ---

// GetOrDefault 获取键对应的值，如果键不存在则返回默认值
func (t *BTree[K, V]) GetOrDefault(key K, defaultValue V) V {
	t.mu.RLock()
	defer t.mu.RUnlock()

	value, err := t.search(t.root, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// PutIfAbsent 仅当键不存在时插入值
func (t *BTree[K, V]) PutIfAbsent(key K, value V) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, err := t.search(t.root, key)
	if err == nil {
		// 键已存在
		return false
	}

	// 键不存在，插入
	root := t.root
	maxKeys := 2*t.degree - 1

	// 如果根节点已满，需要分裂
	if len(root.keys) == maxKeys {
		newRoot := &bTreeNode[K, V]{
			keys:     make([]K, 0),
			values:   make([]V, 0),
			children: []*bTreeNode[K, V]{root},
			leaf:     false,
		}
		t.root = newRoot
		t.splitChild(newRoot, 0)
		t.insertNonFull(newRoot, key, value)
	} else {
		t.insertNonFull(root, key, value)
	}

	return true
}

// FindRange 查找键在指定范围内的所有值
func (t *BTree[K, V]) FindRange(fromKey, toKey K) ([]K, []V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 验证范围
	if t.comparator(fromKey, toKey) >= 0 {
		return nil, nil, ErrInvalidRange
	}

	keys := make([]K, 0)
	values := make([]V, 0)

	t.traverseInOrder(t.root, func(k K, v V) bool {
		if t.comparator(k, fromKey) >= 0 && t.comparator(k, toKey) < 0 {
			keys = append(keys, k)
			values = append(values, v)
		}
		// 如果已经超过了上限，可以提前终止遍历
		return t.comparator(k, toKey) < 0
	})

	return keys, values, nil
}

// BatchInsert 批量插入键值对
func (t *BTree[K, V]) BatchInsert(pairs map[K]V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for k, v := range pairs {
		root := t.root
		maxKeys := 2*t.degree - 1

		// 如果根节点已满，需要分裂
		if len(root.keys) == maxKeys {
			newRoot := &bTreeNode[K, V]{
				keys:     make([]K, 0),
				values:   make([]V, 0),
				children: []*bTreeNode[K, V]{root},
				leaf:     false,
			}
			t.root = newRoot
			t.splitChild(newRoot, 0)
			t.insertNonFull(newRoot, k, v)
		} else {
			t.insertNonFull(root, k, v)
		}
	}
}

// 为电商场景提供的批量获取方法
// 特别适合购物车、订单列表等场景
func (t *BTree[K, V]) BatchGet(keys []K) map[K]V {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[K]V)
	for _, key := range keys {
		value, err := t.search(t.root, key)
		if err == nil {
			result[key] = value
		}
	}

	return result
}

// 原子更新操作，适用于库存、计数器等场景
func (t *BTree[K, V]) ComputeIfPresent(key K, updateFn func(V) V) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	value, err := t.search(t.root, key)
	if err != nil {
		var zero V
		return zero, false
	}

	// 更新值
	newValue := updateFn(value)

	// 查找键的位置
	node := t.root
	for {
		i := 0
		for i < len(node.keys) && t.comparator(key, node.keys[i]) > 0 {
			i++
		}

		if i < len(node.keys) && t.comparator(key, node.keys[i]) == 0 {
			// 找到键，更新值
			node.values[i] = newValue
			return newValue, true
		}

		if node.leaf {
			// 键不存在
			break
		}

		// 递归到下一层
		node = node.children[i]
	}

	// 这不应该发生，因为我们已经检查了键是否存在
	var zero V
	return zero, false
}
