package main

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
)

type ClusterInfo struct {
	Version      version.Info                     `json:"version"`
	ClientConfig RestConfig                       `json:"clientConfig"`
	APIServers   []APIServerConfig                `json:"apiServers"`
	AuthConfig   ExtensionApiserverAuthentication `json:"extension-apiserver-authentication"`
}

type RestConfig struct {
	Host     string
	CABundle string `json:"caBundle"`
	Insecure bool   `json:"insecure"`
}

type APIServerConfig struct {
	PodName                   string
	NodeName                  string
	PodIP                     string
	HostIP                    string
	AdmissionControl          []string
	ClientCAData              string
	RequestheaderClientCAData string
	AllowPrivileged           bool
	AuthorizationMode         []string
}

type ExtensionApiserverAuthentication struct {
	ClientCA      string
	RequestHeader *RequestHeaderConfig `json:"requestHeaderConfig"`
}

type RequestHeaderConfig struct {
	// UsernameHeaders are the headers to check (in order, case-insensitively) for an identity. The first header with a value wins.
	UsernameHeaders []string
	// GroupHeaders are the headers to check (case-insensitively) for a group names.  All values will be used.
	GroupHeaders []string
	// ExtraHeaderPrefixes are the head prefixes to check (case-insentively) for filling in
	// the user.Info.Extra.  All values of all matching headers will be added.
	ExtraHeaderPrefixes []string
	// ClientCA points to CA bundle file which is used verify the identity of the front proxy
	ClientCA string
	// AllowedClientNames is a list of common names that may be presented by the authenticating front proxy.  Empty means: accept any.
	AllowedClientNames []string
}

func (c ClusterInfo) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (c ClusterInfo) Validate() error {
	var errs []error

	{
		if c.ClientConfig.Insecure {
			errs = append(errs, errors.New("Admission webhooks can't be used when kube apiserver is accesible without verifying its TLS certificate (insecure-skip-tls-verify : true)."))
		} else {
			if c.AuthConfig.ClientCA == "" {
				errs = append(errs, errors.Errorf(`"%s/%s" configmap is missing "client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
			} else if c.ClientConfig.CABundle != c.AuthConfig.ClientCA {
				errs = append(errs, errors.Errorf(`"%s/%s" configmap has mismatched "client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
			}

			for _, pod := range c.APIServers {
				if pod.ClientCAData != c.ClientConfig.CABundle {
					errs = append(errs, errors.Errorf(`pod "%s"" has mismatched "client-ca-file".`, pod.PodName))
				}
			}
		}
	}
	{
		if c.AuthConfig.RequestHeader == nil {
			errs = append(errs, errors.Errorf(`"%s/%s" configmap is missing "requestheader-client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
		}
		for _, pod := range c.APIServers {
			if pod.RequestheaderClientCAData != c.AuthConfig.RequestHeader.ClientCA {
				errs = append(errs, errors.Errorf(`pod "%s"" has mismatched "requestheader-client-ca-file".`, pod.PodName))
			}
		}
	}
	{
		for _, pod := range c.APIServers {
			modes := sets.NewString(pod.AuthorizationMode...)
			if !modes.Has("RBAC") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable RBAC authorization mode.`, pod.PodName))
			}
		}
	}
	{
		for _, pod := range c.APIServers {
			adms := sets.NewString(pod.AdmissionControl...)
			if !adms.Has("MutatingAdmissionWebhook") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable MutatingAdmissionWebhook admission controller.`, pod.PodName))
			}
			if !adms.Has("ValidatingAdmissionWebhook") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable ValidatingAdmissionWebhook admission controller.`, pod.PodName))
			}
		}
	}
	return utilerrors.NewAggregate(errs)
}
