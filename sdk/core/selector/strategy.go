package selector

import (
	"context"
	"hash/crc32"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	sx "github.com/jxo-me/netx/sdk/core/internal/util/selector"
	"github.com/jxo-me/netx/sdk/core/logger"
	"github.com/jxo-me/netx/sdk/core/metadata"
	mdutil "github.com/jxo-me/netx/sdk/core/metadata/util"
)

type roundRobinStrategy[T any] struct {
	counter uint64
}

// RoundRobinStrategy is a strategy for node selector.
// The node will be selected by round-robin algorithm.
func RoundRobinStrategy[T any]() Strategy[T] {
	return &roundRobinStrategy[T]{}
}

func (s *roundRobinStrategy[T]) Apply(ctx context.Context, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}

	n := atomic.AddUint64(&s.counter, 1) - 1
	return vs[int(n%uint64(len(vs)))]
}

type randomStrategy[T any] struct {
	rw *randomWeighted[T]
	mu sync.Mutex
}

// RandomStrategy is a strategy for node selector.
// The node will be selected randomly.
func RandomStrategy[T any]() Strategy[T] {
	return &randomStrategy[T]{
		rw: newRandomWeighted[T](),
	}
}

func (s *randomStrategy[T]) Apply(ctx context.Context, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rw.Reset()
	for i := range vs {
		weight := 0
		if md, _ := any(vs[i]).(metadata.IMetaDatable); md != nil {
			weight = mdutil.GetInt(md.Metadata(), labelWeight)
		}
		if weight <= 0 {
			weight = 1
		}
		s.rw.Add(vs[i], weight)
	}

	return s.rw.Next()
}

type fifoStrategy[T any] struct{}

// FIFOStrategy is a strategy for node selector.
// The node will be selected from first to last,
// and will stick to the selected node until it is failed.
func FIFOStrategy[T any]() Strategy[T] {
	return &fifoStrategy[T]{}
}

// Apply applies the fifo strategy for the nodes.
func (s *fifoStrategy[T]) Apply(ctx context.Context, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}
	return vs[0]
}

type hashStrategy[T any] struct {
	r  *rand.Rand
	mu sync.Mutex
}

func HashStrategy[T any]() Strategy[T] {
	return &hashStrategy[T]{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *hashStrategy[T]) Apply(ctx context.Context, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}
	if h := sx.HashFromContext(ctx); h != nil {
		value := uint64(crc32.ChecksumIEEE([]byte(h.Source)))
		logger.Default().Tracef("hash %s %d", h.Source, value)
		return vs[value%uint64(len(vs))]
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return vs[s.r.Intn(len(vs))]
}
