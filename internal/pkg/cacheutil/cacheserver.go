package cacheutil

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/incoming"
)

// MAXTTL to remove plugin is stale for 5
var MAXTTL int64 = 300
var freeList = make(chan *IncomingBuffer, 1000)
var quitCacheServerCh = make(chan struct{})
var debugc = func(format string, data ...interface{}) {} // Default no debugging output

//ApplicationHealthCache  ...
type ApplicationHealthCache struct {
	QpidRouterState    int
	ElasticSearchState int
	LastAccess         int64 //timestamp in seconds
}

//NewApplicationHealthCache  ..
func NewApplicationHealthCache() *ApplicationHealthCache {
	return &ApplicationHealthCache{
		QpidRouterState:    0,
		LastAccess:         0,
		ElasticSearchState: 0,
	}

}

//IncomingBuffer  this is inut data send to cache server
//IncomingBuffer  ..its of type collectd or anything else
type IncomingBuffer struct {
	data incoming.DataTypeInterface
}

//IncomingDataCache cache server converts it into this
type IncomingDataCache struct {
	hosts  map[string]*ShardedIncomingDataCache
	maxTTL int64
	lock   *sync.RWMutex
}

//ShardedIncomingDataCache types of sharded cache collectd, influxdb etc
//ShardedIncomingDataCache  ..
type ShardedIncomingDataCache struct {
	plugin     map[string]incoming.DataTypeInterface
	lastAccess int64
	maxTTL     int64
	lock       *sync.RWMutex
}

//NewCache   .. .
func NewCache(maxttl int64) IncomingDataCache {
	if maxttl == 0 {
		maxttl = MAXTTL
	}
	return IncomingDataCache{
		hosts:  make(map[string]*ShardedIncomingDataCache),
		maxTTL: maxttl,
		lock:   new(sync.RWMutex),
	}
}

//NewShardedIncomingDataCache   .
func NewShardedIncomingDataCache(maxttl int64) *ShardedIncomingDataCache {
	return &ShardedIncomingDataCache{
		plugin: make(map[string]incoming.DataTypeInterface),
		maxTTL: maxttl,
		lock:   new(sync.RWMutex),
	}

}

//FlushAll Flush raw meterics data
func (i *IncomingDataCache) FlushAll() {
	lock, allHosts := i.GetHosts()
	defer lock.Unlock()
	willDelete := []string{}
	for key, plugin := range allHosts {
		//fmt.Fprintln(w, hostname)
		plugin.FlushAllMetrics()
		//this will clean up all zero plugins
		if plugin.Size() == 0 {
			willDelete = append(willDelete, key)
		}
	}
	for _, key := range willDelete {
		delete(allHosts, key)
		log.Printf("Cleaned up host for %s", key)
	}
}

//Put   ..
func (i IncomingDataCache) Put(key string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.hosts[key] = NewShardedIncomingDataCache(i.maxTTL)
}

// GetHosts locks the cache and returns the whole cache together with the lock. Caller needs
// to explicitly unlocks after the operation is done.
func (i IncomingDataCache) GetHosts() (*sync.RWMutex, map[string]*ShardedIncomingDataCache) {
	i.lock.Lock()
	return i.lock, i.hosts
}

//GetLastAccess ..Get last access time ...
func (shard *ShardedIncomingDataCache) GetLastAccess() int64 {
	return shard.lastAccess
}

//Expired  ... add expired test
func (shard *ShardedIncomingDataCache) Expired() bool {
	//clean up if data is not access for max TTL specified
	if time.Now().Unix()-shard.GetLastAccess() > int64(shard.maxTTL) {
		return true
	}
	return false
}

//GetShard  ..
func (i IncomingDataCache) GetShard(key string) *ShardedIncomingDataCache {
	//GetShard .... add shardGetCollectD
	//i.lock.Lock()
	if i.hosts[key] == nil {
		i.Put(key)
	}

	return i.hosts[key]

}

//GetData   ..
func (shard *ShardedIncomingDataCache) GetData(pluginname string) incoming.DataTypeInterface {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return shard.plugin[pluginname]
}

//Size no of plugin per shard
func (i IncomingDataCache) Size() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.hosts)

}

//Size no of plugin per shard
func (shard *ShardedIncomingDataCache) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.plugin)

}

//SetData  TODO : add generic
func (shard *ShardedIncomingDataCache) SetData(data incoming.DataTypeInterface) error {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.plugin[data.GetItemKey()] == nil {
		//TODO: change this to more generic later
		shard.plugin[data.GetItemKey()] = incoming.NewInComing(incoming.COLLECTD)
	}
	shard.lastAccess = time.Now().Unix()
	collectd := shard.plugin[data.GetItemKey()]
	collectd.SetData(data)
	return nil

	//return errors.New("unknow data type while setting data")

}

//CacheServer   ..
type CacheServer struct {
	cache              IncomingDataCache
	ch                 chan *IncomingBuffer
	mincollectinterval int
}

//GetCache  Get All hosts
func (cs *CacheServer) GetCache() *IncomingDataCache {
	return &cs.cache
}

//NewCacheServer   ...
func NewCacheServer(maxTTL int64, debug bool) *CacheServer {
	server := &CacheServer{
		cache: NewCache(maxTTL),
		ch:    make(chan *IncomingBuffer, 1000),
	}
	if debug {
		debugc = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

//Put   ..
func (cs *CacheServer) Put(incomingData incoming.DataTypeInterface) {
	var buffer *IncomingBuffer
	select {
	case buffer = <-freeList:
		//go one from buffer
	default:
		buffer = new(IncomingBuffer)
	}
	buffer.data = incomingData
	cs.ch <- buffer

}

func (cs CacheServer) close() {
	<-quitCacheServerCh
	close(quitCacheServerCh)
}
func (cs CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
LOOP:
	for {
		// Reuse buffer if there's room.
		buffer := <-cs.ch
		shard := cs.cache.GetShard(buffer.data.GetKey())
		shard.SetData(buffer.data)
		select {
		case freeList <- buffer:
		// Buffer on free list; nothing more to do.
		case <-quitCacheServerCh:
			break LOOP
		default:
			// Free list full, just carry on.
		}
		/*select {
		case data := <-s.ch:
			//fmt.Printf("got message in channel %v", data)
			shard := s.cache.GetShard(data.collectd.Host)
			shard.SetCollectD(data.collectd)

		}*/
	}

}

//GenrateSampleData  ....
func (cs *CacheServer) GenrateSampleData(key string, itemCount int, datatype incoming.DataTypeInterface) {
	//100 plugins
	for j := 0; j < itemCount; j++ {
		pluginname := fmt.Sprintf("%s_%d", "plugin_name_", j)
		debugc("Debug:Pluginname %s\n", pluginname)
		//. defer wg.Done()
		newSample := datatype.GenerateSampleData(key, pluginname)
		debugc("Debug:Sample %#v\n", newSample)

		cs.Put(newSample)

	}

}
