package main

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/superbrothers/kubectl-open-svc-plugin/pkg/cmd"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flags := pflag.NewFlagSet("kubectl-open-svc", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdOpenService(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
