# ETCD SDK

`ETCD`SDK

## 快速上手

``` go
package main

import (
  etcd_sdk "github.com/asr-go/etcd_sdk"
)

func main() {
  cnf := new(interface{})
  loader := etcd_sdk.EtcdLoader{}
  loader.NewConfig(cnf)
}
```
