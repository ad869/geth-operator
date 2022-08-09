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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/ad869/geth-operator/api/shared"
	ethereumv1alpha1 "github.com/ad869/geth-operator/api/v1alpha1"
	"github.com/ad869/geth-operator/internal/ethereum"
	"github.com/go-logr/logr"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type reconcileClusterRequestContext struct {
	context.Context

	log     logr.Logger
	cluster *ethereumv1alpha1.Cluster
}

//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=clusters/finalizers,verbs=update

//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ethereum.applying.cool,resources=nodes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=watch;get;create;update;list;delete

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ctx := reconcileClusterRequestContext{
		Context: ctx,
		log:     log.FromContext(ctx).WithValues("node", req.NamespacedName),
		cluster: new(ethereumv1alpha1.Cluster),
	}
	if err := r.Get(ctx, req.NamespacedName, _ctx.cluster); err != nil {
		return RequeueIfError(client.IgnoreNotFound(err))
	}

	shared.UpdateLabels(_ctx.cluster, "goquorum")

	return r.reconcileCluster(_ctx)
}

func (r *ClusterReconciler) reconcileCluster(ctx reconcileClusterRequestContext) (ctrl.Result, error) {

	if !ctx.cluster.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.removeFinalizerAndUpdate(ctx)
	}

	if !ContainsString(ctx.cluster.ObjectMeta.GetFinalizers(), string(FinalizerNode)) {
		return r.addFinalizerAndRequeue(ctx)
	}

	var (
		nodes       = []ethereum.GroupNode{}
		clusterSpec = ctx.cluster.Spec
		cluster     = ctx.cluster.DeepCopy()
	)

	for i := 0; i < clusterSpec.Validator.Number; i++ {
		nodes = append(nodes, ethereum.NewGroupNode(cluster, ethereum.NodeRoleValidator, i))
	}

	for i := 0; i < clusterSpec.Member.Number; i++ {
		nodes = append(nodes, ethereum.NewGroupNode(cluster, ethereum.NodeRoleMember, i))
	}

	if err := r.reconcileSecret(ctx, nodes); err != nil {
		return RequeueIfError(err)
	}

	if err := r.reconcileNode(ctx, nodes); err != nil {
		return RequeueIfError(err)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) reconcileNode(ctx reconcileClusterRequestContext, nodes []ethereum.GroupNode) (err error) {

	staticNodes, err := r.getStaticNodes(ctx, nodes)
	if err != nil {
		return
	}

	validators, err := r.getValidatorsAddress(ctx, nodes)
	if err != nil {
		return
	}

	for _, v := range validators {
		ctx.cluster.Spec.Genesis.Accounts = append(ctx.cluster.Spec.Genesis.Accounts, ethereumv1alpha1.Account{
			Address: v,
			Balance: "1000000000000000000000000000",
		})
	}

	for _, node := range nodes {

		ethNode := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      node.String(),
				Namespace: ctx.cluster.Namespace,
			},
		}

		clusterSpec := ctx.cluster.DeepCopy().Spec

		ctrl.CreateOrUpdate(ctx, r.Client, ethNode, func() (err error) {
			if err := ctrl.SetControllerReference(ctx.cluster, ethNode, r.Scheme); err != nil {
				return err
			}
			ethNode.ObjectMeta.Labels = ctx.cluster.GetLabels()

			ethNode.Spec.Image = clusterSpec.Image

			// setting genesis
			ethNode.Spec.Genesis = clusterSpec.Genesis
			ethNode.Spec.Genesis.QBFT.Validators = validators

			if node.IsValidator() {
				ethNode.Spec.Resources = clusterSpec.Validator.Resources
				ethNode.Spec.Miner = true
			} else {
				ethNode.Spec.Resources = clusterSpec.Member.Resources
				ethNode.Spec.Miner = false
			}

			ethNode.Spec.StaticNodes = staticNodes

			if ethNode.Spec.Coinbase, err = node.Address(); err != nil {
				return
			}

			ethNode.Spec.NodePrivateKeySecretName = node.String()

			// fmt.Printf("%+v\n", ethNode)

			return nil
		})
	}

	return nil
}

func (r *ClusterReconciler) getValidatorsAddress(ctx reconcileClusterRequestContext, nodes []ethereum.GroupNode) (addresses []ethereumv1alpha1.EthereumAddress, err error) {

	for _, node := range nodes {

		if node.IsValidator() {
			var address ethereumv1alpha1.EthereumAddress

			if address, err = node.Address(); err != nil {
				return
			}
			addresses = append(addresses, address)
		}
	}

	return
}

func (r *ClusterReconciler) getStaticNodes(ctx reconcileClusterRequestContext, nodes []ethereum.GroupNode) (enodeURLs []ethereumv1alpha1.Enode, err error) {

	for _, node := range nodes {
		var url ethereumv1alpha1.Enode
		url, err = node.EnodeURL()
		if err != nil {
			return
		}

		enodeURLs = append(enodeURLs, url)
	}

	return
}

func (r *ClusterReconciler) reconcileSecret(ctx reconcileClusterRequestContext, nodes []ethereum.GroupNode) (err error) {

	for _, node := range nodes {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      node.String(),
				Namespace: ctx.cluster.Namespace,
			},
		}

		_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() (err error) {
			if err := ctrl.SetControllerReference(ctx.cluster, secret, r.Scheme); err != nil {
				return err
			}
			secret.ObjectMeta.Labels = ctx.cluster.GetLabels()

			privKey, err := node.PrivateKey()
			if err != nil {
				return
			}

			secret.Data = map[string][]byte{
				"key":      []byte(privKey),
				"password": []byte(""),
			}

			return nil
		})
		if err != nil {
			return err
		}

	}

	return nil
}

// Add finalizer to prevent delete unexpected.
func (r *ClusterReconciler) addFinalizerAndRequeue(ctx reconcileClusterRequestContext) (ctrl.Result, error) {
	ctx.cluster.ObjectMeta.Finalizers = append(ctx.cluster.ObjectMeta.Finalizers, string(FinalizerNode))
	ctx.log.Info("Add finalizer and Requeue")
	if err := r.Update(ctx, ctx.cluster); err != nil {
		ctx.log.Error(err, "Failed to add finalizer")
		return RequeueIfError(err)
	}
	return RequeueImmediately()
}

// Remove the finalizer and update resource.
func (r *ClusterReconciler) removeFinalizerAndUpdate(ctx reconcileClusterRequestContext) (ctrl.Result, error) {
	log := ctx.log.WithName("removeFinalizerAndUpdate")
	ctx.cluster.ObjectMeta.Finalizers = RemoveString(ctx.cluster.ObjectMeta.Finalizers, string(FinalizerNode))
	err := r.Update(ctx, ctx.cluster)
	if err != nil {
		log.Info("Failed to remove finalizer", "error", err)
	}
	log.Info("Finalizer has been removed from the job")
	return RequeueIfError(err)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Cluster{}).
		Complete(r)
}
