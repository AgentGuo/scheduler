# scheduler
目前先简单实现一个单体调度器（scheduler-main），划分了四个文件夹：
* cmd 程序入口，使用cobra库
  * scheduler-main 主调度器入口 √
  * metrics-cli 节点监测器入口 √
  * resourcemanage 资源动态修改器入口 √
  * submit-resourcetask 命令行提交资源任务工具
* pkg 具体实现
  * 主调度器 
    * 任务调度 √
    * 任务绑定 √
    * 区分任务类型 √
    * 提交资源任务 √
    * 资源检查
  * 节点检测器
    * 节点状态监测 √ 
  * 资源动态修改器
    * 接收资源任务 √ 
    * 同步k8s
* task 调度任务相关
  * task 任务定义 √
  * task 调度队列 √
* util 工具库
---
redis存储内容：

| key         | subKey                | 说明         |
|-------------|-----------------------|------------|
| metricsInfo | {主机名}                 | 存储动态监控信息(cpu:m, mem: byte)   |
| nodeInfo    | {主机名}                 | 存储主机静态资源信息(cpu:m, mem: byte) |
| taskInfo    | {podName}-{namespace} | task信息     |

---
```
常规任务类型:(redis中taskInfo)
task{
  info
  Detail: kubequeue.KubeTaskDetails
  ResourceDetail: apis.ResourceValue
}
资源任务类型:
task{
  info
  Detail: apis.KubeResourceTask
  ResourceDetail: nil
}
1.提交资源任务时, 指定KubeResourceTask中的字段内容, 由接口读入并判断合法性, 再封装成task，提交到调度队列中。
2.资源任务执行成功, 将KubeResourceTask中的资源限制更新到taskInfo中的ResourceValue。
```
---
TODO:
- [x] 节点标签
- [x] 插件化
- [ ] 更新一下文档