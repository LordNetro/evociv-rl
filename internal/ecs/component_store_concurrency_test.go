package ecs

import (
    "sync"
    "sync/atomic"
    "testing"
)

// TestComponentStoreConcurrency performs concurrent Set/Get/Delete operations
// to ensure the ComponentStore is safe for concurrent readers/writers.
func TestComponentStoreConcurrency(t *testing.T) {
    cs := NewComponentStore[int]()

    const entities = 1000
    const writers = 8
    const readers = 8

    var wg sync.WaitGroup
    var writes uint64
    var reads uint64

    // Writers
    for w := 0; w < writers; w++ {
        wg.Add(1)
        go func(wid int) {
            defer wg.Done()
            for i := 0; i < entities; i++ {
                e := Entity(uint64(wid*entities + i + 1))
                cs.Set(e, i)
                atomic.AddUint64(&writes, 1)
            }
        }(w)
    }

    // Readers
    for r := 0; r < readers; r++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < entities; i++ {
                e := Entity(uint64((i % entities) + 1))
                _, _ = cs.Get(e)
                _ = cs.Has(e)
                atomic.AddUint64(&reads, 1)
            }
        }()
    }

    // Deleter
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 1; i <= entities; i++ {
            cs.Delete(Entity(uint64(i)))
        }
    }()

    wg.Wait()

    // Basic sanity checks
    _ = cs.Len()
    _ = cs.All()

    if atomic.LoadUint64(&writes) == 0 {
        t.Fatalf("no writes performed")
    }
    if atomic.LoadUint64(&reads) == 0 {
        t.Fatalf("no reads performed")
    }
}
