package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{Group: "mlaiops.io", Version: "v1alpha1"}

type ReplicaSpec struct {
	Min int32 `json:"min,omitempty"`
	Max int32 `json:"max,omitempty"`
}

type LLMSpec struct {
	Backend             string `json:"backend,omitempty"`
	InferenceServiceRef string `json:"inferenceServiceRef,omitempty"`
}

type ToolReference struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type TrafficPolicy struct {
	CanaryWeight int32  `json:"canaryWeight,omitempty"`
	StableRef    string `json:"stableRef,omitempty"`
}

type NexusAgentSpec struct {
	Version         string          `json:"version"`
	Image           string          `json:"image"`
	GraphModule     string          `json:"graphModule"`
	Replicas        ReplicaSpec     `json:"replicas,omitempty"`
	LLM             LLMSpec         `json:"llm,omitempty"`
	Tools           []ToolReference `json:"tools,omitempty"`
	LangfuseProject string          `json:"langfuseProject,omitempty"`
	TrafficPolicy   TrafficPolicy   `json:"trafficPolicy,omitempty"`
}

type NexusAgentStatus struct {
	Phase              string             `json:"phase,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

type NexusAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NexusAgentSpec   `json:"spec,omitempty"`
	Status            NexusAgentStatus `json:"status,omitempty"`
}

type NexusAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NexusAgent `json:"items"`
}

func AddToScheme(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &NexusAgent{}, &NexusAgentList{})
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

func (in *NexusAgent) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(NexusAgent)
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec.Tools = append([]ToolReference(nil), in.Spec.Tools...)
	out.Status.Conditions = append([]metav1.Condition(nil), in.Status.Conditions...)
	return out
}

func (in *NexusAgentList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(NexusAgentList)
	*out = *in
	out.Items = make([]NexusAgent, len(in.Items))
	for i := range in.Items {
		out.Items[i] = *(in.Items[i].DeepCopyObject().(*NexusAgent))
	}
	return out
}
