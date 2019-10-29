package election

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

type item struct {
	cli *clientv3.Client
	id  string
}

func NewEtcdCliOrDie(endpoints [] string, dialTimeoutInSeconds, autoSyncIntervalInSeconds int) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      time.Second * time.Duration(dialTimeoutInSeconds),
		AutoSyncInterval: time.Second * time.Duration(autoSyncIntervalInSeconds),
	})

	if err != nil {
		log.Fatal(err)
	}

	return cli
}

func TestLeaderElection_Run(t *testing.T) {
	cli := NewEtcdCliOrDie([]string{"127.0.0.1:2379"}, 3, 3)
	defer cli.Delete(context.Background(), "/", clientv3.WithPrefix())

	var keyPrefix = "/my-election"
	var ttl = 5

	var leader = make(chan item, 1)

	for i := 1; i <= 3; i++ {
		go func(idx int) {
			cli := NewEtcdCliOrDie([]string{"127.0.0.1:2379"}, 3, 3)

			id := fmt.Sprintf("node%d", idx)

			if idx%2 == 1 {
				log.Infof("%s sleep 1s", id)
				time.Sleep(time.Second)
			}

			le := NewLeaderElection(cli, func(stopCh chan struct{}) {
				leader <- item{
					id: id,
				}
			}, keyPrefix, id, ttl)
			le.Run()
		}(i)
	}

	if rs := <-leader; rs.id != "node2" {
		t.Fatalf("leader expected node2 and actual %s", rs.id)
	}
}
