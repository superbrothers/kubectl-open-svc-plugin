package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestOpenServiceOptionsGetServiceProxyURL(t *testing.T) {
	tests := []struct {
		name  string
		o     *OpenServiceOptions
		svc   *v1.Service
		paths []string
	}{
		{
			"not specified scheme",
			&OpenServiceOptions{},
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
			[]string{
				"/api/v1/namespaces/default/services/nginx/proxy",
			},
		},
		{
			"specified scheme",
			&OpenServiceOptions{
				scheme: "https",
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
			[]string{
				"/api/v1/namespaces/default/services/https:nginx:/proxy",
			},
		},
		{
			"443 port",
			&OpenServiceOptions{},
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
			[]string{
				"/api/v1/namespaces/default/services/https:nginx:/proxy",
			},
		},
		{
			"443 port with specified scheme",
			&OpenServiceOptions{
				scheme: "http",
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
			[]string{
				"/api/v1/namespaces/default/services/http:nginx:/proxy",
			},
		},
		{
			"service port https",
			&OpenServiceOptions{},
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
			[]string{
				"/api/v1/namespaces/default/services/https:nginx:https/proxy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.o.getServiceProxyPaths(tt.svc)
			if !reflect.DeepEqual(tt.paths, paths) {
				assert.Equal(t, tt.paths, paths)
			}
		})
	}
}
