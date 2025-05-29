<!--
Copyright 2024 Humphrey

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

```mermaid
graph TD
    A[E:\GO_Prop\go-generic-utils\ggu] --> B[E:\GO_Prop\go-generic-utils\ggu\README.md]
A --> B2[LICENSE]
    A --> B3[go.mod]
    A --> B4[go.sum]
    A --> B5[.gitignore]
    A --> B6[.github/ (CI/CD, issue templates等)]

    A --> C1[**sliceutils** (切片辅助方法)]
        C1 --> C1F1[slice.go (核心切片函数)]
        C1 --> C1F2[slice_test.go (单元测试)]
        C1 --> C1F3[example_test.go (示例代码)]
        C1 --> C1F4[doc.go (包文档)]

    A --> C2[**maputils** (Map辅助方法)]
        C2 --> C2F1[map.go]
        C2 --> C2F2[map_test.go]
        C2 --> C2F3[example_test.go]
        C2 --> C2F4[doc.go]

    A --> C3[**collections** (扩展集合类型)]
        C3 --> D1[**hashmap** (如果需要封装原生map或实现特定并发HashMap)]
            D1 --> D1F1[hashmap.go]
            D1 --> D1F2[hashmap_test.go]
        C3 --> D2[**treemap** (有序Map, 如红黑树实现)]
            D2 --> D2F1[treemap.go]
            D2 --> D2F2[node.go (树节点定义)]
            D2 --> D2F3[comparator.go (比较器接口/默认实现)]
            D2 --> D2F4[treemap_test.go]
        C3 --> D3[**linkedmap** (保持插入顺序的Map)]
            D3 --> D3F1[linkedmap.go]
            D3 --> D3F2[linkedmap_test.go]
        C3 --> D4[**lists** (列表实现)]
            D4 --> E1[**arraylist** (基于切片的列表)]
                E1 --> E1F1[arraylist.go]
                E1 --> E1F2[arraylist_test.go]
            D4 --> E2[**linkedlist** (双向链表)]
                E2 --> E2F1[linkedlist.go]
                E2 --> E2F2[node.go (链表节点定义)]
                E2 --> E2F3[linkedlist_test.go]
            D4 --> E3[**skiplist** (跳表)]
                E3 --> E3F1[skiplist.go]
                E3 --> E3F2[node.go (跳表节点定义)]
                E3 --> E3F3[skiplist_test.go]
        C3 --> D5[**sets** (集合实现)]
            D5 --> F1[**hashset** (基于map的无序Set)]
                F1 --> F1F1[hashset.go]
                F1 --> F1F2[hashset_test.go]
            D5 --> F2[**treeset** (基于TreeMap的有序Set)]
                F2 --> F2F1[treeset.go]
                F2 --> F2F2[treeset_test.go]
            D5 --> F3[**sortedset** (类似Redis的ZSet, 包含score)]
                F3 --> F3F1[sortedset.go (可能基于SkipList或TreeMap)]
                F3 --> F3F2[sortedset_test.go]
        C3 --> D6[**queues** (队列实现)]
            D6 --> G1[**queue** (普通FIFO队列)]
                G1 --> G1F1[queue.go (可基于LinkedList或ArrayList)]
                G1 --> G1F2[queue_test.go]
            D6 --> G2[**priorityqueue** (优先级队列)]
                G2 --> G2F1[priorityqueue.go (基于container/heap封装)]
                G2 --> G2F2[item.go (队列项定义)]
                G2 --> G2F3[priorityqueue_test.go]
        C3 --> D7[doc.go (collections包文档)]


    A --> C4[**beans** (Bean 操作辅助类)]
        C4 --> C4F1[copier.go (高性能copier实现)]
        C4 --> C4F2[copier_test.go]
        C4 --> C4F3[options.go (copier的选项函数)]
        C4 --> C4F4[doc.go]

    A --> C5[**concurrent** (并发扩展工具)]
        C5 --> H1[**queue** (并发队列)]
            H1 --> H1F1[blocking_queue.go]
            H1 --> H1F2[blocking_priority_queue.go]
            H1 --> H1F3[queue_test.go]
        C5 --> H2[**pool** (协程池)]
            H2 --> H2F1[goroutine_pool.go]
            H2 --> H2F2[task.go (任务定义)]
            H2 --> H2F3[goroutine_pool_test.go]
        C5 --> H3[**cmap** (如果需要提供一个不同于sync.Map的并发Map实现，如分片Map)]
             H3 --> H3F1[concurrent_map.go]
             H3 --> H3F2[concurrent_map_test.go]
        C5 --> H4[doc.go (concurrent包文档)]

    A --> C6[**utils** (其他通用工具，不便归类的)]
        C6 --> C6F1[utils.go]
        C6 --> C6F2[utils_test.go]
        C6 --> C6F3[doc.go]

    A --> C7[**CONTRIBUTING.md** (贡献指南)]
    A --> C8[**examples/** (使用示例目录)]
        C8 --> C8F1[sliceutils_example/main.go]
        C8 --> C8F2[collections_example/main.go]
        C8 --> C8F3[...]

    A --> C9[**internal/** (内部包，不希望外部直接导入的辅助代码)]
        C9 --> C9F1[constraints/constraints.go (如果需要自定义泛型约束)]
        C9 --> C9F2[...]

    A --> C10[**scripts/** (构建、测试、代码生成等辅助脚本)]

    %% Styling
    classDef root fill:#227700,stroke:#333,stroke-width:4px,color:#fff
    classDef module fill:#ccf,stroke:#333,stroke-width:2px
    classDef submodule fill:#e0e0e0,stroke:#333,stroke-width:1px
    classDef file fill:#fff,stroke:#666,stroke-width:1px,color:#333

    class A root;
    class C1,C2,C3,C4,C5,C6,C7,C8,C9,C10 module;
    class D1,D2,D3,D4,D5,D6,E1,E2,E3,F1,F2,F3,G1,G2,H1,H2,H3 submodule;
    class B1,B2,B3,B4,B5,B6,C1F1,C1F2,C1F3,C1F4,C2F1,C2F2,C2F3,C2F4,D1F1,D1F2,D2F1,D2F2,D2F3,D2F4,D3F1,D3F2,E1F1,E1F2,E2F1,E2F2,E2F3,E3F1,E3F2,E3F3,F1F1,F1F2,F2F1,F2F2,F3F1,F3F2,G1F1,G1F2,G2F1,G2F2,G2F3,C4F1,C4F2,C4F3,C4F4,H1F1,H1F2,H1F3,H2F1,H2F2,H2F3,H3F1,H3F2,H4,C6F1,C6F2,C6F3,D7,C8F1,C8F2,C8F3,C9F1,C9F2,H4 file;

```