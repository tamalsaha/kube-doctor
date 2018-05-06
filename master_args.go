package main

import (
	"context"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/pager"
)

func extractMasterArgs(cfg *rest.Config, kc kubernetes.Interface, info *ClusterInfo) error {
	pods, err := findMasterPods(kc)
	if err != nil {
		return err
	}

	var errs []error
	for _, pod := range pods {
		if c, err := processPod(cfg, pod); err != nil {
			errs = append(errs, err)
		} else {
			info.Master = append(info.Master, *c)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func findMasterPods(kc kubernetes.Interface) ([]core.Pod, error) {
	pods, err := findMasterPodsByLabel(kc)
	if err != nil {
		return nil, err
	}
	if len(pods) > 0 {
		return pods, nil
	}

	return findMasterPodsByKubernetesService(kc)
}

func findMasterPodsByLabel(kc kubernetes.Interface) ([]core.Pod, error) {
	pods, err := kc.CoreV1().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			"component": "kube-apiserver",
		}).String(),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func findMasterPodsByKubernetesService(kc kubernetes.Interface) ([]core.Pod, error) {
	ep, err := kc.CoreV1().Endpoints(core.NamespaceDefault).Get("kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	podIPs := sets.NewString()
	for _, subnet := range ep.Subsets {
		for _, addr := range subnet.Addresses {
			podIPs.Insert(addr.IP)
		}
	}

	lister := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		return kc.CoreV1().Pods(metav1.NamespaceSystem).List(opts)
	})
	objects, err := lister.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]core.Pod, 0, podIPs.Len())
	err = meta.EachListItem(objects, func(obj runtime.Object) error {
		pod, ok := obj.(*core.Pod)
		if !ok {
			return errors.Errorf("%v is not a pod", obj)
		}
		if podIPs.Has(pod.Status.PodIP) {
			pods = append(pods, *pod)
		}
		return nil
	})
	return pods, err
}
