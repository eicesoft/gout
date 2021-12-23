package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const ReadTimeout = 5

type Client struct {
	cli       *clientv3.Client
	closeChan chan error
}

func NewClient(conf clientv3.Config) (client *Client, err error) {
	client = new(Client)
	client.cli, err = clientv3.New(conf)
	return
}

type KeyList struct {
	Key            string `json:"key,omitempty"`
	CreateRevision int64  `json:"create_revision,omitempty"`
	ModRevision    int64  `json:"mod_revision,omitempty"`
	Version        int64  `json:"version,omitempty"`
	Value          string `json:"value,omitempty"`
}

func (c *Client) Set(key string, val string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ReadTimeout)
	defer cancel()

	resp, err := c.cli.Put(ctx, key, val)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func (c *Client) List() []*KeyList {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ReadTimeout)
	defer cancel()

	resp, err := c.cli.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		panic(err)
	}

	ret := make([]*KeyList, len(resp.Kvs))
	for i := range resp.Kvs {
		kv := resp.Kvs[i]
		ret[i] = &KeyList{
			string(kv.Key),
			kv.CreateRevision,
			kv.ModRevision,
			kv.Version,
			string(kv.Value),
		}
	}

	return ret
}
