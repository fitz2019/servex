package httpclient

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

// Balancer 负载均衡器接口.
type Balancer interface {
	// Pick 从候选地址列表中选择一个目标地址.
	Pick(addrs []string) string
}

// RoundRobinBalancer 轮询负载均衡器.
type RoundRobinBalancer struct {
	index atomic.Uint64
}

// Pick 按轮询顺序选择地址.
func (b *RoundRobinBalancer) Pick(addrs []string) string {
	if len(addrs) == 0 {
		return ""
	}
	idx := b.index.Add(1) - 1
	return addrs[int(idx)%len(addrs)]
}

// RandomBalancer 随机负载均衡器.
type RandomBalancer struct {
	mu sync.Mutex
}

// Pick 随机选择地址.
func (b *RandomBalancer) Pick(addrs []string) string {
	if len(addrs) == 0 {
		return ""
	}
	b.mu.Lock()
	idx := rand.Intn(len(addrs))
	b.mu.Unlock()
	return addrs[idx]
}
