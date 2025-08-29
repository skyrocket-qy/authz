package util_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"authz/internal/util"
	"github.com/stretchr/testify/assert"
)

// --- Test SimpleLifecycle ---

func TestSimpleLifecycle(t *testing.T) {
	t.Run("should run closers in LIFO order", func(t *testing.T) {
		lc := util.NewSimpleLifecycle()

		var order []string

		lc.Add(func(ctx context.Context) error {
			order = append(order, "first")

			return nil
		})
		lc.Add(func(ctx context.Context) error {
			order = append(order, "second")

			return nil
		})
		lc.Add(func(ctx context.Context) error {
			order = append(order, "third")

			return nil
		})

		err := lc.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []string{"third", "second", "first"}, order)
	})

	t.Run("should stop and return error on first failure", func(t *testing.T) {
		lc := util.NewSimpleLifecycle()

		var order []string

		testErr := errors.New("shutdown failed")

		lc.Add(func(ctx context.Context) error {
			order = append(order, "first")

			return nil
		})
		lc.Add(func(ctx context.Context) error {
			order = append(order, "second")

			return testErr
		})

		err := lc.Shutdown(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		// Only the second closer should have run
		assert.Equal(t, []string{"second"}, order)
	})
}

// --- Test LifecycleParallel ---

func TestLifecycleParallel(t *testing.T) {
	t.Run("should run closers according to dependency order", func(t *testing.T) {
		lc := util.NewLifecycleParallel()
		orderCh := make(chan string, 3)

		// A -> B -> C
		// Shutdown order should be A, then B, then C
		appA := "A"
		appB := "B"
		appC := "C"

		lc.Add(appA, func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)

			orderCh <- appA

			return nil
		}, appB) // A depends on B

		lc.Add(appB, func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)

			orderCh <- appB

			return nil
		}, appC) // B depends on C

		lc.Add(appC, func(ctx context.Context) error {
			orderCh <- appC

			return nil
		}) // C has no dependencies

		err := lc.Shutdown(context.Background())

		close(orderCh)

		assert.NoError(t, err)

		var order []string
		for app := range orderCh {
			order = append(order, app)
		}

		// The shutdown order should be A, then B, then C.
		// For this test, given the sleeps, the order should be deterministic.
		assert.Len(t, order, 3)
		assert.Equal(t, appA, order[0])
		assert.Equal(t, appB, order[1])
		assert.Equal(t, appC, order[2])
	})

	t.Run("should run independent closers in parallel", func(t *testing.T) {
		lc := util.NewLifecycleParallel()

		closerDuration := 50 * time.Millisecond

		var wg sync.WaitGroup
		wg.Add(2)

		lc.Add("A", func(ctx context.Context) error {
			defer wg.Done()

			time.Sleep(closerDuration)

			return nil
		})
		lc.Add("B", func(ctx context.Context) error {
			defer wg.Done()

			time.Sleep(closerDuration)

			return nil
		})

		start := time.Now()
		err := lc.Shutdown(context.Background())

		wg.Wait() // Wait for closers to finish

		duration := time.Since(start)

		assert.NoError(t, err)
		// Duration should be slightly more than one closer, not two.
		assert.Less(t, duration, closerDuration*2)
		assert.Greater(t, duration, closerDuration)
	})

	t.Run("should stop on first error in parallel execution", func(t *testing.T) {
		lc := util.NewLifecycleParallel()
		testErr := errors.New("shutdown failed")

		var wg sync.WaitGroup
		wg.Add(1) // Only one closer should execute

		// A -> B. A will fail. B should not run.
		lc.Add("A", func(ctx context.Context) error {
			defer wg.Done()

			return testErr
		}, "B")

		var bExecuted bool

		lc.Add("B", func(ctx context.Context) error {
			bExecuted = true

			return nil
		})

		err := lc.Shutdown(context.Background())

		wg.Wait()

		assert.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.False(t, bExecuted, "dependent closer should not have been executed")
	})
}
