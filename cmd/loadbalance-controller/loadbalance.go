package main

import (
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/loadbalance-controller/app"
	"k8s.io/component-base/cli"
	"os"
)

func main() {
	command := app.NewLoadBalanceCommand()
	code := cli.Run(command)
	os.Exit(code)
}
