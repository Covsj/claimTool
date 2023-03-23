# claimTool使用帮助

推荐使用nodereal RPC

注册：https://nodereal.io/invite/c19201ec-373e-419a-aa9d-d746e8e06b3b

ARB：arbitrum-nitro-rpc

如果需要claim之后在归集到一个地址 可以提pr或者联系我

***

## Arbitrum模块

### 1 执行go mod下载依赖

```sh
# 需要在项目目录内 不懂的可以自己谷歌go module依赖
go mod tidy
```

### 2.进入Arbitrum模块目录

```sh
# $pwd 为你下载或者克隆代码的目录
cd $pwd/claimTool/arbitrum
```

### 3.修改配置文件

```sh
# config.toml 是程序所需要依赖的配置文件 根据自己情况修改

PrivateKeys 是你想要执行程序的地址私钥数组 根据配置来
其他的配置文件 可改可不改
```

### 4.运行程序

- 4.1 claim任务
```
# 现在应该是在$pwd/claimTool/arbitrum目录下
运行命令：  go run main.go --config ./config.toml
```

- 4.2 归集任务
```
# 现在应该是在$pwd/claimTool/arbitrum目录下
运行命令：  go run main.go --config ./config.toml --task collect
```