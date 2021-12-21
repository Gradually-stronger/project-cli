package internal

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"project-ci/internal/publisher"
	"project-ci/internal/router"
	"project-ci/internal/schema"
	"project-ci/internal/task"
)

var ProjectMap map[string]*schema.Project

func Init() {
	g.Cfg().SetFileName("config.json")
	ProjectMap = parseProjects()
	initServer()
}
func initServer() {
	s := g.Server()
	port := g.Cfg().GetInt("server.port")
	s.SetPort(port)
	pub := publisher.New()
	pub.Start()
	taskQ := task.NewTaskQueue(pub)
	err := taskQ.Do()
	if err != nil {
		panic(err)
	}
	// 注册路由
	router.RegisterRouter(ProjectMap, taskQ, pub)
}

// 解析项目
func parseProjects() map[string]*schema.Project {
	var s schema.Project
	projects := make(map[string]*schema.Project)
	result := g.Cfg().GetJsons("projects", &s)
	for _, r := range result {
		projects[r.GetString("name")] = &schema.Project{
			Name:        r.GetString("name"),
			Ref:         r.GetString("ref"),
			Repo:        r.GetString("repo"),
			Path:        r.GetString("path"),
			BeforeShell: r.GetString("before_shell"),
			AfterShell:  r.GetString("after_shell"),
			Notify:      r.GetBool("notify"),
			WebHook:     r.GetString("web_hook"),
		}
	}
	glog.Infof("初始化%d个项目", len(projects))
	return projects
}
