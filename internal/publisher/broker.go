package publisher

import (
	"github.com/gogf/gf/container/glist"
	"github.com/gogf/gf/os/glog"
	"time"
)

var StdoutChan chan *GitMessage

func init() {
	StdoutChan = make(chan *GitMessage, 20)
}

// Sub is a sub client
type Sub struct {
	MessageChan chan *GitMessage // message channel
	Name        string           // sub name
	Close       chan struct{}    // close flag
}

// GitMessage is a git message
type GitMessage struct {
	CommitUser string `json:"commit_user"` // commit user name
	Name       string `json:"name"`        // project name
	Result     string `json:"result"`      // Compile result
}

// Broker is a server broker
type Broker struct {
	// 保存订阅频道，key为订阅的主题，value 用一个链表保存订阅的客户端channel
	pubSubChannels map[string]*glist.List
	// 订阅客户channel写入超时控制
	rTimeOut time.Duration
}

// New create a broker instance
func New() *Broker {
	return &Broker{
		pubSubChannels: make(map[string]*glist.List),
		rTimeOut:       time.Second * 5,
	}
}

func (b *Broker) Start() {
	glog.Info("start stdout broker")
	go b.publish()
}

// Sub subscribe
func (b *Broker) Sub(name string, sub *Sub) {
	if _, ok := b.pubSubChannels[name]; !ok {
		l := glist.New(true)
		b.pubSubChannels[name] = l
		l.PushBack(sub)
		return
	}
	b.pubSubChannels[name].PushBack(sub)
}

// publish message to clients
func (b *Broker) publish() {
	for {
		select {
		case msg := <-StdoutChan:
			b.send(msg)
		}
	}
}
func (b *Broker) send(msg *GitMessage) {
	for name, chans := range b.pubSubChannels {
		if name == msg.Name {
			length := chans.Len()
			for n := 0; n < length; n++ {
				e := chans.Front()
				if sub, ok := e.Value.(*Sub); ok && sub != nil {
					select {
					case <-sub.Close:
						chans.Remove(e)
					case sub.MessageChan <- msg:
					case <-time.After(b.rTimeOut):
						chans.Remove(e)
					}
				}
				e = e.Next()
			}
		}
	}
}
