package main

import (
	"flag"
	"fmt"
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/models"
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/routers"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/config"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	configPath string
)

func init() {

	if flag.CommandLine.Lookup("log_dir") != nil {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}

	klog.InitFlags(nil)

	defer klog.Flush()

	flag.StringVar(&configPath, "config", "cloud-provider-manager.yml", "config path")
	flag.Parse()
}

func main() {

	cfg, err := config.NewCloudProviderHTTPConfig(configPath)

	if err != nil {
		klog.Errorf("parser config fail: %s", err.Error())
		os.Exit(1)
	}

	if err = models.RegisterDatabase(cfg.DB); err != nil {
		klog.Errorf("connect db fail: %s", err.Error())
		os.Exit(1)
	}

	s := &http.Server{
		Addr: fmt.Sprintf("%s", fmt.Sprintf("%s:%d", cfg.HTTP.Host,
			cfg.HTTP.Port)),
		Handler:        routers.NewRouter(),
		MaxHeaderBytes: 1 << 20,
	}

	if err = s.ListenAndServe(); err != nil {
		klog.Errorf(err.Error())
		os.Exit(3)
	}

	sics := make(chan os.Signal, 1)
	signal.Notify(sics, syscall.SIGINT, syscall.SIGTERM)
	<-sics

}
