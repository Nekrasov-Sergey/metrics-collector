package pool

import (
	"sync"

	"github.com/pkg/errors"
)

type resettable interface {
	Reset()
}

type Pool[T resettable] struct {
	pool  sync.Pool
	newFn func() T
}

func New[T resettable](newFn func() T) (*Pool[T], error) {
	if newFn == nil {
		return nil, errors.New("функция создания объекта не может быть nil")
	}

	p := &Pool[T]{
		newFn: newFn,
	}

	p.pool.New = func() any {
		return newFn()
	}

	return p, nil
}

func (p *Pool[T]) Get() T {
	if v := p.pool.Get(); v != nil {
		return v.(T)
	}
	return p.newFn()
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
