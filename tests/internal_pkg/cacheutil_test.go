package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/infrawatch/smart-gateway/internal/pkg/cacheutil"
	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/stretchr/testify/assert"
)

/*----------------------------- helper functions -----------------------------*/

//GenerateSampleCacheData  ....
func GenerateSampleCacheData(cs *cacheutil.CacheServer, key string, itemCount int) {
	for j := 0; j < itemCount; j++ {
		pluginname := fmt.Sprintf("plugin_name_%d", j)
		// see incoming_collectd_test.go
		newSample := GenerateSampleCollectdData(key, pluginname)
		cs.Put(newSample)
	}
}

/*----------------------------------------------------------------------------*/

func TestCacheServer(t *testing.T) {
	pluginCount := 10
	hostname := "hostname"
	server := cacheutil.NewCacheServer(0, true)
	dataCache := server.GetCache()
	t.Run("Test IncommingDataCache", func(t *testing.T) {
		// test cache Size (1 hor each host)
		GenerateSampleCacheData(server, hostname, pluginCount)
		time.Sleep(time.Millisecond)
		assert.Equal(t, 1, dataCache.Size())
		GenerateSampleCacheData(server, "hostname2", 1)
		time.Sleep(time.Millisecond)
		assert.Equal(t, 2, dataCache.Size())
		// test shard Size (1 for each record for each host)
		assert.Equal(t, pluginCount, dataCache.GetShard(hostname).Size())
		assert.Equal(t, 1, dataCache.GetShard("hostname2").Size())
		// test SetData
		sample := GenerateSampleCollectdData("shardtest", "foo")
		key := sample.GetItemKey()
		shard := dataCache.GetShard("hostname2")
		assert.Equal(t, 1, shard.Size())
		shard.SetData(sample)
		assert.Equal(t, 2, shard.Size())
		//TODO(mmagr): below is not right (expected should be "hostname2" or sample
		//             should be saved to appropriate shard), but will fix that in later release
		data := shard.GetData(key)
		assert.NotNil(t, data)
		assert.Equal(t, "shardtest", data.(*incoming.CollectdMetric).Host)
	})
}

func TestCacheServerCleanUp(t *testing.T) {
	pluginCount := 10
	hostname := "hostname"
	server := cacheutil.NewCacheServer(1, true)
	dataCache := server.GetCache()

	t.Run("Test FlushAll", func(t *testing.T) {
		GenerateSampleCacheData(server, hostname, pluginCount)
		time.Sleep(time.Millisecond)
		assert.Equal(t, 1, dataCache.Size())
		assert.Equal(t, pluginCount, dataCache.GetShard(hostname).Size())
		shard := dataCache.GetShard("hostname")
		assert.Equal(t, false, shard.Expired())
		time.Sleep(time.Second * 2)
		assert.Equal(t, true, shard.Expired())
		dataCache.FlushAll()
		//TODO(mmagr): does it make sense to need to call FlushAll twice to first mark
		//             records as not new and delete it only after second call?
		assert.Equal(t, 1, dataCache.Size())
		dataCache.FlushAll()
		assert.Equal(t, 0, dataCache.Size())
	})
}
