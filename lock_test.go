package kmutex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestBasic(t *testing.T) {
	km := NewKMutex()

	key1 := fmt.Sprintf("%s-%v", t.Name(), 1)
	key2 := fmt.Sprintf("%s-%v", t.Name(), 2)

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Millisecond)
	if err := km.Lock(ctx, key1); err != nil {
		t.Errorf("expected no nil, but got %v", err)
	}

	if err := km.Lock(ctx, key2); err != nil {
		t.Errorf("expected no nil, but got %v", err)
	}

	cancel()
	if err := km.Lock(ctx, key1); err != context.Canceled {
		t.Errorf("expected cancel error, but got %v", err)
	}

	km.Unlock(key1)
	km.Unlock(key2)
}

func TestMultiKeyWaitShouldLockAfterWake(t *testing.T) {
	n := 10

	keys := make(map[string]struct{})
	for i := 0; i < n; i++ {
		keys[fmt.Sprintf("%s-%d", t.Name(), i)] = struct{}{}
	}

	running := make(chan string, n)
	init := false
	hookFn := func(key string) {
		if !init {
			running <- key
		}
	}

	km := NewKMutex()
	km.waitHook = hookFn

	initCtx, initCancel := context.WithTimeout(context.TODO(), 1*time.Minute)
	defer initCancel()

	waitCtx, waitCancel := context.WithCancel(context.TODO())
	g, waitCtx := errgroup.WithContext(waitCtx)
	for k := range keys {
		if err := km.Lock(initCtx, k); err != nil {
			t.Errorf("expected not error, but got %v", err)
		}

		go func(key string) {
			g.Go(func() error {
				if err := km.Lock(waitCtx, key); err != nil {
					return fmt.Errorf("expected not error, but got %v", err)
				}
				return nil
			})
		}(k)
	}

	// make sure that all the goroutines wait
	for i := 0; i < n; i++ {
		<-running
	}
	init = true

	// remove all the keys and broadcast
	km.mu.Lock()
	for k := range km.keys {
		delete(km.keys, k)
	}
	km.cond.Broadcast()
	km.mu.Unlock()

	waitCh := make(chan error, 1)
	go func() {
		defer close(waitCh)
		waitCh <- g.Wait()
	}()

	select {
	case <-time.After(3 * time.Second):
		waitCancel()

		if err := <-waitCh; err == nil {
			t.Errorf("expected error, but got nil")
		}
	case err := <-waitCh:
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}
	}

	if len(km.keys) != len(keys) {
		t.Errorf("expected all the keys(%v) are locked, but got %v", keys, km.keys)
	}

	for k := range keys {
		if _, ok := km.keys[k]; !ok {
			t.Errorf("expected key(%v) is locked, but got nothing", k)
		}
	}
}
