# 项目目标
以只读方式分析 qcow2 格式文件，获取 qcow2 相关信息
# 需求来源
近期的项目在 qcow2 over block 场景需要获取 qcow2 使用量来扩容 block 大小以达到精简置备的效果，但遇到如下问题：
1. 官方 qemu-img info 工具不能获取到 qcow2 实际占用空间
```
# qemu-img info /dev/sp-7dtjui47yvj3pub9sdzl/pvc-a2633618-4491-4768-8871-996f0a0d3201
image: /dev/sp-7dtjui47yvj3pub9sdzl/pvc-a2633618-4491-4768-8871-996f0a0d3201
file format: qcow2
virtual size: 20 GiB (21474836480 bytes)
disk size: 0 B
cluster_size: 65536
Format specific information:
    compat: 1.1
    compression type: zlib
    lazy refcounts: false
    refcount bits: 16
    corrupt: false
    extended l2: false
```
2. 官方 qemu-img check 可以获取到 qcow2 image_end_offset，镜像偏移量，实际也是使用量，但在加密的情况下，命令耗时太久
```
# date; qemu-img check --object secret,id=sec0,data=MTczNzAzNTM5Mi1jcElTQVNMQVczV2ttd01abFROQg==,format=base64 --image-opts driver=qcow2,encrypt.key-secret=sec0,file.filename=/dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035794,backing.driver=qcow2,backing.file.filename=/dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035774,backing.encrypt.key-secret=sec0,backing.backing.driver=qcow2,backing.backing.file.filename=/dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035752,backing.backing.encrypt.key-secret=sec0,backing.backing.backing.driver=qcow2,backing.backing.backing.file.filename=/dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5,backing.backing.backing.encrypt.key-secret=sec0; date
Thu Jan 16 14:04:39 UTC 2025
No errors were found on the image.
1393/327680 = 0.43% allocated, 12.99% fragmented, 0.00% compressed clusters
Image end offset: 93061120
Thu Jan 16 14:04:48 UTC 2025
```

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
## 使用示例
```
# date; qcow2-analyze -f /dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035794; date
Thu Jan 16 14:05:56 UTC 2025
{
    "qcow_version": 3,
    "backing_file": "./pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035774",
    "cluster_size": 65536,
    "virtual_size": 21474836480,
    "crypt_method": "LUKS",
    "snapshots": null,
    "image_end_offset": 93061120,
    "author": "xuhao@cestc.cn"
}
Thu Jan 16 14:05:56 UTC 2025
```

```
# date; qcow2-analyze -f /dev/sp-7dtjui47yvj3pub9sdzl/pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035794 -v; date
Thu Jan 16 14:06:16 UTC 2025
{
    "qcow_version": 3,
    "backing_file": "./pvc-10b12db8-7c57-4af9-b101-8e6477ca2eb5-t1737035774",
    "cluster_size": 65536,
    "virtual_size": 21474836480,
    "crypt_method": "LUKS",
    "snapshots": null,
    "image_end_offset": 93061120,
    "author": "xuhao@cestc.cn",
    "l1_size": 40,
    "l1_table_offset": 196608,
    "refcount_table_offset": 65536,
    "refcount_table_clusters": 1,
    "dirty_bit": 0,
    "corrupt_bit": 0,
    "external_data_file": 0,
    "compression_type": 0,
    "extended_l2": 0,
    "lazy_refcount": 0,
    "bitmaps_extension": 0,
    "raw_external_data": 0,
    "refcount_order": 4,
    "header_length": 112,
    "compression_method": 0
}
Thu Jan 16 14:06:16 UTC 2025
```
