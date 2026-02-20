# OpenContextManager

Go 语言大模型上下文管理器 — 架构清晰、存储解耦、业务可定制。

## 架构设计

```
┌─────────────────────────────────────────────┐
│              Manager (业务层)                │
│  ┌─────────────┐    ┌────────────────────┐  │
│  │  Strategy    │    │  Storage (接口)     │  │
│  │  (上下文策略) │    │                    │  │
│  └─────────────┘    └────────┬───────────┘  │
│                              │              │
└──────────────────────────────┼──────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼──────┐ ┌──────▼───────┐ ┌──────▼───────┐
     │ memory.Store  │ │  file.Store  │ │ 自定义 Store  │
     │  (内存存储)    │ │  (文件存储)   │ │ (Redis 等)   │
     └───────────────┘ └──────────────┘ └──────────────┘
```

### 核心设计原则

1. **存储解耦** — `Storage` 接口将业务逻辑与数据存储完全分离，可随时切换存储后端
2. **策略模式** — `Strategy` 接口支持自定义上下文处理策略（滑动窗口、Token 限制等）
3. **可扩展模型** — `Message` 和 `Context` 支持自定义 `Metadata`，满足不同业务场景

## 安装

```bash
go get github.com/VeryGoodStudy/OpenContextManager
```

## 快速开始

```go
package main

import (
    ocm "github.com/VeryGoodStudy/OpenContextManager"
    "github.com/VeryGoodStudy/OpenContextManager/store/memory"
)

func main() {
    // 创建存储后端（可替换为 file.Store 或自定义实现）
    store := memory.New()

    // 创建管理器，可选配置上下文策略
    manager := ocm.NewManager(store, ocm.WithStrategy(
        ocm.NewSlidingWindowStrategy(20),
    ))

    // 创建上下文
    manager.Create("chat-1")

    // 添加消息
    manager.Append("chat-1", ocm.NewMessage(ocm.RoleSystem, "You are a helpful assistant."))
    manager.Append("chat-1", ocm.NewMessage(ocm.RoleUser, "Hello!"))

    // 获取上下文（自动应用策略）
    ctx, _ := manager.Get("chat-1")
    _ = ctx
}
```

## 切换存储后端

```go
// 内存存储
store := memory.New()

// 文件存储
store, _ := file.New("/path/to/data")

// 自定义存储 — 只需实现 ocm.Storage 接口
type MyRedisStore struct { /* ... */ }
func (s *MyRedisStore) Save(ctx *ocm.Context) error   { /* ... */ }
func (s *MyRedisStore) Load(id string) (*ocm.Context, error) { /* ... */ }
func (s *MyRedisStore) Delete(id string) error         { /* ... */ }
func (s *MyRedisStore) List() ([]string, error)        { /* ... */ }
```

## 项目结构

```
├── model.go              # 核心数据模型 (Message, Context, Role)
├── storage.go            # Storage 存储接口定义
├── strategy.go           # Strategy 策略接口及内置实现
├── manager.go            # Manager 业务管理器
├── store/
│   ├── memory/           # 内存存储实现
│   │   └── store.go
│   └── file/             # 文件存储实现
│       └── store.go
└── cmd/
    └── example/          # 使用示例
        └── main.go
```

## 测试

```bash
go test ./...
```
