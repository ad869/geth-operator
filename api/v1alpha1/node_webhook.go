/*
Copyright 2022.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var nodelog = logf.Log.WithName("node-resource")

func (r *Node) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ethereum-applying-cool-v1alpha1-node,mutating=true,failurePolicy=fail,sideEffects=None,groups=ethereum.applying.cool,resources=nodes,verbs=create;update,versions=v1alpha1,name=mnode.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Node) Default() {
	nodelog.Info("default", "name", r.Name)

	if r.Spec.Genesis == nil {
		r.Spec.Genesis = &Genesis{}
	}

	r.Spec.Genesis.Default()

	r.DefaultNodeResources()

	r.DefaultNodePorts()

	if r.Spec.Verbosity == 0 {
		r.Spec.Verbosity = 5
	}
}

func (r *Node) DefaultNodePorts() {

	if r.Spec.P2PPort == 0 {
		r.Spec.P2PPort = DefaultP2PPort
	}
	if r.Spec.RPCPort == 0 {
		r.Spec.RPCPort = DefaultRPCPort
	}
	if r.Spec.WSPort == 0 {
		r.Spec.WSPort = DefaultWSPort
	}
	if r.Spec.GraphQLPort == 0 {
		r.Spec.GraphQLPort = DefaultGraphQLPort
	}
	if r.Spec.RLPXPort == 0 {
		r.Spec.RLPXPort = DefaultRLPXPort
	}
	if r.Spec.MetricsPort == 0 {
		r.Spec.MetricsPort = DefaultMetricsPort
	}
}

func (r *Node) DefaultNodeResources() {

	if r.Spec.Resources.CPU == "" {
		r.Spec.Resources.CPU = DefaultNodeCPURequest
	}
	if r.Spec.Resources.CPULimit == "" {
		r.Spec.Resources.CPULimit = DefaultNodeCPULimit
	}
	if r.Spec.Resources.Memory == "" {
		r.Spec.Resources.Memory = DefaultkNodeMemoryRequest
	}
	if r.Spec.Resources.MemoryLimit == "" {
		r.Spec.Resources.MemoryLimit = DefaultNodeMemoryLimit
	}
	if r.Spec.Resources.Storage == "" {
		r.Spec.Resources.Storage = DefaultNodeStorageRequest
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ethereum-applying-cool-v1alpha1-node,mutating=false,failurePolicy=fail,sideEffects=None,groups=ethereum.applying.cool,resources=nodes,verbs=create;update,versions=v1alpha1,name=vnode.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() error {
	nodelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateUpdate(old runtime.Object) error {
	nodelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
