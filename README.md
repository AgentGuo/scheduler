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
- [ ] 亲和性
- [ ] 插件化