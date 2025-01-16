# 项目目标
分析 qcow2 格式文件，获取 qcow2 相关信息
# 需求来源
近期的项目在 qcow2 over block 场景需要获取 qcow2 使用量来扩容 block 大小以达到精简置备的效果，但遇到如下问题：
1. 官方 qemu-img info 工具不能获取到 qcow2 实际占用空间
2. 官方 qemu-img check 可以获取到 qcow2 image_end_offset，镜像偏移量，实际也是使用量，但在镜像使用比较大或者存在 backing_file 以及加密的情况下，命令耗时太久

## 二进制编译使用
```
go build -o qcow2-analyze main.go
```
```
root@ubuntu:~/workspace/awesomeGo/src/qcow2-analyze# ./qcow2-analyze  -h
analyze a qcow2 file with output

Usage:
  qcow2-analyze [flags]

Flags:
  -f, --file string     qcow2 file path
  -h, --help            help for qcow2-analyze
  -H, --hex             hexadecimal output, not supported now
  -o, --output string   output format, only json now (default "raw")
  -v, --verbose         verbose output
```

