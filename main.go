/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	api_v1 "crd.tanjunchen.io/pkg/apis/crdcontroller/v1"
	"crd.tanjunchen.io/pkg/generated/clientset/versioned"
	"crd.tanjunchen.io/pkg/generated/clientset/versioned/typed/crdcontroller/v1"
	"crd.tanjunchen.io/pkg/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

// Home returns home path.
func Home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// Kubeconfig returns kube config path.
func Kubeconfig() string {
	return filepath.Join(Home(), ".kube", "config")
}

func main() {
	if kubeconfig == "" {
		kubeconfig = Kubeconfig()
	}

	config, e := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if e != nil {
		fmt.Println(e)
	}
	client, e := v1.NewForConfig(config)
	if e != nil {
		fmt.Println(e)
	}
	fooList, e := client.Tanjunchens("test").List(context.TODO(), metav1.ListOptions{})
	fmt.Println(fooList, e)

	clientset, e := versioned.NewForConfig(config)

	informerFactory := externalversions.NewSharedInformerFactory(clientset, 30*time.Second)
	informer := informerFactory.Crd().V1().Tanjunchens()
	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    onAdd,
			UpdateFunc: onUpdate,
			DeleteFunc: onDelete,
		})
	lister := informer.Lister()

	stopCh := make(chan struct{})
	defer close(stopCh)
	informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		return
	}

	tanjunchens, err := lister.Tanjunchens("test").List(labels.Everything())
	if err != nil {
		fmt.Println(err)
	}
	for _, tanjunchen := range tanjunchens {
		fmt.Printf("%s\t%s\t%v%s\t%s\t\r\n", tanjunchen.Name, tanjunchen.Spec.Name, tanjunchen.Spec.Age, tanjunchen.Spec.Location, tanjunchen.Spec.Occupations)
	}
	<-stopCh
}

func onAdd(obj interface{}) {
	tanjunchen := obj.(*api_v1.Tanjunchen)
	fmt.Printf("onAdd : %v", tanjunchen)
}

func onUpdate(old, new interface{}) {
	oldTanjunchen := old.(*api_v1.Tanjunchen)
	newTanjunchen := new.(*api_v1.Tanjunchen)
	fmt.Printf("onUpdate : %v to %v\r\n", oldTanjunchen, newTanjunchen)
}

func onDelete(obj interface{}) {
	tanjunchen := obj.(*api_v1.Tanjunchen)
	fmt.Printf("onDelete : %v\r\n", tanjunchen)
}
