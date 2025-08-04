package pkg

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Closer func() error

type Lifecycle struct {
	closers []Closer
}

func (l *Lifecycle) Add(fn Closer) {
	l.closers = append(l.closers, fn)
}

func (l *Lifecycle) Shutdown(ctx context.Context) error {
	for i := len(l.closers) - 1; i >= 0; i-- {
		if err := l.closers[i](); err != nil {
			return err
		}
	}

	return nil
}

type GoLifecycle struct {
	depMap    map[any][]any
	appCloser map[any]Closer
}

func (l *GoLifecycle) Add(app any, fn Closer, deps ...any) {
	for _, dep := range deps {
		l.depMap[dep] = append(l.depMap[dep], app)
	}

	l.appCloser[app] = fn
}

func (l *GoLifecycle) Shutdown() error {
	var indegree sync.Map
	for app := range l.appCloser {
		indegree.Store(app, new(int32))
	}
	for dep, dependents := range l.depMap {
		for _, dependent := range dependents {
			val, _ := indegree.Load(dependent)
			atomic.AddInt32(val.(*int32), 1)
		}
		if _, ok := indegree.Load(dep); !ok {
			indegree.Store(dep, new(int32))
		}
	}

	var mu sync.Mutex
	shutdownErrors := []error{}
	var wg sync.WaitGroup
	readyCh := make(chan any, 128) // buffered to prevent deadlock

	// Populate initial zero-indegree apps
	indegree.Range(func(key, value any) bool {
		if atomic.LoadInt32(value.(*int32)) == 0 {
			wg.Add(1)
			readyCh <- key
		}
		return true
	})

	// Worker goroutine: waits for shutdowns to complete
	go func() {
		wg.Wait()
		close(readyCh)
	}()

	// Process nodes
	for app := range readyCh {
		go func(app any) {
			defer wg.Done()

			if closer, ok := l.appCloser[app]; ok {
				if err := closer(); err != nil {
					mu.Lock()
					shutdownErrors = append(shutdownErrors, err)
					mu.Unlock()
				}
			}

			for _, dependent := range l.depMap[app] {
				if val, ok := indegree.Load(dependent); ok {
					if atomic.AddInt32(val.(*int32), -1) == 0 {
						wg.Add(1)
						readyCh <- dependent
					}
				}
			}
		}(app)
	}

	// Wait for shutdown and check for errors
	wg.Wait()

	// Check for cycles (any node with indegree > 0)
	hasCycle := false
	indegree.Range(func(_, value any) bool {
		if atomic.LoadInt32(value.(*int32)) > 0 {
			hasCycle = true
			return false
		}
		return true
	})

	if hasCycle {
		return fmt.Errorf("lifecycle shutdown cycle detected")
	}
	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}
	return nil
}
