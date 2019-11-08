# expiring_gocache

This provides a simple way to enforce expirations around caches that would otherwise not honor the `Expiration` store option.

## Usage

```go

inMemoryCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
if err != nil {
    // handle the error
}

inMemoryStore := store.NewRistretto(inMemoryCache, &store.Options{
    Cost:       1,
})

expiringStore := expiring.New(inMemoryStore, &store.Options{Expiration: 1 * time.Minute})

```