package kmutex

import (
	"context"
	"fmt"
	"sync"
)

// KMutex allows to lock by key.
type KMutex struct {
	cond *sync.Cond
	mu   *sync.Mutex
	keys map[string]struct{}

	// NOTE: waitHook is only used for testing
	waitHook func(key string)
}

// NewKMutex returns multiple key locker.
func NewKMutex() *KMutex {
	km := &KMutex{
		mu:   new(sync.Mutex),
		keys: make(map[string]struct{}),
	}
	km.cond = sync.NewCond(km.mu)
	return km
}

// Lock locks by key and allows to use context to cancel.
func (km *KMutex) Lock(ctx context.Context, key string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, ok := km.keys[key]
			if !ok {
				km.keys[key] = struct{}{}
				return nil
			}
		}

		if km.waitHook != nil {
			km.waitHook(key)
		}
		km.cond.Wait()
	}
	panic("unreachable")
}

// Unlock unlocks the key.
func (km *KMutex) Unlock(key string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if _, ok := km.keys[key]; !ok {
		panic(fmt.Sprintf("unlock non-exist key: %s", key))
	}
	delete(km.keys, key)
	km.cond.Broadcast()
}
