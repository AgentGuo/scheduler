# scheduler
目前先简单实现一个单体调度器（scheduler-main），划分了四个文件夹：
* cmd 程序入口，使用cobra库 √
  * scheduler-main 主调度器入口
  * metrics-cli 节点监测器入口
* pkg 具体实现
  * 主调度器 
    * 任务调度 √
    * 任务绑定 √
  * 节点检测器
    * 节点状态监测 √ 
* task 调度任务相关
  * task 任务定义 √
  * task 调度队列 √
* util 工具库
---
redis存储内容：

| key         | subKey                | 说明         |
|-------------|-----------------------|------------|
| metricsInfo | {主机名}                 | 存储动态监控信息   |
| nodeInfo    | {主机名}                 | 存储主机静态资源信息 |
| taskInfo    | {podName}-{namespace} | 调度结果信息     |

---
TODO:
- [x] 节点标签
- [x] 插件化
- [ ] 更新一下文档
---
使用方法：
1. 启动redis服务器(启动失败大概率是端口占用了`ps -elf | grep redis`, `kill $pid`)
```bash
redis-server ./cmd/metrics-cli/redis.conf
```
2. 各个集群节点启动资源监控进程
```bash
cd cmd/metrics-cli/ && go build -v
./metrics-cli -f config.yaml
```
3. 启动调度器进程
```bash
cd cmd/scheduler-main/ && go build -v
./scheduler-main -f config.yaml
```
4. 启动测试pod`kubectl create -f test_pod.yaml`(目前特化的调度逻辑只会调度default下的名为pause的pod，见https://github.com/AgentGuo/scheduler/blob/main/pkg/schedulermain/schedulermain.go#L49)， test_pod.yaml如下：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pause
  namespace: default
spec:
  schedulerName: my-scheduler
  containers:
  - name: pause
    image: registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.5
    command:
      - /pause
    imagePullPolicy: IfNotPresent
  restartPolicy: Always
```