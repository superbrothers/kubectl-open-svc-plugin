package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/pkg/browser"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/pluginutils"
	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	namespace = "kube-system"
	proxyPort = "9001"
)

func init() {
	// Initialize glog flags
	flag.CommandLine.Set("logtostderr", "true")
	flag.CommandLine.Set("v", os.Getenv("KUBECTL_PLUGINS_GLOBAL_FLAG_V"))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kubectl plugin service SERVICE_NAME")
		os.Exit(1)
	}

	svcName := os.Args[1]
	restConfig, kubeConfig, err := pluginutils.InitClientAndConfig()
	if err != nil {
		log.Fatalf("Failed to init client and config: %v", err)
	}

	client := kubernetes.NewForConfigOrDie(restConfig)
	namespace, _, _ := kubeConfig.Namespace()

	svc, err := client.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get service/%s in %s namespace: %v\n", svcName, namespace, err)
	}

	var urls []string

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

			urls = append(urls, "http://127.0.0.1:"+proxyPort+"/api/v1/namespaces/"+namespace+"/services/"+name+"/proxy")
		}
	}

	if len(urls) == 0 {
		log.Fatalf("Looks like service/%s in %s namespace is a headless service\n", svcName, namespace)
	}

	// TODO: implements a proxy server instead of using kubectl proxy
	cmd := exec.Command("kubectl", "proxy", "--port", proxyPort)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond) // Wait for running kubectl proxy...
	fmt.Printf("Opening service/%s in %s namespace in the default browser...\n", svcName, namespace)

	for _, url := range urls {
		if err := browser.OpenURL(url); err != nil {
			log.Fatalf("Failed to open %s in the default browser\n", url)
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
