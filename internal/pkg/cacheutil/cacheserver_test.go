package cacheutil

import (
	"context"
	"testing"
	"time"

	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/incoming"
)

func TestCacheServer(t *testing.T) {
	//ch:=  make(chan IncomingBuffer)
	hostname := "host"
	collectd := incoming.NewInComing(incoming.COLLECTD)
	newSample := collectd.GenerateSampleData(hostname, "pg")

	if newSample.GetKey() != hostname {
		t.Errorf("Data Key is not matching , expected %s and got %s", hostname, newSample.GetKey())
	}

}

func TestCacheServer2(t *testing.T) {
	pluginCount := 10
	hostname := "hostname"
	ctx := context.Background()
	//	var hostCount=1
	//	var freeListToCollectSample = make(chan *IncomingBuffer, 100)

	//  collectd:=incoming.NewInComing(incoming.COLLECTD)
	server := NewCacheServer(ctx, 0, true)
	collectd := incoming.NewInComing(incoming.COLLECTD)
	server.GenrateSampleData(hostname, pluginCount, collectd)

	time.Sleep(time.Second * 2)

	incomingDataCache := server.GetCache()
	if size := incomingDataCache.Size(); size != 1 {
		t.Errorf("wrong count of host , expected 1 and got %d", size)
	}
	if size := incomingDataCache.GetShard(hostname).Size(); size != pluginCount {
		t.Errorf("wrong count of plugin per host , expected %d and got %d", pluginCount, size)
	}

}

func TestCacheServerForCleanUP(t *testing.T) {
	pluginCount := 10
	hostname := "hostname"
	ctx := context.Background()
	//	var hostCount=1
	//	var freeListToCollectSample = make(chan *IncomingBuffer, 100)

	//  collectd:=incoming.NewInComing(incoming.COLLECTD)
	server := NewCacheServer(ctx, 4, true)
	collectd := incoming.NewInComing(incoming.COLLECTD)
	server.GenrateSampleData(hostname, pluginCount, collectd)
	incomingDataCache := server.GetCache()
	time.Sleep(time.Second * 2)
	incomingDataCache.FlushAll()
	if size := incomingDataCache.Size(); size != 1 {
		t.Errorf("wrong count of host , expected 1 and got %d", size)
	}
	if size := incomingDataCache.GetShard(hostname).Size(); size != pluginCount {
		t.Errorf("wrong count of plugin per host , expected %d and got %d", pluginCount, size)
	}
	time.Sleep(time.Second * 5)
	incomingDataCache.FlushAll()
	if size := incomingDataCache.Size(); size != 0 {
		t.Errorf("wrong count of host , expected 0 and got %d", size)
	}

}
