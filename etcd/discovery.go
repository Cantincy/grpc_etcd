package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	cli        *clientv3.Client  // etcd连接
	serviceMap map[string]string // 服务列表(k-v列表)
	lock       sync.RWMutex      // 读写互斥锁
}

func NewServiceDiscovery(endpoints []string) (*ServiceDiscovery, error) {
	// 创建etcdClient对象
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return &ServiceDiscovery{
		cli:        cli,
		serviceMap: make(map[string]string), // 初始化kvMap
	}, nil
}

// WatchService 读取etcd的服务并开启监视协程
func (s *ServiceDiscovery) WatchService(prefix string) error {
	// 根据服务名称的前缀，获取所有的注册服务
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// 遍历key-value存储到本地map
	for _, kv := range resp.Kvs {
		s.SetService(string(kv.Key), string(kv.Value))
	}

	// 开启监听协程，监听prefix的变化
	go s.watcher(prefix)

	return nil
}

// watcher 监听prefix
func (s *ServiceDiscovery) watcher(prefix string) {
	watchRespChan := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for resp := range watchRespChan {
		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.PUT: // 发生了修改或者新增
				s.SetService(string(event.Kv.Key), string(event.Kv.Value)) // ServiceMap中进行相应的修改或新增
			case mvccpb.DELETE: //发生了删除
				s.DelService(string(event.Kv.Key)) // ServiceMap中进行相应的删除
			}
		}
	}
}

// SetService 新增或修改本地服务
func (s *ServiceDiscovery) SetService(key, val string) {
	s.lock.Lock()
	s.serviceMap[key] = val
	s.lock.Unlock()
	log.Println("put key :", key, "val:", val)
}

// DelService 删除本地服务
func (s *ServiceDiscovery) DelService(key string) {
	s.lock.Lock()
	delete(s.serviceMap, key)
	s.lock.Unlock()
	log.Println("del key:", key)
}

// GetService 获取本地服务
func (s *ServiceDiscovery) GetService(serviceName string) (string, error) {
	s.lock.RLock()
	serviceAddr, ok := s.serviceMap[serviceName]
	s.lock.RUnlock()
	if !ok {
		return "", fmt.Errorf("can not get serviceAddr")
	}
	return serviceAddr, nil
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
