package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/browser"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/pluginutils"
	"k8s.io/kubernetes/pkg/kubectl/proxy"
	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	proxyAddress = "127.0.0.1"
)

func init() {
	// Initialize glog flags
	flag.CommandLine.Set("logtostderr", "true")
	flag.CommandLine.Set("v", os.Getenv("KUBECTL_PLUGINS_GLOBAL_FLAG_V"))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kubectl plugin open-svc SERVICE_NAME")
		os.Exit(1)
	}

	svcName := os.Args[1]
	port, err := strconv.Atoi(os.Getenv("KUBECTL_PLUGINS_LOCAL_FLAG_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	if err := openService(svcName, port); err != nil {
		log.Fatal(err)
	}
}

func openService(svcName string, port int) error {
	restConfig, kubeConfig, err := pluginutils.InitClientAndConfig()
	if err != nil {
		log.Fatalf("Failed to init client and config: %v", err)
	}

	client := kubernetes.NewForConfigOrDie(restConfig)
	namespace, _, _ := kubeConfig.Namespace()

	svc, err := client.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get service/%s: %v\n", svcName, err)
	}

	var urls []string
	var paths []string

	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ingress := svc.Status.LoadBalancer.Ingress[0]
		ip := ingress.IP
		if ip == "" {
			ip = ingress.Hostname
		}
		for _, port := range svc.Spec.Ports {
			urls = append(urls, "http://"+ip+":"+strconv.Itoa(int(port.Port)))
		}
	} else {
		name := svc.ObjectMeta.Name

		if len(svc.Spec.Ports) > 0 {
			port := svc.Spec.Ports[0]

			// guess if the scheme is https
			scheme := ""
			if port.Name == "https" || port.Port == 443 {
				scheme = "https"
			}

			// format is <scheme>:<service-name>:<service-port-name>
			name = utilnet.JoinSchemeNamePort(scheme, svc.ObjectMeta.Name, port.Name)

			paths = append(paths, "/api/v1/namespaces/"+namespace+"/services/"+name+"/proxy")
		}
	}

	if len(urls) == 0 && len(paths) == 0 {
		return fmt.Errorf("Looks like service/%s is a headless service\n", svcName)
	}

	server, err := newProxyServer(restConfig)
	if err != nil {
		return err
	}

	l, err := server.Listen(proxyAddress, port)
	if err != nil {
		return err
	}

	addr := l.Addr().String()

	for _, path := range paths {
		urls = append(urls, fmt.Sprintf("http://%s%s", addr, path))
	}

	fmt.Printf("Starting to serve on %s\n", addr)
	go func() {
		glog.Fatal(server.ServeOnListener(l))
	}()

	fmt.Printf("Opening service/%s in the default browser...\n", svcName)
	for _, url := range urls {
		if err := browser.OpenURL(url); err != nil {
			return fmt.Errorf("Failed to open %s in the default browser\n", url)
		}
	}

	// receive signals and exit
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	return nil
}

func newProxyServer(cfg *rest.Config) (*proxy.Server, error) {
	filter := &proxy.FilterServer{
		AcceptPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathAcceptRE),
		RejectPaths:   proxy.MakeRegexpArrayOrDie(proxy.DefaultPathRejectRE),
		AcceptHosts:   proxy.MakeRegexpArrayOrDie(proxy.DefaultHostAcceptRE),
		RejectMethods: proxy.MakeRegexpArrayOrDie(proxy.DefaultMethodRejectRE),
	}
	return proxy.NewServer("", "/", "", filter, cfg)
}
