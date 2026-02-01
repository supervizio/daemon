//! TTL-based cache entry implementation.

use std::time::{Duration, Instant};

/// A cached value with timestamp for TTL-based expiration.
#[derive(Debug, Clone)]
pub struct CacheEntry<T> {
    /// The cached value.
    pub value: T,
    /// When the value was cached.
    pub cached_at: Instant,
}

impl<T> CacheEntry<T> {
    /// Create a new cache entry with the current timestamp.
    pub fn new(value: T) -> Self {
        Self { value, cached_at: Instant::now() }
    }

    /// Check if the cache entry is still valid based on TTL.
    pub fn is_valid(&self, ttl: Duration) -> bool {
        if ttl.is_zero() {
            return false;
        }
        self.cached_at.elapsed() < ttl
    }

    /// Check if the cache entry has expired.
    pub fn is_expired(&self, ttl: Duration) -> bool {
        !self.is_valid(ttl)
    }

    /// Get the age of the cache entry.
    pub fn age(&self) -> Duration {
        self.cached_at.elapsed()
    }

    /// Get a reference to the cached value.
    pub fn get(&self) -> &T {
        &self.value
    }

    /// Get the cached value, consuming the entry.
    pub fn into_value(self) -> T {
        self.value
    }
}

/// A simple TTL cache for arbitrary keys.
#[derive(Debug)]
pub struct TtlCache<K, V> {
    entries: std::collections::HashMap<K, CacheEntry<V>>,
    default_ttl: Duration,
}

impl<K: std::hash::Hash + Eq, V> TtlCache<K, V> {
    /// Create a new TTL cache with the given default TTL.
    pub fn new(default_ttl: Duration) -> Self {
        Self { entries: std::collections::HashMap::new(), default_ttl }
    }

    /// Insert a value into the cache.
    pub fn insert(&mut self, key: K, value: V) {
        self.entries.insert(key, CacheEntry::new(value));
    }

    /// Get a value from the cache if it exists and is not expired.
    pub fn get(&self, key: &K) -> Option<&V> {
        self.entries
            .get(key)
            .filter(|entry| entry.is_valid(self.default_ttl))
            .map(|entry| &entry.value)
    }

    /// Get a value from the cache with a custom TTL.
    pub fn get_with_ttl(&self, key: &K, ttl: Duration) -> Option<&V> {
        self.entries.get(key).filter(|entry| entry.is_valid(ttl)).map(|entry| &entry.value)
    }

    /// Remove a value from the cache.
    pub fn remove(&mut self, key: &K) -> Option<V> {
        self.entries.remove(key).map(|entry| entry.value)
    }

    /// Clear all entries from the cache.
    pub fn clear(&mut self) {
        self.entries.clear();
    }

    /// Remove all expired entries from the cache.
    pub fn cleanup(&mut self) {
        self.entries.retain(|_, entry| entry.is_valid(self.default_ttl));
    }

    /// Get the number of entries in the cache (including expired ones).
    pub fn len(&self) -> usize {
        self.entries.len()
    }

    /// Check if the cache is empty.
    pub fn is_empty(&self) -> bool {
        self.entries.is_empty()
    }

    /// Get the default TTL.
    pub fn default_ttl(&self) -> Duration {
        self.default_ttl
    }

    /// Set the default TTL.
    pub fn set_default_ttl(&mut self, ttl: Duration) {
        self.default_ttl = ttl;
    }
}

impl<K: std::hash::Hash + Eq, V: Clone> TtlCache<K, V> {
    /// Get a cloned value from the cache if it exists and is not expired.
    pub fn get_cloned(&self, key: &K) -> Option<V> {
        self.get(key).cloned()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;

    #[test]
    fn test_cache_entry_valid() {
        let entry = CacheEntry::new(42);
        assert!(entry.is_valid(Duration::from_secs(10)));
        assert!(!entry.is_expired(Duration::from_secs(10)));
    }

    #[test]
    fn test_cache_entry_expired() {
        let entry = CacheEntry::new(42);
        thread::sleep(Duration::from_millis(10));
        assert!(!entry.is_valid(Duration::from_millis(5)));
        assert!(entry.is_expired(Duration::from_millis(5)));
    }

    #[test]
    fn test_cache_entry_zero_ttl() {
        let entry = CacheEntry::new(42);
        assert!(!entry.is_valid(Duration::ZERO));
    }

    #[test]
    fn test_ttl_cache_basic() {
        let mut cache: TtlCache<&str, i32> = TtlCache::new(Duration::from_secs(10));

        cache.insert("key1", 100);
        assert_eq!(cache.get(&"key1"), Some(&100));
        assert_eq!(cache.get(&"key2"), None);
    }

    #[test]
    fn test_ttl_cache_expiry() {
        let mut cache: TtlCache<&str, i32> = TtlCache::new(Duration::from_millis(10));

        cache.insert("key1", 100);
        thread::sleep(Duration::from_millis(20));
        assert_eq!(cache.get(&"key1"), None);
    }

    #[test]
    fn test_ttl_cache_custom_ttl() {
        let mut cache: TtlCache<&str, i32> = TtlCache::new(Duration::from_millis(10));

        cache.insert("key1", 100);
        thread::sleep(Duration::from_millis(5));

        // With default TTL (10ms), should still be valid
        assert_eq!(cache.get(&"key1"), Some(&100));

        // With custom short TTL (1ms), should be expired
        assert_eq!(cache.get_with_ttl(&"key1", Duration::from_millis(1)), None);
    }

    #[test]
    fn test_ttl_cache_cleanup() {
        let mut cache: TtlCache<&str, i32> = TtlCache::new(Duration::from_millis(10));

        cache.insert("key1", 100);
        cache.insert("key2", 200);
        assert_eq!(cache.len(), 2);

        thread::sleep(Duration::from_millis(20));
        cache.cleanup();
        assert_eq!(cache.len(), 0);
    }

    #[test]
    fn test_cache_entry_age() {
        let entry = CacheEntry::new(42);
        thread::sleep(Duration::from_millis(5));
        assert!(entry.age() >= Duration::from_millis(5));
    }
}
