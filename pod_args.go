package main

import (
	"strconv"
	"strings"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/meta"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

func processPod(cfg *rest.Config, pod core.Pod) (*APIServerConfig, error) {
	running, err := core_util.PodRunningAndReady(pod)
	if err != nil {
		return nil, err
	}
	if !running {
		return nil, errors.Errorf("pod %s is not running", pod.Name)
	}

	if len(pod.Spec.Containers) != 1 {
		return nil, errors.Errorf("pod %s has %d containers, expected 1 container", len(pod.Spec.Containers))
	}
	container := pod.Spec.Containers[0]
	args := map[string]string{}
	if len(container.Command) > 1 {
		if container.Command[0] != "kube-apiserver" {
			return nil, errors.Errorf(`pod %s is using command %s, expected "kube-apiserver"`, pod.Name, container.Command[0])
		}
		args = meta.ParseArgumentListToMap(container.Command)
	} else if len(container.Args) > 0 {
		args = meta.ParseArgumentListToMap(container.Args)
	}

	var config APIServerConfig

	config.PodName = pod.Name
	config.NodeName = pod.Spec.NodeName
	config.PodIP = pod.Status.PodIP
	config.HostIP = pod.Status.HostIP

	if v, ok := args["admission-control"]; ok && v != "" {
		config.AdmissionControl = strings.Split(v, ",")
	}
	if v, ok := args["enable-admission-plugins"]; ok && v != "" {
		config.AdmissionControl = strings.Split(v, ",")
	}

	if v, ok := args["client-ca-file"]; ok && v != "" {
		data, err := core_util.ExecIntoPod(cfg, &pod, "cat", v)
		if err != nil {
			return nil, err
		}
		config.ClientCAData = strings.TrimSpace(data)
	}

	if v, ok := args["requestheader-client-ca-file"]; ok && v != "" {
		data, err := core_util.ExecIntoPod(cfg, &pod, "cat", v)
		if err != nil {
			return nil, err
		}
		config.RequestheaderClientCAData = strings.TrimSpace(data)
	}

	config.AllowPrivileged, err = strconv.ParseBool(args["allow-privileged"])
	if err != nil {
		return nil, err
	}

	if v, ok := args["authorization-mode"]; ok && v != "" {
		config.AuthorizationMode = strings.Split(v, ",")
	}
	return &config, nil
}
