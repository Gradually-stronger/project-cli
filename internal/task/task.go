package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/container/gqueue"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
	"os/exec"
	"project-ci/internal/publisher"
	"project-ci/internal/schema"
	"sync"
	"time"
)

type TaskQueue struct {
	projects *gqueue.Queue // 需要执行部署的项目列表
	pub      *publisher.Broker
	idle     bool
	mLock    sync.Mutex
}

func NewTaskQueue(pub *publisher.Broker) *TaskQueue {
	return &TaskQueue{
		projects: gqueue.New(),
		pub:      pub,
		idle:     true,
	}
}

func (a *TaskQueue) Insert(project *schema.Project) {
	a.projects.Push(project)
}

func (a *TaskQueue) setIdle(bool2 bool) {
	a.mLock.Lock()
	defer a.mLock.Unlock()
	a.idle = bool2
}

func (a *TaskQueue) Do() error {
	go func() {
		for {
			if p := a.projects.Pop(); p != nil && a.idle {
				a.setIdle(false)
				if !a.idle {
					a.job(p.(*schema.Project))
					a.setIdle(true)
				}
			}
		}
	}()
	return nil
}

func (a *TaskQueue) job(project *schema.Project) error {
	var result bytes.Buffer
	sourcePath := fmt.Sprintf("%s/%s", project.Path, g.Cfg().GetString("server.default_source_name"))
	if !gfile.Exists(sourcePath) {
		err := RunScript(&result, []*exec.Cmd{
			exec.Command("git", "clone", "-b", project.Ref, project.Repo, sourcePath),
		})
		if err != nil {
			glog.Errorf("clone项目失败%s，名称:[%s], repo:[%s], 路径:[%s]",
				err.Error(), project.Name, project.Repo, project.Path)
			return err
		}
	}
	// 切换源码目录
	_ = gfile.Chdir(sourcePath)
	// 直接拉取项目
	err := RunScript(&result, []*exec.Cmd{
		exec.Command("git", "pull", "origin", project.Ref),
		exec.Command("git", "checkout", project.Ref),
	})
	if err != nil {
		glog.Errorf("拉取分支失败%s，名称:[%s], repo:[%s]，分支:[%s], 路径:[%s]",
			err.Error(), project.Name, project.Repo, project.Ref, project.Path)
		return err
	}
	// 切换工作目录
	_ = gfile.Chdir(project.Path)
	// 执行前置脚本
	if project.BeforeShell != "" {
		err = RunScript(&result, []*exec.Cmd{
			exec.Command("sh", project.BeforeShell),
		})
		if err != nil {
			glog.Errorf("执行before_shell出错%s", err.Error())
		}
	}
	// 执行后置脚本
	if project.AfterShell != "" {
		err = RunScript(&result, []*exec.Cmd{
			exec.Command("/bin/bash", "-c", project.AfterShell),
		})
		if err != nil {
			glog.Errorf("执行after_shell出错%s", err.Error())
		}
	}
	glog.Infof("项目构建完成！名称:[%s], 提交人:[%s]",
		project.Name, project.Pusher)
	glog.Infof("执行结果：%s", result.String())
	// 通知
	if project.Notify {
		msg := map[string]interface{}{
			"content": fmt.Sprintf("项目名称:%s\r\n提交人:%s\r\n编译结果:\r\n%s", project.Name, project.Pusher, result.String()),
		}
		data := map[string]interface{}{
			"msgtype": "text",
			"text":    msg,
		}
		jsondata, _ := json.Marshal(data)
		_, err := g.Client().ContentJson().Post(project.WebHook, jsondata)
		if err != nil {
			glog.Error(err)
		}
	}
	go func(r string) {
		timer := time.NewTicker(time.Second * 3)
		msg := &publisher.GitMessage{
			CommitUser: project.Pusher,
			Name:       project.Name,
			Result:     r,
		}
		select {
		case publisher.StdoutChan <- msg:
			return
		case <-timer.C:
			return
		}
	}(result.String())
	return nil
}

func RunScript(result *bytes.Buffer, cmdList []*exec.Cmd) error {
	for _, cmd := range cmdList {
		glog.Debugf("当前执行命令:%s", cmd.String())
		out, err := cmd.CombinedOutput()
		result.Write(out)
		if err != nil {
			result.WriteString(err.Error())
			return err
		}
	}
	return nil
}
