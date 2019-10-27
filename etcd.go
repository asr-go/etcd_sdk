package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"golang.org/x/net/context"
)

var (
	dialTimeout    = time.Second * 10 // 拨号超时
	contextTimeout = time.Second * 5  // 连接超时
	reloadDelay    = time.Second * 5  // 加载延迟
)

// EtcdLoader ...
type EtcdLoader struct {
	loaded         bool   // 已加载
	etcdEndpoints  string // ETCD 节点
	etcdConfigPath string // ETCD 配置路径

	// Cnf ...
	Cnf interface{}
}

func init() {
}

// NewConfig 加载参数
func (l *EtcdLoader) NewConfig(etcdEndpoints string, etcdConfigPath string, cnf interface{}) {
	// ETCD 节点
	l.etcdEndpoints = etcdEndpoints
	// ETCD 配置路径
	l.etcdConfigPath = etcdConfigPath

	// 是否已加载
	if l.loaded {
		cnf = l.Cnf
	}

	// 是否未加载
	if !l.loaded {
		// 加载配置
		err := l.loadConfig(cnf)

		if err != nil {
			logrus.Fatal(err)
			os.Exit(1)
		}

		// 更新配置
		l.refreshConfig(cnf)

		// 设置已加载
		l.loaded = true
		logrus.Info("配置加载完成！")
	}

	// 监听配置变化
	go l.keepalive(cnf)
}

// LoadConfig 加载配置并设置
func (l *EtcdLoader) loadConfig(cnf interface{}) error {
	// 新建 etcd客户端
	cli, err := newEtcdClient(l.etcdEndpoints, "", "", "")
	if err != nil {
		return err
	}
	defer cli.Close()

	// 加载配置
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	resp, err := cli.Get(ctx, l.etcdConfigPath)
	cancel()

	if err != nil {
		switch err {
		case context.Canceled:
			return fmt.Errorf("上下文被另外的协程关闭: %v", err)
		case context.DeadlineExceeded:
			return fmt.Errorf("上下文超时: %v", err)
		case rpctypes.ErrEmptyKey:
			return fmt.Errorf("客户端错误: %v", err)
		default:
			return fmt.Errorf("etcd节点错误: %v", err)
		}
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("不存在对应的配置: %s", l.etcdConfigPath)
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, cnf); err != nil {
		return err
	}

	return nil
}

// RefreshConfig 更新配置
func (l *EtcdLoader) refreshConfig(cnf interface{}) {
	l.Cnf = cnf
}

func (l *EtcdLoader) keepalive(cnf interface{}) {
	for {
		// 监听延迟
		<-time.After(reloadDelay)

		// 尝试重新加载配置
		err := l.loadConfig(cnf)
		if err != nil {
			logrus.Error(err)
			continue
		}

		// 更新配置
		l.refreshConfig(cnf)

		// 设置已加载
		l.loaded = true
	}
}

func newEtcdClient(theEndpoints, certFile, keyFile, caFile string) (*clientv3.Client, error) {
	// ETCD config
	etcdConfig := clientv3.Config{
		Endpoints:   strings.Split(theEndpoints, ","),
		DialTimeout: dialTimeout,
	}

	// ETCD client
	return clientv3.New(etcdConfig)
}
