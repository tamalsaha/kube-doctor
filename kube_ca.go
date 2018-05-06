package main

import (
	"io/ioutil"

	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

func extractKubeCA(cfg *rest.Config, info *ClusterInfo) error {
	info.ClientConfig.Host = cfg.Host
	info.ClientConfig.Insecure = cfg.Insecure

	if len(cfg.CAData) > 0 {
		info.ClientConfig.CABundle = strings.TrimSpace(string(cfg.CAData))
	} else if len(cfg.CAFile) > 0 {
		data, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load ca file %s", cfg.CAFile)
		}
		info.ClientConfig.CABundle = strings.TrimSpace(string(data))
	}
	return nil
}
