# kubectl open-svc SERVICE_NAME

This is a kubectl plugin that open the Kubernetes URL(s) for the specified service in your browser. Unlike the `kubectl port-forward` command, this plugin makes services accessible via their ClusterIP.

![Screenshot](./screenshots/kubectl-open-svc-plugin.gif)

```
$ kubectl open-svc -h
Open the Kubernetes URL(s) for the specified service in your browser through a local proxy server.

Usage:
  kubectl open-svc SERVICE [--port=8001] [--address=127.0.0.1] [--keepalive=0] [flags]

Examples:
  # Open service/kubernetes-dashboard in namespace/kube-system
  kubectl open-svc kubernetes-dashboard -n kube-system

  # Use "https" scheme with --scheme option for connections between the apiserver
  # and service/rook-ceph-mgr-dashboard in namespace/rook-ceph
  kubectl open-svc rook-ceph-mgr-dashboard -n rook-ceph --scheme https

Flags:
      --address string                   The IP address on which to serve on. (default "127.0.0.1")
      --alsologtostderr                  log to standard error as well as files
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string                 Default HTTP cache directory (default "/home/dev/.kube/http-cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
  -h, --help                             help for kubectl
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --keepalive duration               keepalive specifies the keep-alive period for an active network connection. Set to 0 to disable keepalive.
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
  -n, --namespace string                 If present, the namespace scope for this CLI request
  -p, --port int                         The port on which to run the proxy. Set to 0 to pick a random port. (default 8001)
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --scheme string                    The scheme for connections between the apiserver and the service. It must be "http" or "https" if specfied.
  -s, --server string                    The address and port of the Kubernetes API server
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## Install the plugin

1. Install [krew](https://github.com/GoogleContainerTools/krew) that is a plugin manager for kubectl
2. Run:

        kubectl krew install open-svc

3. Try it out

        kubectl open-svc -h

## License

This software is released under the MIT License.
