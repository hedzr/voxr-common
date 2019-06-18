/*
 * Copyright © 2019 Hedzr Yeh.
 */

package etcd

import (
	"github.com/hedzr/voxr-common/kvs/store"
	"go.etcd.io/etcd/clientv3"
)

type MyKvPair struct {
	Key1 []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// create_revision is the revision of last creation on this key.
	CreateRevision int64 `protobuf:"varint,2,opt,name=create_revision,json=createRevision,proto3" json:"create_revision,omitempty"`
	// mod_revision is the revision of last modification on this key.
	ModRevision int64 `protobuf:"varint,3,opt,name=mod_revision,json=modRevision,proto3" json:"mod_revision,omitempty"`
	// version is the version of the key. A deletion resets
	// the version to zero and any modification of the key
	// increases its version.
	Version int64 `protobuf:"varint,4,opt,name=version,proto3" json:"version,omitempty"`
	// value is the value held by the key, in bytes.
	Value1 []byte `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	// lease is the ID of the lease that attached to key.
	// When the attached lease expires, the key will be deleted.
	// If lease is 0, then no lease is attached to the key.
	Lease int64 `protobuf:"varint,6,opt,name=lease,proto3" json:"lease,omitempty"`
}

func (s *MyKvPair) Key() string {
	return string(s.Key1)
}

func (s *MyKvPair) Value() []byte {
	return s.Value1
}

func (s *MyKvPair) ValueString() string {
	return string(s.Value1)
}

type MyKvPairs struct {
	resp *clientv3.GetResponse
}

func (s *MyKvPairs) Count() int {
	return int(s.resp.Count) // len(s.pairs)
}

func (s *MyKvPairs) Item(index int) store.KvPair {
	keyValue := s.resp.Kvs[index]
	pp := MyKvPair{
		keyValue.Key,
		keyValue.CreateRevision,
		keyValue.ModRevision,
		keyValue.Version,
		keyValue.Value,
		keyValue.Lease,
	}
	return &pp
}
