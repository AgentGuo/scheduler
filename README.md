# scheduler
目前先简单实现一个单体调度器（scheduler-main），划分了四个文件夹：
* cmd 程序入口，使用cobra库 √
* pkg 具体调度器实现
  * 任务调度 ×
* task 调度任务相关
  * task 任务定义 √
  * task 调度队列 √
* util 工具库

目前在实现任务调度...