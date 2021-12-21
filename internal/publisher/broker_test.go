package publisher

import (
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	broker := New()
	sub := newSub("mall")
	sub2 := newSub("release")
	broker.Sub("mall", sub)
	broker.Sub("release", sub2)
	broker.Start()

	// 模拟客户端断开
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	close(sub.Close)
	//}()

	clientClose := make(chan struct{})
	go func() {
		select {
		case <-time.After(time.Second * 30):
			close(clientClose)
		}
	}()
	// client read
	go func() {
		for {
			select {
			case msg := <-sub.MessageChan:
				t.Log(msg.Result)
			case <-clientClose:
				return
			}
		}
	}()
	// client read
	go func() {
		for {
			select {
			case msg := <-sub2.MessageChan:
				t.Log(msg.Result)
			}
		}
	}()

	// server write
	go func() {
		for {
			StdoutChan <- &GitMessage{
				CommitUser: "lijian",
				Name:       "mall",
				Result:     "sdflkasdfasdfsdf",
			}
			time.Sleep(time.Second * 3)
		}
	}()
	select {}
}

func newSub(name string) *Sub {
	sub := new(Sub)
	sub.Name = "name"
	sub.MessageChan = make(chan *GitMessage)
	sub.Close = make(chan struct{})
	return sub
}
