# kubectl plugin open-svc SERVICE_NAME

This is a kubectl plugin that open the Kubernetes URL(s) for the specified service in your browser.

![Screenshot](./screenshots/kubectl-open-svc-plugin.gif)

```
$ kubectl plugin open-svc -h
Open the Kubernetes URL(s) for the specified service in your browser through a local proxy server using kubectl proxy.

Examples:
  # Open service/kubernetes-dashboard in kube-system namespace.
  kubectl plugin open-svc kubernetes-dashboard -n kube-system

Options:
  -p, --port='8001': The port on which to run the proxy. Set to 0 to pick a random port.

Usage:
  kubectl plugin open-svc [flags] [options]

Use "kubectl options" for a list of global command-line options (applies to all commands).
```

## Install the plugin

You can install this plugin with [krew](https://github.com/GoogleContainerTools/krew) that is package manager for kubectl plugins.
```
$ kubectl plugin install open-svc
```

If you are on macOS, you can install with homebrew:
```
$ brew tap superbrothers/kubectl-open-svc-plugin
$ brew install kubectl-open-svc-plugin
```

If you are on Linux, you can install with the following steps:
```
$ curl -sL -o open-svc.zip https://github.com/superbrothers/kubectl-open-svc-plugin/releases/download/$(curl -sL https://raw.githubusercontent.com/superbrothers/kubectl-open-svc-plugin/master/version.txt)/open-svc-$(uname | tr '[:upper:]' '[:lower:]')-amd64.zip
$ mkdir -p ~/.kube/plugins/open-svc
$ unzip open-svc.zip -d ~/.kube/plugins/open-svc
```

## License

This software is released under the MIT License.
