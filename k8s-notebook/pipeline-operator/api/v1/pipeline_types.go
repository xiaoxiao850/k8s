/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelinePhase string

const (
	PipelineAvailable   PipelinePhase = "Available"
	PipelineUnAvailable PipelinePhase = "UnAvailable"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// +kubebuilder:default:listenPort=9080
// +kubebuilder:default:replicas=1
// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Pipeline. Edit pipeline_types.go to remove/update
	//Foo string `json:"foo,omitempty"`

	//all steps of a pipeline
	Steps []Step `json:"steps"`
	//step's pod
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`
	//listenPort: "9080"
	ListenPort int32 `json:"listenPort"`
	//storage of model documents
	ModelStorage ModelStorage `json:"modelStorage"`
}

type Step struct {
	Args     map[string]string `json:"args,omitempty"` //可以为空
	Image    string            `json:"image"`
	Location []string          `json:"location"` //schedule result
	Model    string            `json:"model"`    //model name
}

type ModelStorage struct {
	CSIParameter map[string]string `json:"csiParameter"` // nfs: server+share path
	Type         string            `json:"type"`         //csi type
}

type StepPhase struct {
	DeploymentPhase string `json:"deploymentPhase"`
	// ServicePhase    string `json:"servicePhase"`
}

type PipelineDetailPhase struct {
	StepsPhase []StepPhase `json:"stepsPhase,omitempty"`
	PVCPhase   string      `json:"pvcPhase,omitempty"`
	PVPhase    string      `json:"pvPhase,omitempty"`
}

// +kubebuilder:default:phase=UnAvailable
// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//pipeline phase: Available Unavailable
	Phase       string              `json:"phase,omitempty"`
	DetailPhase PipelineDetailPhase `json:"detailPhase,omitempty"`
	StepsLength int                 `json:"stepsLength,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Pipeline is the Schema for the pipelines API
// +kubebuilder:printcolumn:name="StepsLength",type="integer",JSONPath=".status.stepsLength"
// +kubebuilder:printcolumn:name="Status.phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Status.DetailPhase",type="string",JSONPath=".status.detailPhase"
// +kubebuilder:resource:path=pipelines,scope=Cluster
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
