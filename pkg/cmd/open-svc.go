package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/superbrothers/kubectl-open-svc-plugin/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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

		# Open http-monitoring port name of service/istiod in namespace/istio-system
		kubectl open-svc istiod -n istio-system --svc-port http-monitoring

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
	svcPort   string
	address   string
	keepalive time.Duration
	scheme    string

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

		IOStreams: streams,
	}
}

// NewCmdOpenService provides a cobra command wrapping OpenServiceOptions
func NewCmdOpenService(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOpenServiceOptions(streams)

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("kubectl open-svc SERVICE [--port=%d] [--address=%s] [--keepalive=%d]", defaultPort, defaultAddress, defaultKeepalive),
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
	cmd.Flags().StringVar(&o.svcPort, "svc-port", o.svcPort, "The service port name. default is empty and uses the first port")
	cmd.Flags().StringVar(&o.address, "address", o.address, "The IP address on which to serve on.")
	cmd.Flags().DurationVar(&o.keepalive, "keepalive", o.keepalive, "keepalive specifies the keep-alive period for an active network connection. Set to 0 to disable keepalive.")
	cmd.Flags().StringVar(&o.scheme, "scheme", o.scheme, `The scheme for connections between the apiserver and the service. It must be "http" or "https" if specfied.`)
	o.configFlags.AddFlags(cmd.Flags())

	// add the klog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

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

	proxyPath, err := o.getServiceProxyPath(service)
	if err != nil {
		return err
	}

	server, err := proxy.NewServer("", "/", "", nil, restConfig, o.keepalive, false)
	if err != nil {
		return err
	}

	l, err := server.Listen("127.0.0.1", 0)
	if err != nil {
		return err
	}

	klog.V(4).Infof("Starting to serve kubectl proxy on %s\n", l.Addr().String())

	go func() {
		klog.Fatal(server.ServeOnListener(l))
	}()

	target, err := url.Parse("http://" + l.Addr().String() + proxyPath)
	if err != nil {
		return err
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.ModifyResponse = utils.StripModifierFunc(target.Path)
	srv := &http.Server{
		Addr:    o.getListenAddr(),
		Handler: reverseProxy,
	}

	fmt.Fprintf(o.Out, "Starting to serve on %s\n", o.getListenAddr())

	go func() {
		klog.Fatal(srv.ListenAndServe())
	}()

	fmt.Fprintf(o.Out, "Opening service/%s in the default browser...\n", serviceName)
	if err := browser.OpenURL(o.getListenURL()); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	defer close(quit)

	signal.Notify(quit, os.Interrupt)
	<-quit

	return nil
}

func (o *OpenServiceOptions) getListenAddr() string {
	return fmt.Sprintf("%s:%d", o.address, o.port)
}

func (o *OpenServiceOptions) getListenURL() string {
	return "http://" + o.getListenAddr()
}

func (o *OpenServiceOptions) getServiceProxyPath(svc *v1.Service) (string, error) {
	l := len(svc.Spec.Ports)

	if l == 0 {
		return "", fmt.Errorf("Looks like service/%s is a headless service", svc.GetName())
	}

	var port v1.ServicePort

	if o.svcPort == "" {
		port = svc.Spec.Ports[0]

		if l > 1 {
			fmt.Fprintf(o.ErrOut, "service/%s has %d ports, defaulting port %d\n", svc.GetName(), l, port.Port)
		}
	} else {
		for _, p := range svc.Spec.Ports {
			if p.Name == o.svcPort {
				port = p
				break
			}
		}

		if len(port.Name) == 0 {
			return "", fmt.Errorf("port %s not found in service/%s", o.svcPort, svc.GetName())
		}
	}

	scheme := o.scheme
	if scheme == "" {
		// guess if the scheme is https
		if port.Name == "https" || port.Port == 443 {
			scheme = "https"
		}
	}

	// format is <scheme>:<service-name>:<service-port-name>
	name := utilnet.JoinSchemeNamePort(scheme, svc.GetName(), port.Name)
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s/proxy", svc.GetNamespace(), name), nil
}
