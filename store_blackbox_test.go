package expiring_gocache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/store"
	expiring "github.com/nabowler/expiring_gocache"
	"github.com/stretchr/testify/assert"
)

type (
	MapStore struct {
		cache           map[interface{}]interface{}
		setCount        int
		getCount        int
		deleteCount     int
		clearCount      int
		invalidateCount int
	}

	NonClearable struct{}
)

var (
	MapStoreMiss = errors.New("miss")
)

const (
	defaultExpiration = 1 * time.Second
	defaultSleep      = defaultExpiration + 10*time.Millisecond
)

func TestDefaultExpiration(t *testing.T) {
	key := "key"
	value := "value"

	ms := MapStore{cache: map[interface{}]interface{}{}}
	es := expiring.New(&ms, &store.Options{Expiration: defaultExpiration})

	// Nothing inserted yet. Should miss.
	val, err := es.Get(key)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	assert.Equal(t, 1, ms.getCount)

	// Insert a value with a no expiration.
	err = es.Set(key, value, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.setCount)

	// Verify value is retrievable
	val, err = es.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 2, ms.getCount)

	time.Sleep(defaultSleep)
	// the cached value should be expired
	_, err = es.Get(key)
	assert.Equal(t, expiring.ValueExpiredError, err)
	assert.Equal(t, 3, ms.getCount)

	// the value should now be gone from the cache
	_, err = es.Get(key)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 4, ms.getCount)
}

func TestPerKeyExpiration(t *testing.T) {
	longerKey := "longer"
	shorterKey := "shorter"
	value := "value"

	ms := MapStore{cache: map[interface{}]interface{}{}}
	es := expiring.New(&ms, &store.Options{Expiration: defaultExpiration})

	// Nothing inserted yet. Should miss.
	val, err := es.Get(longerKey)
	assert.Nil(t, val)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 1, ms.getCount)

	// Insert a value with a longer-than-default expiration.
	longerExpiration := 2 * defaultExpiration
	err = es.Set(longerKey, value, &store.Options{Expiration: longerExpiration})
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.setCount)
	// Insert a value with a shorter-than-default expiration.
	err = es.Set(shorterKey, value, &store.Options{Expiration: defaultExpiration - 10*time.Millisecond})
	assert.Nil(t, err)
	assert.Equal(t, 2, ms.setCount)

	// Verify values are retrievable
	val, err = es.Get(longerKey)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 2, ms.getCount)

	val, err = es.Get(shorterKey)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 3, ms.getCount)

	time.Sleep(defaultSleep)
	//the value should be expired because the expiration was shorter than the default
	_, err = es.Get(shorterKey)
	assert.Equal(t, expiring.ValueExpiredError, err)
	assert.Equal(t, 4, ms.getCount)
	// the cached value should still be valid because the expiration is longer than the default
	val, err = es.Get(longerKey)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 5, ms.getCount)

	time.Sleep(longerExpiration - defaultExpiration)
	// the cached value should be expired
	_, err = es.Get(longerKey)
	assert.Equal(t, expiring.ValueExpiredError, err)
	assert.Equal(t, 6, ms.getCount)

	// the value should now be gone from the cache
	_, err = es.Get(longerKey)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 7, ms.getCount)
}

func TestDelete(t *testing.T) {
	key := "key"
	value := "value"

	ms := MapStore{cache: map[interface{}]interface{}{}}
	es := expiring.New(&ms, &store.Options{Expiration: defaultExpiration})

	// Nothing inserted yet. Should miss.
	val, err := es.Get(key)
	assert.Nil(t, val)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 1, ms.getCount)

	err = es.Set(key, value, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.setCount)

	// Verify value is retrievable
	val, err = es.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 2, ms.getCount)

	// Delete the value.
	err = es.Delete(key)
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.deleteCount)

	// Value should now miss
	_, err = es.Get(key)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 3, ms.getCount)

	time.Sleep(defaultSleep)
	// the value should continue to miss
	_, err = es.Get(key)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 4, ms.getCount)
}

func TestClearClearer(t *testing.T) {
	key := "key"
	value := "value"

	ms := MapStore{cache: map[interface{}]interface{}{}}
	es := expiring.New(&ms, &store.Options{Expiration: defaultExpiration})

	// Nothing inserted yet. Should miss.
	val, err := es.Get(key)
	assert.Nil(t, val)
	assert.Equal(t, err, MapStoreMiss)
	assert.Equal(t, 1, ms.getCount)

	err = es.Set(key, value, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.setCount)

	// Verify value is retrievable
	val, err = es.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.Equal(t, 2, ms.getCount)

	err = es.Clear()
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.clearCount)
}

func TestClearNonClearable(t *testing.T) {
	es := expiring.New(NonClearable{}, &store.Options{Expiration: defaultExpiration})
	assert.Nil(t, es.Clear())
}

func TestInvalidate(t *testing.T) {
	ms := MapStore{cache: map[interface{}]interface{}{}}
	es := expiring.New(&ms, nil)

	err := es.Invalidate(store.InvalidateOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 1, ms.invalidateCount)
}

func TestGetType(t *testing.T) {
	es := expiring.New(nil, nil)
	assert.Equal(t, expiring.ExpiringStoreType, es.GetType())
}

// mapstore implementation

func (ms *MapStore) Get(key interface{}) (interface{}, error) {
	ms.getCount++
	val, ok := ms.cache[key]
	if !ok {
		return val, MapStoreMiss
	}
	return val, nil
}

func (ms *MapStore) Set(key interface{}, value interface{}, options *store.Options) error {
	ms.setCount++
	ms.cache[key] = value
	return nil
}

func (ms *MapStore) Delete(key interface{}) error {
	ms.deleteCount++
	delete(ms.cache, key)
	return nil
}

func (ms *MapStore) Invalidate(options store.InvalidateOptions) error {
	ms.invalidateCount++
	return nil
}

func (ms *MapStore) Clear() error {
	ms.clearCount++
	for k := range ms.cache {
		delete(ms.cache, k)
	}
	return nil
}

func (ms MapStore) GetType() string {
	return "MapStore"
}

// NonClearable impelmentation
func (nc NonClearable) Set(key interface{}, value interface{}, options *store.Options) error {
	return nil
}
func (nc NonClearable) Get(key interface{}) (interface{}, error)         { return nil, nil }
func (nc NonClearable) Delete(key interface{}) error                     { return nil }
func (nc NonClearable) Invalidate(options store.InvalidateOptions) error { return nil }
func (nc NonClearable) GetType() string                                  { return "non-clearable" }
