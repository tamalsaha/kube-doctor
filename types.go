package main

import (
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/version"
)

type ClusterInfo struct {
	Version       version.Info        `json:"version"`
	Master        MasterInfo          `json:"master"`
	RequestHeader RequestHeaderConfig `json:"requestHeaderConfig"`
}

type MasterInfo struct {
	IPs      []string `json:"ips"`
	CABundle string   `json:"caBundle"`
	Insecure bool     `json:"insecure"`
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
