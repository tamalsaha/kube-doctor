package main

import (
	"fmt"
	"path/filepath"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var info ClusterInfo

	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		glog.Fatalf("Could not get Kubernetes config: %s", err)
	}

	err = extractKubeCA(config, &info)
	if err != nil {
		glog.Fatalln(err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	err = extractVersion(kc, &info)
	if err != nil {
		glog.Fatalln(err)
	}

	err = extractRequestheaderInfo(kc, &info)
	if err != nil {
		glog.Fatalln(err)
	}

	err = extractMasterArgs(config, kc, &info)
	if err != nil {
		glog.Fatalln(err)
	}

	fmt.Println("-------------------------------------------------")
	fmt.Println(info)
	fmt.Println("-------------------------------------------------")

	err = info.Validate()
	if err != nil {
		glog.Fatalln(err)
	}
}
