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

package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	_ "embed"

	"github.com/ad869/geth-operator/api/shared"
	ethereumv1alpha1 "github.com/ad869/geth-operator/api/v1alpha1"
	"github.com/ad869/geth-operator/internal/ethereum"
	"github.com/go-logr/logr"
)

var (
	//go:embed geth_init_genesis.sh
	GethInitScript string

	//go:embed geth_import_account.sh
	gethImportAccountScript string
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type reconcileNodeRequestContext struct {
	context.Context

	log  logr.Logger
	node *ethereumv1alpha1.Node
}

//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
//+kubebuilder:rbac:groups=core,resources=secrets;services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	_ctx := reconcileNodeRequestContext{
		Context: ctx,
		log:     log.FromContext(ctx).WithValues("node", req.NamespacedName),
		node:    new(ethereumv1alpha1.Node),
	}
	if err := r.Get(ctx, req.NamespacedName, _ctx.node); err != nil {
		return RequeueIfError(client.IgnoreNotFound(err))
	}
	_ctx.log.Info(fmt.Sprintf("start reconcile node %s", _ctx.node.Name))

	_ctx.node.Default()

	shared.UpdateLabels(_ctx.node, "quorum")

	return r.reconcileNode(_ctx)
}

func (r *NodeReconciler) reconcileNode(ctx reconcileNodeRequestContext) (ctrl.Result, error) {

	if !ctx.node.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.removeFinalizerAndUpdate(ctx)
	}

	if !ContainsString(ctx.node.ObjectMeta.GetFinalizers(), string(FinalizerNode)) {
		return r.addFinalizerAndRequeue(ctx)
	}

	// Because the pvc needs to be managed by the csi-controller, so adding a delay
	if err := r.reconcilePVC(ctx); err != nil {
		return RequeueAfterInterval(5*time.Second, err)
	}

	if err := r.reconcileConfigMap(ctx); err != nil {
		return RequeueIfError(err)
	}

	if _, err := r.reconcileService(ctx); err != nil {
		return RequeueIfError(err)
	}

	if err := r.reconcileStatefulSet(ctx); err != nil {
		return RequeueIfError(err)
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) reconcileStatefulSet(ctx reconcileNodeRequestContext) (err error) {

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctx.node.Name,
			Namespace: ctx.node.Namespace,
		},
	}

	client := ethereum.NewClient(*ctx.node.DeepCopy())

	volumeMounts := r.createNodeVolumeMounts(ctx)

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(ctx.node, sts, r.Scheme); err != nil {
			return err
		}

		template := corev1.PodSpec{
			Volumes: r.createNodeVolumes(ctx),
			InitContainers: []corev1.Container{
				{
					Name:         "init",
					Image:        ctx.node.Spec.Image,
					Env:          client.ENV(),
					Command:      []string{"/bin/sh"},
					Args:         []string{fmt.Sprintf("%s/geth-init.sh", client.PathConfig())},
					VolumeMounts: volumeMounts,
				},
				{
					Name:         "import-account",
					Image:        ctx.node.Spec.Image,
					Env:          client.ENV(),
					Command:      []string{"/bin/sh"},
					Args:         []string{fmt.Sprintf("%s/import-account.sh", client.PathConfig())},
					VolumeMounts: volumeMounts,
				},
			},
			Containers: []corev1.Container{{
				Name:  "node",
				Image: ctx.node.Spec.Image,
				Env:   client.ENV(),
				// Command: []string{"/bin/sh", "-c", "sleep 3600"},
				Args: client.Args(),
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(ctx.node.Spec.Resources.CPU),
						corev1.ResourceMemory: resource.MustParse(ctx.node.Spec.Resources.Memory),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(ctx.node.Spec.Resources.CPULimit),
						corev1.ResourceMemory: resource.MustParse(ctx.node.Spec.Resources.MemoryLimit),
					},
				},
				VolumeMounts: volumeMounts,
			}},
		}

		// spec is imutable, except 'replicas', 'template', 'updateStrategy' and 'minReadySeconds'
		if !sts.CreationTimestamp.IsZero() {
			sts.Spec.Template.Spec = template
			return nil
		}

		// set lables
		labels := ctx.node.GetLabels()
		sts.ObjectMeta.Labels = labels
		sts.Spec.Selector = &metav1.LabelSelector{}
		sts.Spec.Selector.MatchLabels = labels
		sts.Spec.Template.ObjectMeta.Labels = labels

		// set service
		sts.Spec.ServiceName = ctx.node.Name

		//set template
		sts.Spec.Template.Spec = template
		return nil
	})

	return err
}

// createNodeVolumeMounts creates all required volume mounts for the node
func (r *NodeReconciler) createNodeVolumeMounts(ctx reconcileNodeRequestContext) []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{}

	var client ethereum.GoQuorum

	// secrets mount
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "secrets",
		MountPath: client.PathSecrets(),
		ReadOnly:  true,
	})

	// config mount
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "config",
		MountPath: client.PathConfig(),
		ReadOnly:  true,
	})

	// data mount
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "data",
		MountPath: client.PathData(),
	})

	return volumeMounts
}

func (r *NodeReconciler) createNodeVolumes(ctx reconcileNodeRequestContext) []corev1.Volume {
	volumes := []corev1.Volume{}
	projections := []corev1.VolumeProjection{}

	// nodekey (node private key) projection
	nodekeyProjection := corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: ctx.node.Spec.NodePrivateKeySecretName,
			},
			Items: []corev1.KeyToPath{
				{
					Key:  "key",
					Path: "nodekey",
				},
				{
					Key:  "key",
					Path: "account.key",
				},
				{
					Key:  "password",
					Path: "account.password",
				},
			},
		},
	}
	projections = append(projections, nodekeyProjection)

	if len(projections) != 0 {
		secretsVolume := corev1.Volume{
			Name: "secrets",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: projections,
				},
			},
		}
		volumes = append(volumes, secretsVolume)
	}

	configVolume := corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ctx.node.Name,
				},
			},
		},
	}
	volumes = append(volumes, configVolume)

	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: ctx.node.Name,
			},
		},
	}
	volumes = append(volumes, dataVolume)

	return volumes
}

func (r *NodeReconciler) reconcileService(ctx reconcileNodeRequestContext) (ip string, err error) {
	log := ctx.log.WithName("reconcileConfigMap")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctx.node.Name,
			Namespace: ctx.node.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {

		if err := ctrl.SetControllerReference(ctx.node, svc, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference on service")
			return err
		}

		labels := ctx.node.GetLabels()

		svc.ObjectMeta.Labels = labels

		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "discovery",
				Port:       int32(ctx.node.Spec.P2PPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.P2PPort)),
				Protocol:   corev1.ProtocolUDP,
			},
			{
				Name:       "json-rpc",
				Port:       int32(ctx.node.Spec.RPCPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.RPCPort)),
				Protocol:   corev1.ProtocolTCP,
			},
			{
				Name:       "ws",
				Port:       int32(ctx.node.Spec.WSPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.WSPort)),
				Protocol:   corev1.ProtocolTCP,
			},
			{
				Name:       "graphql",
				Port:       int32(ctx.node.Spec.GraphQLPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.GraphQLPort)),
				Protocol:   corev1.ProtocolTCP,
			},
			{
				Name:       "rlpx",
				Port:       int32(ctx.node.Spec.RLPXPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.RLPXPort)),
				Protocol:   corev1.ProtocolTCP,
			},
			{
				Name:       "metrics",
				Port:       int32(ctx.node.Spec.MetricsPort),
				TargetPort: intstr.FromInt(int(ctx.node.Spec.MetricsPort)),
				Protocol:   corev1.ProtocolTCP,
			},
		}
		svc.Spec.Selector = labels

		return nil
	})

	if err != nil {
		return
	}

	ip = svc.Spec.ClusterIP

	return
}

func (r *NodeReconciler) reconcileConfigMap(ctx reconcileNodeRequestContext) error {
	log := ctx.log.WithName("reconcileConfigMap")

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctx.node.Name,
			Namespace: ctx.node.Namespace,
		},
	}
	client := ethereum.NewClient(*ctx.node.DeepCopy())

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(ctx.node, configmap, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference on configmap")
			return err
		}

		if configmap.Data == nil {
			configmap.Data = map[string]string{}
		}
		// generate genesis.json
		content, err := client.Genesis()
		if err != nil {
			log.Error(err, "Generate genesis file was failed.")
			return err
		}
		configmap.Data["genesis.json"] = content

		// geth init script
		configmap.Data["geth-init.sh"] = GethInitScript

		configmap.Data["import-account.sh"] = gethImportAccountScript

		// generate static-node.json
		configmap.Data["static-nodes.json"] = client.EncodeStaticNodes()

		return nil
	})

	return err
}

func (r *NodeReconciler) reconcilePVC(ctx reconcileNodeRequestContext) error {

	ctx.log.WithName("reconcile pvc").Info(fmt.Sprintf("expect %s", ctx.node.Spec.Resources.Storage))

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctx.node.Name,
			Namespace: ctx.node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(ctx.node, pvc, r.Scheme); err != nil {
			return err
		}
		request := corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(ctx.node.Spec.Resources.Storage),
		}
		// spec is immutable after creation except resources.requests
		if !pvc.CreationTimestamp.IsZero() {
			pvc.Spec.Resources.Requests = request
			return nil
		}

		pvc.ObjectMeta.Labels = ctx.node.GetLabels()
		pvc.Spec = corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: request,
			},
			StorageClassName: ctx.node.Spec.Resources.StorageClass,
		}
		return nil
	})

	return err
}

// Add finalizer to prevent delete unexpected.
func (r *NodeReconciler) addFinalizerAndRequeue(ctx reconcileNodeRequestContext) (ctrl.Result, error) {
	log := ctx.log.WithName("addFinalizerAndRequeue")

	ctx.node.ObjectMeta.Finalizers = append(ctx.node.ObjectMeta.Finalizers, string(FinalizerNode))
	if err := r.Update(ctx, ctx.node); err != nil {
		log.Error(err, "Adding finalizer was failed.")
		return RequeueIfError(err)
	}
	log.Info("Add finalizer and Requeue.")
	return RequeueImmediately()
}

// Remove the finalizer and update resource.
func (r *NodeReconciler) removeFinalizerAndUpdate(ctx reconcileNodeRequestContext) (ctrl.Result, error) {
	log := ctx.log.WithName("removeFinalizerAndUpdate")

	ctx.node.ObjectMeta.Finalizers = RemoveString(ctx.node.ObjectMeta.Finalizers, string(FinalizerNode))
	err := r.Update(ctx, ctx.node)
	if err != nil {
		log.Info("Removing finalizer was failed.", "error", err)
		return RequeueIfError(err)
	}
	log.Info("Finalizer has been removed from the rescource.")
	return RequeueIfError(err)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
