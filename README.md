# qqwry
纯真IP，golang 内存版

> * 2021-09-09
> * 初始化时优先加载本地磁盘数据，减少服务冷启动时间
> * 初始化完成后，无需等待定时更新任务，立刻更新一次在线数据
> * 每次更新在线数据库时，将新数据覆盖写入至本地磁盘

## 1.依赖
> * 参照了 https://github.com/yinheli/qqwry 的代码，虽然文件操作也比较快，还是习惯用内存。所以算法不变，把ip搜索从文件改为内存
> * 依赖 https://github.com/yinheli/mahonia


## 2.使用
* 直接使用全局变量查询
```golang
package main

import (
    "qqwry"
    "fmt"
)

func main() {
    result, err := qqwry.Find("202.106.0.20")
    if err!=nil {
        panic(err)
    }
    fmt.Printf("%+v\n", result)
}
```

* 初始化局部变量使用
```golang
package main

import (
    "qqwry"
    "fmt"
)

func main() {
    var result *qqway.Rq
    var err error
    var q *qqwry.QQwry
    if q, err = NewQQwry(); err != nil{
        panic(err)
    }
    
    result, err = q.Find("202.106.0.20")
    if err!=nil {
        panic(err)
    }
    fmt.Printf("%+v\n", result)
}
```

## 3.注意
> 线程安全，缓存操作。方便做服务器来查询。

> 每小时从官方地址获取新的纯真库文件(`如果太频繁, 可能会被屏蔽掉`), 如果发现不一致, 则进行更新。
