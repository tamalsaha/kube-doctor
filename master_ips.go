package main

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func extractMasterIPs(kc kubernetes.Interface, info *ClusterInfo) error {
	if err := tryGKE(info); err == nil {
		return nil
	}

	nodes, err := kc.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master",
	})
	if err != nil {
		return err
	}
	ips := make([]net.IP, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		ip := nodeIP(node)
		if ip != nil {
			ips = append(ips, ip)
		}
	}
	info.MasterIPs = ips
	return nil
}

func nodeIP(node core.Node) []byte {
	for _, addr := range node.Status.Addresses {
		if addr.Type == core.NodeExternalIP {
			return ipBytes(net.ParseIP(addr.Address))
		}
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == core.NodeInternalIP {
			return ipBytes(net.ParseIP(addr.Address))
		}
	}
	return nil
}

func ipBytes(ip net.IP) []byte {
	if ip == nil {
		return nil
	}
	v4 := ip.To4()
	if v4 != nil {
		return v4
	}
	v6 := ip.To16()
	if v6 != nil {
		return v6
	}
	return nil
}

// Product file path that contains the cloud service name.
// This is a variable instead of a const to enable testing.
var gceProductNameFile = "/sys/class/dmi/id/product_name"

// ref: https://cloud.google.com/compute/docs/storing-retrieving-metadata
func tryGKE(info *ClusterInfo) error {
	// ref: https://github.com/kubernetes/kubernetes/blob/a0f94123616c275f94e7a5b680d60d6f34e92f37/pkg/credentialprovider/gcp/metadata.go#L115
	data, err := ioutil.ReadFile(gceProductNameFile)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(string(data))
	if name != "Google" && name != "Google Compute Engine" {
		return errors.New("not GKE")
	}

	client := &http.Client{Timeout: time.Millisecond * 100}
	req, err := http.NewRequest(http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/instance/attributes/kube-env", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	content := make(map[string]interface{})
	err = yaml.Unmarshal(body, &content)
	if err != nil {
		return err
	}
	v, ok := content["KUBERNETES_MASTER_NAME"]
	if !ok {
		return errors.New("missing  KUBERNETES_MASTER_NAME")
	}
	masterIP := v.(string)

	info.MasterIPs = []string{masterIP}
	return nil
}
