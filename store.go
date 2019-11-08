package expiring_gocache

import (
	"errors"
	"time"

	"github.com/eko/gocache/store"
)

type (
	Store struct {
		expiration time.Duration
		store      store.StoreInterface
	}

	wrappedValue struct {
		expireAt time.Time
		value    interface{}
	}

	clearer interface {
		Clear() error
	}
)

const (
	ExpiringStoreType = "expiring"

	DefaultExpiration = 720 * time.Hour
)

var (
	ValueExpiredError = errors.New("cached value has expired")
)

func New(store store.StoreInterface, options *store.Options) Store {
	expiration := DefaultExpiration
	if options != nil {
		expiration = options.ExpirationValue()
	}

	return Store{
		expiration: expiration,
		store:      store,
	}
}

// Get retrieves the value from the underlying store. If the value is
// expired, `(_, ValueExpiredError)` is returned; no guarantee is made
// about the first returned value.
func (es Store) Get(key interface{}) (interface{}, error) {
	val, err := es.store.Get(key)
	if err != nil || val == nil {
		return val, err
	}

	ew, ok := val.(wrappedValue)
	if !ok {
		// value was not a wrapped value. return it directly.
		return val, nil
	}

	if ew.expireAt.Before(time.Now()) {
		// value is expired. try to delete it from the store and return ValueExpiredError
		_ = es.store.Delete(key) //best effort delete
		return ew.value, ValueExpiredError
	}

	return ew.value, nil
}

func (es Store) Set(key interface{}, value interface{}, options *store.Options) error {
	expireAt := time.Now().Add(es.expiration)
	if options != nil && options.ExpirationValue() > 0 {
		expireAt = time.Now().Add(options.ExpirationValue())
	}
	return es.store.Set(key, wrappedValue{expireAt: expireAt, value: value}, options)
}

func (es Store) Delete(key interface{}) error {
	return es.store.Delete(key)
}

func (es Store) Invalidate(options store.InvalidateOptions) error {
	return es.store.Invalidate(options)
}

func (es Store) Clear() error {
	// Clear() is in the StoreInterface on Master, but isn't in the latest (v0.2.0) release.
	// Target v0.2.0, support current HEAD on Master
	clear, ok := es.store.(clearer)
	if ok {
		return clear.Clear()
	}
	return nil
}

func (es Store) GetType() string {
	return ExpiringStoreType
}
