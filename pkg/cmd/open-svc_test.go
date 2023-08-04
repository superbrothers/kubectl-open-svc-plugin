package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestOpenServiceOptionsValidate(t *testing.T) {
	tests := []struct {
		name string
		o    *OpenServiceOptions
		err  error
	}{
		{
			"single arg",
			&OpenServiceOptions{
				args: []string{"nginx"},
			},
			nil,
		},
		{
			"args >= 2",
			&OpenServiceOptions{
				args: []string{"svc", "nginx"},
			},
			errors.New("exactly one SERVICE is required, got 2"),
		},
		{
			"valid scheme",
			&OpenServiceOptions{
				args:   []string{"nginx"},
				scheme: "https",
			},
			nil,
		},
		{
			"invalid scheme",
			&OpenServiceOptions{
				args:   []string{"nginx"},
				scheme: "tcp",
			},
			errors.New(`scheme must be "http" or "https" if specified`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.o.Validate()
			if err == nil {
				if tt.err != err {
					assert.Equal(t, tt.err, err)
				}
			} else if tt.err.Error() != err.Error() {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestOpenServiceOptionsGetServiceProxyPath(t *testing.T) {
	tests := []struct {
		name      string
		o         *OpenServiceOptions
		svc       *v1.Service
		proxyPath string
		err       string
	}{
		{
			"not specified scheme",
			&OpenServiceOptions{IOStreams: genericclioptions.NewTestIOStreamsDiscard()},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/nginx/proxy",
			"",
		},
		{
			"specified scheme",
			&OpenServiceOptions{
				scheme:    "https",
				IOStreams: genericclioptions.NewTestIOStreamsDiscard(),
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/https:nginx:/proxy",
			"",
		},
		{
			"443 port",
			&OpenServiceOptions{IOStreams: genericclioptions.NewTestIOStreamsDiscard()},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port: 443,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/https:nginx:/proxy",
			"",
		},
		{
			"443 port with specified scheme",
			&OpenServiceOptions{
				scheme:    "http",
				IOStreams: genericclioptions.NewTestIOStreamsDiscard(),
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port: 443,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/http:nginx:/proxy",
			"",
		},
		{
			"service port https",
			&OpenServiceOptions{IOStreams: genericclioptions.NewTestIOStreamsDiscard()},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "https",
							Port: 80,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/https:nginx:https/proxy",
			"",
		},
		{
			"multiple ports",
			&OpenServiceOptions{IOStreams: genericclioptions.NewTestIOStreamsDiscard()},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "https",
							Port: 80,
						},
						{
							Port: 8080,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/https:nginx:https/proxy",
			"",
		},
		{
			"no ports",
			&OpenServiceOptions{IOStreams: genericclioptions.NewTestIOStreamsDiscard()},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{},
				},
			},
			"",
			"Looks like service/nginx is a headless service",
		},
		{
			"no ports by portName",
			&OpenServiceOptions{
				IOStreams: genericclioptions.NewTestIOStreamsDiscard(),
				portName:  "noport",
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Port: 8080,
						},
					},
				},
			},
			"",
			"port noport not found for service/nginx",
		},
		{
			"not specified scheme by portName",
			&OpenServiceOptions{
				IOStreams: genericclioptions.NewTestIOStreamsDiscard(),
				portName:  "metrics",
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "https",
							Port: 443,
						},
						{
							Name: "metrics",
							Port: 10254,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/nginx:metrics/proxy",
			"",
		},
		{
			"specified scheme by portName",
			&OpenServiceOptions{
				IOStreams: genericclioptions.NewTestIOStreamsDiscard(),
				portName:  "https",
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "https",
							Port: 443,
						},
						{
							Name: "metrics",
							Port: 10254,
						},
					},
				},
			},
			"/api/v1/namespaces/default/services/nginx:https:https/proxy",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyPath, err := tt.o.getServiceProxyPath(tt.svc)
			if err != nil {
				assert.Equal(t, tt.err, err.Error())
			}
			assert.Equal(t, tt.proxyPath, proxyPath)
		})
	}
}
