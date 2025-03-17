package workerpool

import (
	"fmt"
	"time"

	"github.com/panjf2000/ants/v2"
)

type FuncArgs struct {
	Fn   func(...interface{})
	Args []interface{}
}

type WorkerPool struct {
	Pool *ants.PoolWithFunc
}

var pool *WorkerPool

func Initialize(size int, expiryDuration time.Duration) {
	funcArgs := func(i interface{}) {
		obj := i.(FuncArgs)
		fun := obj.Fn
		fun(obj.Args...)
	}
	workerPool, err := ants.NewPoolWithFunc(
		size,
		funcArgs,
		ants.WithExpiryDuration(expiryDuration),
	)
	if err != nil {
		fmt.Printf("could not initialize worker pool: %v\n", err)
		panic(err)
	}
	pool = &WorkerPool{Pool: workerPool}
}

func Pool() *WorkerPool {
	return pool
}

func (p *WorkerPool) Release() {
	p.Pool.Release()
}

func (p *WorkerPool) SubmitTask(fun func(...interface{}), args ...interface{}) error {
	return p.Pool.Invoke(FuncArgs{
		Fn:   fun,
		Args: args,
	})
}
