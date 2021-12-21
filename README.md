# 项目CI工具


## 项目编译监控
为方便查询CI工具的执行日志，保证项目正常运行，提供一个项目监控页面
[点击监控](http://127.0.0.1:10090/web/projects)

## 使用说明

[comment]: <> (首先安装[gxt-cli工具]&#40;http://127.0.0.1/project-cli.git&#41;)
* 调试项目
`gxt run main.go`
* 编译项目
`gxt build` 编译后的文件在bin目录下

CI工具在运行时，监听到项目的hook消息后，工作流如下:
* 有源码目录的情况: `切换工作目录-->拉取分支代码-->切换相应分支-->执行构建前脚本-->构建-->执行构建后脚本`
* 无源码目录的情况: `克隆项目源码-->切换工作目录-->拉取分支代码-->切换相应分支-->执行构建前脚本-->构建-->执行构建后脚本`

> 目前因为项目构建差异过大，各个构建过程在脚本内完成

### 配置文件说明
config.json 结构如下:
``` json
{
  "server": {
    "port": 10090,
    "default_source_name": "source"
  },
  "projects": [
    {
      "name": "release",
      "ref": "dev",
      "path": "/root/projects/project-release-api",
      "repo": "https://gitlab.workai.com/test/gxt-release-api.git",
      "before_shell": "",
      "after_shell": ""
    }
  ]
}
```

|  字段名 | 含义 |
|---|---|
|  server.port  |当前CI服务监听端口|
|  server.default_source_name  |默认的源码目录名称|
| projects | 项目列表|

* projects 字段说明

|  字段名 | 含义 |
|---|---|
|  name  |项目名称(对应hook url中的项目参数)|
| ref | 分支名称(如:master,dev等)|
| path | 本地工作目录,包含项目源码及脚本等|
| repo | 仓库地址 |
| before_shell | 构建前执行的脚本,不填不执行 |
| after_shell | 构建后脚本,不填不执行 |


### 开发服务器项目目录配置
理论上一个项目一个文件夹，此文件夹中包含一份项目源码(用于自动编译构建),N个脚本文件,用于Hook后，编译前或编译后运行

### Web Hook配置
* HOOK URL: `http://127.0.0.1:10090/api/v1/hook/{项目名称}`, 其中的`{项目名称}`对应projects.name字段

> 进入gitlab-项目仓库-仓库设置-管理WEB钩子-添加钩子`[gitlab]`,推送地址填HookURL
> 数据格式：application/json
> 请设置您希望触发 Web 钩子的事件：选择`[只推送push事件]`
> 点击『添加WEB钩子』

### 举个例子
假设一个项目的配置如下:
```json
{
      "name": "release",
      "ref": "dev",
      "path": "/root/services/project-release-api",
      "repo": "https://127.0.0.1/project-release-api.git",
      "after_shell": "build.sh"
    }
```
build.sh脚本内容如下:
```shell script
  cd source
  go build -o release-server
  echo "构建完成"
  cp release-server ../release-server2
  cd ../
  pid=`echo $(pgrep release-server)`
  echo $pid
  kill $pid
  chmod +x release-server2
  mv release-server2 release-server
  nohup ./release-server > nohup.out 2>&1 &
```