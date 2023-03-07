# Im2.0

项目代号: zim

Another chat messenger server written by Go.

按: 之前工作中做的一个实验项目, 最终没有被采纳. 虽然还是有一些瑕疵, 不过写的还是蛮认真的, 
性能也还可以, 就厚着脸皮放出来了. 并发还可以, 支持超级群. 性能报告参见BENCHMARK.txt

### 构建docker镜像
```
docker build \
    --build-arg APP=zimapi \
    --build-arg PORT=1840 \
    -t swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimapi .

docker build \
    --build-arg APP=zimpush \
    --build-arg PORT=1937 \
    -t swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimpush .

docker build \
    --build-arg APP=zimcron \
    --build-arg PORT=1860 \
    -t swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimcron .
```
### 提交到华为云docker仓库
```
docker push swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimapi
docker push swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimpush
docker push swr.cn-north-1.myhuaweicloud.com/zh0mmc/zimcron
```
