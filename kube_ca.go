package main

import (
	"io/ioutil"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

func extractKubeCA(cfg *rest.Config, info *ClusterInfo) error {
	info.Master.Insecure = cfg.Insecure

	if len(cfg.CAData) > 0 {
		info.Master.CABundle = cfg.CAData
	} else if len(cfg.CAFile) > 0 {
		data, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load ca file %s", cfg.CAFile)
		}
		info.Master.CABundle = data
	}
	return nil
}
