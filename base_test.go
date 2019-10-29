package election

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/embed"
)

func TestMain(m *testing.M) {
	StartEtcdServer()
	m.Run()
	defer StopEtcdServer()
}

var path = "/tmp/default.etcd"

var etcdServer *embed.Etcd

func StartEtcdServer() {
	start := time.Now()
	for {
		time.Sleep(time.Second * 1)
		cfg := embed.NewConfig()
		cfg.Dir = path
		e, err := embed.StartEtcd(cfg)
		if err != nil {
			if time.Now().Sub(start).Seconds() > 45 {
				panic(err)
			}

			if !strings.Contains(err.Error(), "address already in use") {
				panic(err)
			}
			continue
		}
		select {
		case <-e.Server.ReadyNotify():
			log.Printf("Server is ready!")
		}
		etcdServer = e
		return
	}
}

func StopEtcdServer() {
	etcdServer.Server.Stop()
	etcdServer.Close()
}
