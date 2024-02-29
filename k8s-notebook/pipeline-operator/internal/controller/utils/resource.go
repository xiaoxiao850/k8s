package utils

import (
	"bytes"
	"path/filepath"
	"runtime"
	"text/template"

	v1 "github.com/pipeline-operator/api/v1"
	"gopkg.in/yaml.v2"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 解析模板文件并将 pipeline 对象的值填充到模板中
func parseTemplate(templateName string, data interface{}) []byte {
	// Get the runtime path of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get the runtime path")
	}

	// Construct the path to the template file
	templatePath := filepath.Join(filepath.Dir(filepath.Dir(filename)), "template", templateName+".yaml")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, data)
	if err != nil {
		panic(err)
	}
	return b.Bytes()

}

type tempPipline struct {
	ObjectMeta metav1.ObjectMeta
	Spec       v1.PipelineSpec
	Step       v1.Step
	Index      int
	NextStep   v1.Step
	Begin      bool
}

// 解析deployment.yaml模板文件 ,解析为Deployment对象
func NewDeployment(temp *tempPipline) *appv1.Deployment {
	d := &appv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", temp), d)
	if err != nil {
		panic(err)
	}
	return d
}

// range steps -> deployments
func NewDeployments(pipeline *v1.Pipeline) *[]appv1.Deployment {
	ds := []appv1.Deployment{}

	for index, step := range pipeline.Spec.Steps {
		//create struct for template
		temp := tempPipline{
			ObjectMeta: pipeline.ObjectMeta,
			Spec:       pipeline.Spec,
			Step:       step,
			Index:      index + 1, //从1开始
			Begin:      index == 0,
		}
		if index == len(pipeline.Spec.Steps)-1 { //is ending
			temp.NextStep = pipeline.Spec.Steps[0] //next step is beging

		} else {
			temp.NextStep = pipeline.Spec.Steps[index+1]
		}
		d := NewDeployment(&temp)
		// d.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0] = corev1.NodeSelectorRequirement{
		// 	Key:      "kubernetes.io/hostname",
		// 	Operator: "In",
		// 	Values:   temp.Step.Location,
		// }
		ds = append(ds, *d)
	}

	return &ds
}

func NewService(temp *tempPipline) *corev1.Service {
	s := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", temp), s)
	if err != nil {
		panic(err)
	}
	return s
}

func NewServices(pipeline *v1.Pipeline) *[]corev1.Service {
	sers := []corev1.Service{}
	for index, step := range pipeline.Spec.Steps {
		//create struct for template
		temp := tempPipline{
			ObjectMeta: pipeline.ObjectMeta,
			Spec:       pipeline.Spec,
			Step:       step,
			Index:      index + 1, //从1开始
			Begin:      index == 0,
		}
		if index == len(pipeline.Spec.Steps)-1 { //is ending
			temp.NextStep = pipeline.Spec.Steps[0] //next step is beging

		} else {
			temp.NextStep = pipeline.Spec.Steps[index+1]
		}
		ser := NewService(&temp)
		sers = append(sers, *ser)
	}

	return &sers
}

type tempStrorageNFS struct {
	ObjectMeta metav1.ObjectMeta
	Spec       v1.PipelineSpec
	NFSServer  string
	NFSShare   string
}

func NewStorageVolume(pipeline *v1.Pipeline) (*corev1.PersistentVolume, *corev1.PersistentVolumeClaim) {
	pv := &corev1.PersistentVolume{}
	pvc := &corev1.PersistentVolumeClaim{}
	if pipeline.Spec.ModelStorage.Type == "nfs" {
		temp := tempStrorageNFS{
			ObjectMeta: pipeline.ObjectMeta,
			Spec:       pipeline.Spec,
			NFSServer:  pipeline.Spec.ModelStorage.CSIParameter["server"],
			NFSShare:   pipeline.Spec.ModelStorage.CSIParameter["share"],
		}
		pvbyts := parseTemplate("static-nfs-pv", temp)
		err := yaml.Unmarshal(pvbyts, pv)
		if err != nil {
			panic(err)
		}
		pvcbyts := parseTemplate("static-nfs-pvc", temp)
		err = yaml.Unmarshal(pvcbyts, pvc)
		if err != nil {
			panic(err)
		}
		return pv, pvc

	} else {
		return nil, nil
	}

}
