package main

import (
	"fmt"

	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"

	"sync"
	"testing"
	"time"
)

func BenchmarkWithSampleData(b *testing.B) {
	/* example
	b.RunParallel(func(pb *testing.PB) {
	        for pb.Next() {
	            SomeFunction()
	        }
	    })
	*/
	fmt.Printf("\nTest data  will run for %d times\n ", b.N)
	var cacheServer *cacheutil.CacheServer
	cacheServer = cacheutil.NewCacheServer(0, false)
	incomingType := incoming.NewInComing(incoming.COLLECTD)
	for hostid := 0; hostid < b.N; hostid++ {
		var hostname = fmt.Sprintf("%s_%d", "redhat.boston.nfv", hostid)
		go cacheServer.GenrateSampleData(hostname, 100, incomingType)
	}

}

func TestPut(t *testing.T) {
	var cacheserver = cacheutil.NewCacheServer(0, false)
	var noofiteration, noofhosts, noofpluginperhosts int = 4, 2, 100

	var hostwaitgroup sync.WaitGroup

	for times := 1; times <= noofiteration; times++ {
		hostwaitgroup.Add(noofhosts)
		for hosts := 0; hosts < noofhosts; hosts++ {
			go func(host_id int) {
				defer hostwaitgroup.Done()
				//100o hosts
				//pluginChannel := make(chan cacheutil.Collectd)
				//for each host make it on go routine
				var hostname = fmt.Sprintf("%s_%d", "redhat.boston.nfv", host_id)
				//fmt.Printf("Iteration %d hostname %s\n",times,hostname)

				collectd := incoming.NewInComing(incoming.COLLECTD)
				cacheserver.GenrateSampleData(hostname, noofpluginperhosts, collectd)

			}(hosts)

		}
		/*for _,shard :=range cacheserver.cache.hosts{
		  fmt.Printf("Whole map %d",len(shard.plugin))
		  }*/
		hostwaitgroup.Wait()
		time.Sleep(time.Second * 1)

		if size := cacheserver.GetCache().Size(); size != noofhosts {
			t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
		}
		for hostname, plugins := range cacheserver.GetCache().GetHosts() {
			if size := plugins.Size(); size != 100 {
				t.Errorf("wrong count of plugin per host %s, expected 100 and got %d", hostname, size)
			}
		}

	}
	//after everything is done
	if size := cacheserver.GetCache().Size(); size != noofhosts {
		t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
	}
}
