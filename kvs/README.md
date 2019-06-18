# kvs - K/V Store Layer

It's an abstract layer for common operations of K/V stores like:

1. consul
2. etcd

TODO to support more backends could be a scheduling plan.



## Usages

Main interface is lied in [`store.KVStore`](store/store.go#L35)  

We make new store.KVStore with `kvs.New()` call:

```go

store := kvs.New(&consul_util.ConsulConfig{
    Scheme: "http",
    Addr: "127.0.0.1:8500",
    Insecure: true,
    CertFile: "", KeyFile: "", CACertFile: "",
    Username: "", Password: "",
    Root: "",
    DeregisterCriticalServiceAfter: "30s",
})
defer store.Close() // optional

// or
storeEtcd := kvs.New(
    &etcd.Etcdtool{
        Peers: "127.0.0.1:2379",
        Cert: "",
        Key: "", CA: "", User: "",
        Timeout: time.Second * 10,
        CommandTimeout: time.Second * 5,
        Routes: []etcd.Route{},
        PasswordFilePath: "",
        Root: "",
    }
)
defer storeEtcd.Close() // optional

```

Once `store` got, we can get/set key/value pair:

```go
store.Put("x", "yz")
log.Info(store.Get("x"))

// Or, with yaml encode/decode transparently
state := map[string]string{
    "ab": "111",
    "cd": "222",
}
store.PutYaml("state", state)
state1 := store.GetYaml("state")
log.Infof("state1: %v", state1)

// And, with path separatly
store.Put("config/gwapi/enabled", "true")

// And, delete them:
store.Delete("x")
store.Delete("state")
store.Delete("config/gwapi")
```

For more usages, see also `consul/core_test.go` and `etcd/i_test.go`, ....

