package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/loadbalance-controller/app/options"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/cloudprovider/controllers"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	componentName = "loadbalance-controller"
)

func NewLoadBalanceCommand() *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(componentName, pflag.ContinueOnError)
	serverOption := options.NewLoadBalanceServer()
	cmd := &cobra.Command{
		Use:                componentName,
		Long:               `a loadbalance-controller based on kubernetes can configure the service type as loadbalance, so as to automatically assign ip based on the controller`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// initial flag parse, since we disable cobra's flag parsing
			if err := cleanFlagSet.Parse(args); err != nil {
				return fmt.Errorf("failed to parse %s flag: %w", componentName, err)
			}

			// check if there are non-flag arguments in the command line
			cmds := cleanFlagSet.Args()
			if len(cmds) > 0 {
				return fmt.Errorf("unknown command %+s", cmds[0])
			}

			// short-circuit on help
			help, err := cleanFlagSet.GetBool("help")
			if err != nil {
				return errors.New(`"help" flag is non-bool, programmer error, please correct`)
			}
			if help {
				return cmd.Help()
			}

			if serverOption.LoadBalanceConfig == "" {
				return fmt.Errorf(`%s config flag is required`, componentName)
			}

			// set options default values
			serverOption.SetDefaultRequiredValue()

			cliflag.PrintFlags(cleanFlagSet)

			return run(serverOption)
		},
	}
	serverOption.AddFlags(cleanFlagSet)
	cleanFlagSet.BoolP("help", "h", false, fmt.Sprintf("help for %s", cmd.Name()))

	// ugly, but necessary, because Cobra's default UsageFunc and HelpFunc pollute the flagset with global flags
	const usageFmt = "Usage:\n  %s\n\nFlags:\n%s"
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine(), cleanFlagSet.FlagUsagesWrapped(2))
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine(), cleanFlagSet.FlagUsagesWrapped(2))
	})

	return cmd
}

func buildKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func run(server *options.LoadBalanceServer) error {

	kubeClientConfig, err := buildKubeConfig(server.KubeConfig)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	loadbalanceConfig, err := config.NewLoadBalanceConfig(server.LoadBalanceConfig)
	if err != nil {
		return err
	}

	kubeClientConfig.ContentType = "application/json"

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loadbalanceController, err := controllers.NewLoaBalanceController(kubeClient, loadbalanceConfig)
	if err != nil {
		return err
	}

	// main service blocking signal
	stop := make(chan struct{})

	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		close(stop)
		cancel()
	}()

	runFunc := func(ctx context.Context) {
		// complete your controller loop here
		klog.Info("Controller loop...")
		go loadbalanceController.Run(1, stop)
		select {}
	}

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      server.LeaseLockName,
			Namespace: server.LeaseLockNamespace,
		},
		Client: kubeClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: server.LeaseLockId,
		},
	}

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				runFunc(ctx)
			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				klog.Infof("leader lost: %s", server.LeaseLockId)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == server.LeaseLockId {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
	return nil
}
