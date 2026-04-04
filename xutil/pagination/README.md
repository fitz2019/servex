# pagination

`github.com/Tsukikage7/servex/xutil/pagination` -- 分页工具。

## 概述

pagination 包提供通用分页参数处理与分页结果封装，自动进行参数规范化（边界校验、默认值填充），支持泛型分页结果。

## 功能特性

- 自动规范化分页参数（页码最小为 1，每页数量限制在 1-100 之间）
- 提供 Offset/Limit 方法，方便数据库查询
- 泛型分页结果 `Result[T]`，支持任意数据类型
- 计算总页数、判断是否有上/下一页

## API

### 类型

| 类型 | 说明 |
|------|------|
| `Pagination` | 分页参数，包含 Page（int32）和 PageSize（int32） |
| `Result[T]` | 分页结果，包含 Items、Total、Page、PageSize |

### 常量

| 常量 | 值 | 说明 |
|------|-----|------|
| `DefaultPage` | `1` | 默认页码 |
| `DefaultPageSize` | `20` | 默认每页数量 |
| `MaxPageSize` | `100` | 最大每页数量 |

### 函数

| 函数 | 说明 |
|------|------|
| `New(page, pageSize int32) Pagination` | 创建分页参数，自动规范化 |
| `NewResult[T](items, total, pagination) Result[T]` | 创建分页结果 |

### Pagination 方法

| 方法 | 说明 |
|------|------|
| `Offset() int` | 计算偏移量 `(Page - 1) * PageSize` |
| `Limit() int` | 返回每页数量 |

### Result[T] 方法

| 方法 | 说明 |
|------|------|
| `TotalPages() int32` | 计算总页数 |
| `HasNext() bool` | 是否有下一页 |
| `HasPrev() bool` | 是否有上一页 |
