package etcd

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

// ServiceRegister 服务注册
type ServiceRegister struct {
	etcdCli *clientv3.Client   // etcdClient对象
	leaseId clientv3.LeaseID   // 租约id
	ctx     context.Context    // context上下文
	cancel  context.CancelFunc // context终止方法
}

// CreateLease 创建租约。expire表示有效期(s)
func (s *ServiceRegister) CreateLease(expire int64) error {

	lease, err := s.etcdCli.Grant(s.ctx, expire)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	s.leaseId = lease.ID // 记录生成的租约Id
	return nil
}

// BindLease 绑定租约。将租约与对应的key-value绑定
func (s *ServiceRegister) BindLease(key string, value string) error {

	res, err := s.etcdCli.Put(s.ctx, key, value, clientv3.WithLease(s.leaseId))
	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Printf("bind lease success %v \n", res)
	return nil
}

// KeepAlive 获取续租的通道
func (s *ServiceRegister) KeepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	resChan, err := s.etcdCli.KeepAlive(s.ctx, s.leaseId)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return resChan, nil
}

// Watcher 监听并续约
func (s *ServiceRegister) Watcher(resChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case l := <-resChan:
			log.Printf("续约成功,val:%+v \n", l)
		case <-s.ctx.Done():
			log.Printf("续约关闭")
			return
		}
	}
}

// Close 关闭服务
func (s *ServiceRegister) Close() error {
	s.cancel()
	log.Printf("closed...\n")
	// 撤销租约
	s.etcdCli.Revoke(s.ctx, s.leaseId)
	return s.etcdCli.Close()
}

// NewEtcdRegister 初始化etcd服务注册对象
func NewEtcdRegister(etcdServerAddr string) (*ServiceRegister, error) {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdServerAddr},
		DialTimeout: 3 * time.Second,
	})

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	svr := &ServiceRegister{
		etcdCli: client,
		ctx:     ctx,
		cancel:  cancelFunc,
	}
	return svr, nil
}

// RegisterService 服务注册。expire表示过期时间,serviceName和serviceAddr分别是服务名与服务地址
func (svr *ServiceRegister) RegisterService(serviceName, serviceAddr string, expire int64) (err error) {

	// 创建租约
	err = svr.CreateLease(expire)
	if err != nil {
		return err
	}

	// 将租约与k-v绑定
	err = svr.BindLease(serviceName, serviceAddr)
	if err != nil {
		return err
	}

	// 获取续约通道
	keepAliveChan, err := svr.KeepAlive()
	if err != nil {
		return err
	}

	// 监听并续约
	go svr.Watcher(keepAliveChan)

	return nil
}
