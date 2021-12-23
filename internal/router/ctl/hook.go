package ctl

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"project-ci/internal/publisher"
	"project-ci/internal/schema"
	"project-ci/internal/task"
	"strings"
)

type Hook struct {
	projects map[string]*schema.Project
	taskQ    *task.TaskQueue
	pub      *publisher.Broker
}

func NewHook(projects map[string]*schema.Project,
	broker *publisher.Broker,
	queue *task.TaskQueue) *Hook {
	return &Hook{projects: projects, taskQ: queue, pub: broker}
}

// Hook 接收web hook
func (a *Hook) Hook(r *ghttp.Request) {
	name := r.GetString("name")
	glog.Info("项目名称:" + name)
	if name == "" {
		r.Response.WriteExit("项目名称为空")
	}
	if _, ok := a.projects[name]; !ok {
		r.Response.WriteExit("没有找到项目:" + name)
	}
	project := a.projects[name]
	body, err := r.GetJson()
	if err != nil {
		r.Response.WriteExit(err.Error())
		return
	}
	// 判断分支
	ref := body.GetString("ref")
	ref = strings.TrimPrefix(ref, "refs/heads/")
	glog.Info("当前分支:" + ref)
	project.Pusher = body.GetString("user_name")
	glog.Info("提交人:" + project.Pusher)
	if ref == project.Ref {
		a.taskQ.Insert(project)
	}
}

// GetList 获取项目列表
func (a *Hook) GetList(r *ghttp.Request) {
	_ = r.Response.WriteTpl("index.html", g.Map{
		"projects": a.projects,
	})
}

// Monitor 监控项目编译输出
func (a *Hook) Monitor(r *ghttp.Request) {
	name := r.GetString("name")
	_ = r.Response.WriteTpl("monitor.html", g.Map{
		"name": name,
	})
}

// WebSocket 通知执行结果
func (a *Hook) WebSocket(r *ghttp.Request) {
	ws, err := r.WebSocket()
	if err != nil {
		r.Response.WriteExit("不是websocket请求")
	}

	sub := &publisher.Sub{
		MessageChan: make(chan *publisher.GitMessage),
		Name:        r.GetString("name"),
		Close:       make(chan struct{}),
	}
	a.pub.Sub(r.GetString("name"), sub)
	disconnectChan := make(chan struct{})
	// read
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				close(disconnectChan)
				return
			}
		}
	}()
	for {
		select {
		case msg := <-sub.MessageChan:
			err := ws.WriteJSON(msg)
			if err != nil {
				break
			}
		case <-disconnectChan:
			close(sub.Close)
			return
		}
	}
}
