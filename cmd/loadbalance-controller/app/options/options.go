package options

import (
	"github.com/spf13/pflag"
	"os"
)

type LoadBalanceFlags struct {
	KubeConfig         string
	LoadBalanceConfig  string
	Region             string
	LeaseLockId        string
	LeaseLockName      string
	LeaseLockNamespace string
}

type LoadBalanceServer struct {
	LoadBalanceFlags
}

func (l *LoadBalanceFlags) AddFlags(mainfs *pflag.FlagSet) {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	defer func() {
		fs.VisitAll(func(f *pflag.Flag) {
			if len(f.Deprecated) > 0 {
				f.Hidden = false
			}
		})
		mainfs.AddFlagSet(fs)
	}()

	fs.StringVar(&l.KubeConfig, "kubeconfig", l.KubeConfig, "the configuration file for communicating with kubernetes. If it is empty, it will access the token in the cluster by default. The passed value is generally $HOME/.kube/config")
	fs.StringVar(&l.LoadBalanceConfig, "loadbalanceconfig", l.LoadBalanceConfig, "configuration file for loadbalance controller")
	fs.StringVar(&l.LeaseLockId, "lease-lock", l.LeaseLockId, "the lease lock id. should unique")
	fs.StringVar(&l.LeaseLockName, "lease-lock-name", l.LeaseLockName, "the lease lock resource name")
	fs.StringVar(&l.LeaseLockNamespace, "lease-lock-namespace", l.LeaseLockNamespace, "the lease lock resource namespace")
}

func (l *LoadBalanceFlags) SetDefaultRequiredValue() {
	if l.LeaseLockName == "" {
		l.LeaseLockName = "loadbalance-controller"
	}

	if l.LeaseLockNamespace == "" {
		l.LeaseLockNamespace = "kube-system"
	}

	if l.LeaseLockId == "" {
		id, _ := os.Hostname()
		l.LeaseLockId = id
	}
}

func NewLoadBalanceServer() *LoadBalanceServer {
	return &LoadBalanceServer{
		LoadBalanceFlags{
			LoadBalanceConfig:  "config.yml",
			LeaseLockName:      "loadbalance-controller",
			LeaseLockNamespace: "kube-system",
		},
	}

}
