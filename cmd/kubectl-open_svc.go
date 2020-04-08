package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	"github.com/superbrothers/kubectl-open-svc-plugin/pkg/cmd"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

func init() {
	// Initialize glog flags
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Set("logtostderr", "true")
}

func main() {
	flags := pflag.NewFlagSet("kubectl-open-svc", pflag.ExitOnError)
	pflag.CommandLine = flags

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer func() {
		signal.Stop(quit)
		cancel()
	}()
	go func() {
		select {
		case <-quit: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-quit // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()

	root := cmd.NewCmdOpenService(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(exitCodeErr)
	}
}
