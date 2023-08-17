package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/khulnasoft-lab/vul-kubernetes/pkg/artifacts"
	"github.com/khulnasoft-lab/vul-kubernetes/pkg/k8s"
	"github.com/khulnasoft-lab/vul-kubernetes/pkg/vulk8s"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"context"
)

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx := context.Background()

	cluster, err := k8s.GetCluster()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current namespace:", cluster.GetCurrentNamespace())

	vulk8s := vulk8s.New(cluster, logger.Sugar())

	fmt.Println("Scanning cluster")

	//vul k8s #cluster
	artifacts, err := vulk8s.ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning namespace 'default'")
	//vul k8s --namespace default
	artifacts, err = vulk8s.Namespace("default").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)
	fmt.Println("Scanning all namespaces ")
	artifacts, err = vulk8s.AllNamespaces().ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning namespace 'default', resource 'deployment/orion'")

	//vul k8s --namespace default deployment/orion
	artifact, err := vulk8s.Namespace("default").GetArtifact(ctx, "deploy", "orion")
	if err != nil {
		log.Fatal(err)
	}
	printArtifact(artifact)

	fmt.Println("Scanning 'deployments'")

	//vul k8s deployment
	artifacts, err = vulk8s.Namespace("default").Resources("deployment").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning 'cm,pods'")
	//vul k8s clusterroles,pods
	artifacts, err = vulk8s.Namespace("default").Resources("cm,pods").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	tolerations := []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: pointer.Int64(300),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: pointer.Int64(300),
		},
	}

	// collect node info
	ar, err := vulk8s.ListArtifactAndNodeInfo(ctx, "vul-temp", map[string]string{"chen": "test"}, tolerations...)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range ar {
		if a.Kind != "NodeInfo" {
			continue
		}
		fmt.Println(a.RawResource)
	}

	bi, err := vulk8s.ListBomInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bb, err := json.Marshal(bi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(bb))
}

func printArtifacts(artifacts []*artifacts.Artifact) {
	for _, artifact := range artifacts {
		printArtifact(artifact)
	}
}

func printArtifact(artifact *artifacts.Artifact) {
	fmt.Printf(
		"Name: %s, Kind: %s, Namespace: %s, Images: %v\n",
		artifact.Name,
		artifact.Kind,
		artifact.Namespace,
		artifact.Images,
	)
}
