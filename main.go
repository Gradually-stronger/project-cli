package main

import (
	"github.com/gogf/gf/frame/g"
	"project-ci/internal"
)

func main() {
	internal.Init()
	s := g.Server()
	s.EnablePProf()
	s.Run()
}
