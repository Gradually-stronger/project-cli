package router

import (
	"github.com/gogf/gf/frame/g"
	"project-ci/internal/publisher"
	"project-ci/internal/router/ctl"
	"project-ci/internal/schema"
	"project-ci/internal/task"
)

func RegisterRouter(
	projects map[string]*schema.Project,
	taskQ *task.TaskQueue,
	broker *publisher.Broker,
	) {
	s := g.Server()
	gr :=s.Group("/api")
	v1 := gr.Group("/v1")
	cHook := ctl.NewHook(projects, broker, taskQ)
	{
		v1.POST("hook/:name", cHook.Hook)
	}
	grWeb := s.Group("/web")
	{
		grWeb.GET("projects", cHook.GetList)
		grWeb.GET("projects/:name", cHook.Monitor)
		grWeb.GET("ws/:name", cHook.WebSocket)
	}

}