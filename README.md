# ETCD SDK

`ETCD`SDK

## 快速上手

``` go
package main

import (
  etcdsdk "github.com/asr-go/etcdsdk"
)

func main() {
  cnf := new(interface{})
  loader := etcdsdk.EtcdLoader{}
  loader.NewConfig(cnf)
}
```
