package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	v1 "github.com/pipeline-operator/api/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parseTemp(templatePathName string) *v1.Pipeline {
	// 读取 YAML 文件内容
	yamlFile, err := os.ReadFile(templatePathName)
	if err != nil {
		panic(err)
	}

	// 创建 Pipeline 结构体实例
	var pipeline v1.Pipeline

	// 使用 YAML 解析库解析 YAML 内容到结构体
	err = yaml.Unmarshal(yamlFile, &pipeline)
	if err != nil {
		panic(err)
	}

	// 打印解析结果
	//fmt.Printf("%+v\n", pipeline)
	return &pipeline

}
func TestNewDeployment(t *testing.T) {
	// Get the runtime path of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get the runtime path")
	}

	// Construct the path to the YAML file
	yamlPath := filepath.Join(filepath.Dir(filename), "../../../config/samples/distri-infer_v1_pipeline.yaml")

	pipeline := parseTemp(yamlPath)
	ds := *NewDeployments(pipeline)

	fmt.Printf("label: %+v\n", ds[0].Labels)
	fmt.Printf("name: %+v\n", ds[0].Name)
	fmt.Printf("namespace: %+v\n", ds[0].Namespace)
	fmt.Printf("Containers: %+v\n", ds[0].Spec.Template.Spec.Containers)
	fmt.Printf("volumes: %+v\n", ds[0].Spec.Template.Spec.Volumes)
	fmt.Printf("affinity: %+v\n", ds[0].Spec.Template.Spec.Affinity)
	fmt.Printf("env: %+v\n", ds[0].Spec.Template.Spec.Containers[0].Env)
	fmt.Printf("location: %+v\n", ds[0].Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions)

	// 打印模板渲染结果
	//fmt.Printf("%+v\n", ds)

}

func TestNewService(t *testing.T) {
	pipeline := parseTemp("../../../config/samples/distri-infer_v1_pipeline.yaml")
	ser := *NewServices(pipeline)

	fmt.Printf("name: %+v\n", ser[0].Name)
	fmt.Printf("namespace: %+v\n", ser[0].Namespace)
	fmt.Printf("selector: %+v\n", ser[0].Spec.Selector)
	fmt.Printf("ports: %+v\n", ser[0].Spec.Ports)

	// 打印模板渲染结果
	//fmt.Printf("%+v\n", ser)

}

func TestNewStorage(t *testing.T) {
	pipeline := parseTemp("../../../config/samples/distri-infer_v1_pipeline.yaml")
	pv, pvc := NewStorageVolume(pipeline)

	fmt.Printf("name: %+v\n", pv.Name)
	fmt.Printf("namespace: %+v\n", pv.Namespace)
	fmt.Printf("CSI: %+v\n", pv.Spec.CSI)
	fmt.Printf("CSI.VolumeAttributes: %+v\n", pv.Spec.CSI.VolumeAttributes)

	fmt.Printf("name: %+v\n", pvc.Name)
	fmt.Printf("namespace: %+v\n", pvc.Namespace)
	fmt.Printf("volumeName: %+v\n", pvc.Spec.VolumeName)

	// 打印模板渲染结果
	//fmt.Printf("%+v\n", ser)

}
