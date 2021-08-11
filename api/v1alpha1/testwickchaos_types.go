// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base
// +chaos-mesh:oneshot=true

// TestWickChaos is the Schema for the TestWickChaos API
type TestWickChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec     TestWickChaosSpec   `json:"spec,omitempty"`
	Status   TestWickChaosStatus `json:"status,omitempty"`
	TestWick TestWick            `json:"testWick,omitempty"`
}

// TestWickChaosSpec is the content of the specification for a TestWickChaos
type TestWickChaosSpec struct {
	ContainerSelector `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`
}

// TestWickChaosStatus represents the status of a TestWickChaos
type TestWickChaosStatus struct {
	ChaosStatus `json:",inline"`
}

type TestWick struct {
	ProvisionerURL  *string `json:"provisionerURL,omitempty"`
	HostedZone      *string `json:"hostedZone,omitempty"`
	Samples         *int    `json:"samples,omitempty"`
	ChannelSamples  *int    `json:"channelSamples,omitempty"`
	ChannelMessages *int    `json:"channelMessages,omitempty"`
	Owner           *string `json:"owner,omitempty"`
	Size            *string `json:"size,omitempty"`
	AffinityType    *string `json:"affinityType,omitempty"`
	DBType          *string `json:"dbType,omitempty"`
	FileStore       *string `json:"fileStore,omitempty"`
}

// GetSelectorSpecs is a getter for selectors
func (obj *TestWickChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.ContainerSelector,
	}
}
