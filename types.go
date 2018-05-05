package main

import (
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/version"
)

type ClusterInfo struct {
	Version version.Info `json:"version"`
	Master  MasterInfo    `json:"master"`
}

type MasterInfo struct {
	IPs      []string `json:"ips"`
	CABundle []byte   `json:"caBundle"`
	Insecure bool     `json:"insecure"`
}

func (c ClusterInfo) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(data)
}
