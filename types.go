package main

import (
	"github.com/ghodss/yaml"
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
