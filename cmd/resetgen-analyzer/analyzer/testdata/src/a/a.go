package a

import "sync"

type Buffer struct {
	data []byte
}

func (b *Buffer) Reset() {
	b.data = b.data[:0]
}

var bufferPool = sync.Pool{
	New: func() any { return new(Buffer) },
}

func GoodUsage() {
	buf := bufferPool.Get().(*Buffer)
	buf.data = append(buf.data, "hello"...)
	buf.Reset()
	bufferPool.Put(buf) // OK: Reset() was called
}

func BadUsage() {
	buf := bufferPool.Get().(*Buffer)
	buf.data = append(buf.data, "hello"...)
	bufferPool.Put(buf) // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on buf"
}

func BadUsageAssign() {
	var buf *Buffer
	buf = bufferPool.Get().(*Buffer)
	buf.data = append(buf.data, "world"...)
	bufferPool.Put(buf) // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on buf"
}

func MultipleVars() {
	buf1 := bufferPool.Get().(*Buffer)
	buf2 := bufferPool.Get().(*Buffer)

	buf1.Reset()
	bufferPool.Put(buf1) // OK

	bufferPool.Put(buf2) // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on buf2"
}

// Test struct field pool
type Service struct {
	pool sync.Pool
}

func NewService() *Service {
	return &Service{
		pool: sync.Pool{New: func() any { return new(Buffer) }},
	}
}

func (s *Service) GoodFieldPool() {
	buf := s.pool.Get().(*Buffer)
	buf.data = append(buf.data, "test"...)
	buf.Reset()
	s.pool.Put(buf) // OK
}

func (s *Service) BadFieldPool() {
	buf := s.pool.Get().(*Buffer)
	buf.data = append(buf.data, "test"...)
	s.pool.Put(buf) // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on buf"
}

// Test wrapped in struct (e.g., buf stored in wrapper)
type Wrapper struct {
	buf *Buffer
}

func WrappedGood() {
	w := &Wrapper{}
	w.buf = bufferPool.Get().(*Buffer)
	w.buf.Reset()
	bufferPool.Put(w.buf) // OK - tracked by root var 'w'
}

func WrappedBad() {
	w := &Wrapper{}
	w.buf = bufferPool.Get().(*Buffer)
	bufferPool.Put(w.buf) // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on w"
}

// Generic pool wrapper that handles Reset internally - should NOT trigger
type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	p sync.Pool
}

func NewPool[T Resetter](newFn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{New: func() any { return newFn() }},
	}
}

func (p *Pool[T]) Get() T  { return p.p.Get().(T) }
func (p *Pool[T]) Put(v T) { v.Reset(); p.p.Put(v) } // Reset called internally

var genericPool = NewPool(func() *Buffer { return new(Buffer) })

func GenericPoolUsage() {
	buf := genericPool.Get()
	buf.data = append(buf.data, "test"...)
	genericPool.Put(buf) // OK: generic pool calls Reset() internally
}

// Bad wrapper - doesn't call Reset internally
type BadPool[T Resetter] struct {
	p sync.Pool
}

func NewBadPool[T Resetter](newFn func() T) *BadPool[T] {
	return &BadPool[T]{
		p: sync.Pool{New: func() any { return newFn() }},
	}
}

func (p *BadPool[T]) Get() T  { return p.p.Get().(T) }
func (p *BadPool[T]) Put(v T) { p.p.Put(v) } // want "sync.Pool.Put\\(\\) called without Reset\\(\\) on v"
