package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/templates"

	"k8s.io/kubectl/pkg/proxy"
	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	defaultPort      = 8001
	defaultAddress   = "127.0.0.1"
	defaultKeepalive = 0 * time.Second
	defaultScheme    = ""
	defaultURL       = false

	schemeTypes = map[string]interface{}{
		"":      nil,
		"http":  nil,
		"https": nil,
	}

	openServiceLong = templates.LongDesc(`
		Open the Kubernetes URL(s) for the specified service in your browser
		through a local proxy server.
	`)
	openServiceExample = templates.Examples(`
		# Open service/kubernetes-dashboard in namespace/kube-system
		kubectl open-svc kubernetes-dashboard -n kube-system

		# Use "https" scheme with --scheme option for connections between the apiserver
		# and service/rook-ceph-mgr-dashboard in namespace/rook-ceph
		kubectl open-svc rook-ceph-mgr-dashboard -n rook-ceph --scheme https
	`)
)

// OpenServiceOptions provides information required to open the service in the
// browser
type OpenServiceOptions struct {
	configFlags *genericclioptions.ConfigFlags

	args      []string
	port      int
	address   string
	keepalive time.Duration
	scheme    string
	url       bool

	genericclioptions.IOStreams
}

// NewOpenServiceOptions provides an instance of OpenServiceOptions with
// default values
func NewOpenServiceOptions(streams genericclioptions.IOStreams) *OpenServiceOptions {
	return &OpenServiceOptions{
		configFlags: genericclioptions.NewConfigFlags(true),

		port:      defaultPort,
		address:   defaultAddress,
		keepalive: defaultKeepalive,
		scheme:    defaultScheme,
		url:       defaultURL,

		IOStreams: streams,
	}
}

// NewCmdOpenService provides a cobra command wrapping OpenServiceOptions
func NewCmdOpenService(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOpenServiceOptions(streams)

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("kubectl open-svc SERVICE [--port=%d] [--address=%s] [--keepalive=%d] [--url=%t]", defaultPort, defaultAddress, defaultKeepalive, defaultURL),
		Short:   "Open the Kubernetes URL(s) for the specified service in your browser.",
		Long:    openServiceLong,
		Example: openServiceExample,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			c.SilenceUsage = true
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&o.port, "port", "p", o.port, "The port on which to run the proxy. Set to 0 to pick a random port.")
	cmd.Flags().StringVar(&o.address, "address", o.address, "The IP address on which to serve on.")
	cmd.Flags().DurationVar(&o.keepalive, "keepalive", o.keepalive, "keepalive specifies the keep-alive period for an active network connection. Set to 0 to disable keepalive.")
	cmd.Flags().StringVar(&o.scheme, "scheme", o.scheme, `The scheme for connections between the apiserver and the service. It must be "http" or "https" if specfied.`)
	cmd.Flags().BoolVarP(&o.url, "url", "u", o.url, "Print the URL instead of opening it.")
	o.configFlags.AddFlags(cmd.Flags())

	// add the klog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	_ = flag.CommandLine.Parse([]string{})

	return cmd
}

// Complete sets all information required for opening the service
func (o *OpenServiceOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *OpenServiceOptions) Validate() error {
	if len(o.args) != 1 {
		return fmt.Errorf("exactly one SERVICE is required, got %d", len(o.args))
	}

	if _, ok := schemeTypes[o.scheme]; !ok {
		return fmt.Errorf(`scheme must be "http" or "https" if specified`)
	}

	return nil
}

// Run opens the service in the browser
func (o *OpenServiceOptions) Run() error {
	serviceName := o.args[0]

	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	kubeConfig := o.configFlags.ToRawKubeConfigLoader()

	client := kubernetes.NewForConfigOrDie(restConfig)
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return err
	}

	service, err := client.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get service/%s in namespace/%s: %v\n", serviceName, namespace, err)
	}

	var urls []string
	var paths []string

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		ip := ingress.IP
		if ip == "" {
			ip = ingress.Hostname
		}
		for _, port := range service.Spec.Ports {
			urls = append(urls, "http://"+ip+":"+strconv.Itoa(int(port.Port)))
		}
	} else if len(service.Spec.Ports) > 0 {
		port := service.Spec.Ports[0]

		// guess if the scheme is https
		scheme := ""
		if port.Name == "https" || port.Port == 443 {
			scheme = "https"
		}
		// override the scheme if scheme option is specified
		if o.scheme != "" {
			scheme = o.scheme
		}

		// format is <scheme>:<service-name>:<service-port-name>
		name := utilnet.JoinSchemeNamePort(scheme, service.Name, port.Name)

		paths = append(paths, "/api/v1/namespaces/"+namespace+"/services/"+name+"/proxy")
	}

	if len(urls) == 0 && len(paths) == 0 {
		return fmt.Errorf("Looks like service/%s is a headless service\n", serviceName)
	}

	if o.url {
		for _, path := range paths {
			urls = append(urls, restConfig.Host+path)
		}

		for _, url := range urls {
			fmt.Fprintln(o.Out, url)
		}

		return nil
	}

	server, err := proxy.NewServer("", "/", "", nil, restConfig, o.keepalive)
	if err != nil {
		return err
	}

	l, err := server.Listen(o.address, o.port)
	if err != nil {
		return err
	}

	addr := l.Addr().String()

	for _, path := range paths {
		urls = append(urls, "http://"+addr+path)
	}

	fmt.Fprintf(o.Out, "Starting to serve on %s\n", addr)
	go func() {
		klog.Fatal(server.ServeOnListener(l))
	}()

	fmt.Fprintf(o.Out, "Opening service/%s in the default browser...\n", serviceName)
	for _, url := range urls {
		if err := browser.OpenURL(url); err != nil {
			return fmt.Errorf("Failed to open %s in the default browser\n", url)
		}
	}

	// receive signals and exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	return nil
}
