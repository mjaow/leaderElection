package election

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"

	log "github.com/sirupsen/logrus"
)

type leaderElection struct {
	etcdCli          *clientv3.Client
	onStartedLeading func(stopCh chan struct{})
	key              string
	id               string
	ttl              int
}

func NewLeaderElection(
	etcdCli *clientv3.Client,
	onStartedLeading func(stopCh chan struct{}),
	key,
	id string,
	ttl int) *leaderElection {
	return &leaderElection{
		etcdCli:          etcdCli,
		onStartedLeading: onStartedLeading,
		key:              key,
		id:               id,
		ttl:              ttl,
	}
}

func (le *leaderElection) campaign(electCh chan *concurrency.Election) {
	for {
		s, err := concurrency.NewSession(le.etcdCli, concurrency.WithTTL(le.ttl))
		if err != nil {
			log.Errorf("new session with error %v", err)
			continue
		}
		e := concurrency.NewElection(s, le.key)

		log.Infof("[%s]campaign for leader", le.id)
		if err = e.Campaign(context.TODO(), le.id); err != nil {
			log.Errorf("campaign with error %v", err)
			continue
		}

		resp, err := e.Leader(context.TODO())

		if err != nil {
			log.Errorf("get leader with error %v", err)
			continue
		}

		val := string(resp.Kvs[0].Value)
		if val != le.id {
			log.Errorf("[%s]i'm not leader actual %s", le.id, val)
			continue
		}

		log.Infof("[%s]elect: success", val)
		electCh <- e
		select {
		case <-s.Done():
		}
	}
}

func (le *leaderElection) Run() {
	electCh := make(chan *concurrency.Election, 1)

	// 选主逻辑
	go le.campaign(electCh)

	<-electCh

	var stopCh = make(chan struct{})

	// 初始化controller manager主流程
	le.onStartedLeading(stopCh)
}
