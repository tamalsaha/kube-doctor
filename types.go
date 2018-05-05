package main

import (
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/version"
)

type ClusterInfo struct {
	Version *version.Info `json:"version"`
}

func (c *ClusterInfo) String() string {
	if c == nil {
		return ""
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(data)
}
