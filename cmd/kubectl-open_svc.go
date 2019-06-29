package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"github.com/superbrothers/kubectl-open-svc-plugin/pkg/cmd"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

func init() {
	// Initialize glog flags
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Set("logtostderr", "true")
}

func main() {
	flags := pflag.NewFlagSet("kubectl-open-svc", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdOpenService(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
