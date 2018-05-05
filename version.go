package main

import "k8s.io/client-go/kubernetes"

func extractVersion(kc kubernetes.Interface, info *ClusterInfo) error {
	v, err := kc.Discovery().ServerVersion()
	if err != nil {
		return err
	}
	info.Version = v
	return err
}
