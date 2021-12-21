package schema

type Project struct {
	Name        string `json:"name"`         // 项目名称
	Ref         string `json:"ref"`          // 分支
	Repo        string `json:"repo"`         // 仓库
	Path        string `json:"path"`         // 本地路径
	BeforeShell string `json:"before_shell"` // 拉取前的命令
	AfterShell  string `json:"after_shell"`  // 拉取后的命令
	Pusher      string // 提交人
	Notify      bool   `json:"notify"`   // 通知开关
	WebHook     string `json:"web_hook"` // 钉钉机器人的web hook url
}
