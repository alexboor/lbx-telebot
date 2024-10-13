package memory

// Memory defines the interface for a key-value storage.
type Memory interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string) bool
	Clear()
}
