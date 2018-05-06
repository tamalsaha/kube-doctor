package main

import (
	core "k8s.io/api/core/v1"
	"github.com/appscode/kutil/meta"
	"strconv"
	core_util "github.com/appscode/kutil/core/v1"
	"strings"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

func processPod(config *rest.Config, pod core.Pod) error {
	running, err := core_util.PodRunningAndReady(pod)
	if err != nil {
		return err
	}
	if !running {
		return errors.Errorf("pod %s is not running", pod.Name)
	}

	var cfg APIServerConfig

	cfg.IP = pod.Status.PodIP
	cfg.Name = pod.Name
	cfg.NodeName = pod.Spec.NodeName

	args := meta.ParseArgumentListToMap(pod.Spec.Containers[0].Args)

	if v, ok := args["admission-control"]; ok && v != "" {
		cfg.AdmissionControl = strings.Split(v, ",")
	}
	if v, ok := args["enable-admission-plugins"]; ok && v != "" {
		cfg.AdmissionControl = strings.Split(v, ",")
	}

	if v, ok := args["client-ca-file"]; ok && v != "" {
		data, err := core_util.ExecIntoPod(config, &pod, "cat", v)
		if err != nil {
			return err
		}
		cfg.ClientCAData = data
	}

	if v, ok := args["requestheader-client-ca-file"]; ok && v != "" {
		data, err := core_util.ExecIntoPod(config, &pod, "cat", v)
		if err != nil {
			return err
		}
		cfg.RequestheaderClientCAData = data
	}

	cfg.AllowPrivileged, err = strconv.ParseBool(args["allow-privileged"])
	if err != nil {
		return err
	}

	if v, ok := args["authorization-mode"]; ok && v != "" {
		cfg.AuthorizationMode = strings.Split(v, ",")
	}
	return nil
}
