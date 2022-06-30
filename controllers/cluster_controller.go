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
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/ad869/geth-operator/api/shared"
	ethereumv1alpha1 "github.com/ad869/geth-operator/api/v1alpha1"
	"github.com/ad869/geth-operator/internal/crypto"
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

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ctx := reconcileClusterRequestContext{
		Context: ctx,
		log:     log.FromContext(ctx).WithValues("node", req.NamespacedName),
		cluster: new(ethereumv1alpha1.Cluster),
	}
	if err := r.Get(ctx, req.NamespacedName, _ctx.cluster); err != nil {
		return RequeueIfError(client.IgnoreNotFound(err))
	}

	shared.UpdateLabels(_ctx.cluster, "quorum")

	return r.reconcileCluster(_ctx)
}

func (r *ClusterReconciler) reconcileCluster(ctx reconcileClusterRequestContext) (ctrl.Result, error) {

	if !ctx.cluster.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.removeFinalizerAndUpdate(ctx)
	}

	if !ContainsString(ctx.cluster.ObjectMeta.GetFinalizers(), string(FinalizerNode)) {
		return r.addFinalizerAndRequeue(ctx)
	}

	if err := r.reconcileSecret(ctx); err != nil {
		return RequeueIfError(err)
	}

	if err := r.reconcileNode(ctx); err != nil {
		return RequeueIfError(err)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) reconcileNode(ctx reconcileClusterRequestContext) (err error) {

	staticNodes, err := r.staticNodes(ctx)
	if err != nil {
		return
	}

	validators, err := r.getValidatorsAddress(ctx)
	if err != nil {
		return
	}

	for i := 0; i < ctx.cluster.Spec.Validator.Number; i++ {

		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, i),
				Namespace: ctx.cluster.Namespace,
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					Istanbul: ethereumv1alpha1.Istanbul{
						Validators: []ethereumv1alpha1.EthereumAddress{},
					},
				},
			},
		}

		address, err := r.getNodeAddress(ctx, ethereumv1alpha1.NodeTypeValidator, i)
		if err != nil {
			return err
		}

		_, err = ctrl.CreateOrUpdate(ctx, r.Client, node, func() (err error) {
			if err := ctrl.SetControllerReference(ctx.cluster, node, r.Scheme); err != nil {
				return err
			}
			node.ObjectMeta.Labels = ctx.cluster.GetLabels()

			node.Spec = ctx.cluster.Spec.Validator.Spec

			node.Default()

			// fmt.Printf("%+v\n", node)
			node.Spec.Genesis = ctx.cluster.DeepCopy().Spec.Genesis

			node.Spec.StaticNodes = staticNodes

			node.Spec.Genesis.Istanbul.Validators = validators

			node.Spec.Coinbase = ethereumv1alpha1.EthereumAddress(address)

			node.Spec.Import = &ethereumv1alpha1.ImportedAccount{}
			node.Spec.Import.PasswordSecretName = fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, i)
			node.Spec.Import.PrivateKeySecretName = fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, i)

			return nil
		})
	}

	for i := 0; i < ctx.cluster.Spec.Member.Number; i++ {

		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, i),
				Namespace: ctx.cluster.Namespace,
			},
		}

		address, err := r.getNodeAddress(ctx, ethereumv1alpha1.NodeTypeValidator, i)
		if err != nil {
			return err
		}

		_, err = ctrl.CreateOrUpdate(ctx, r.Client, node, func() (err error) {
			if err := ctrl.SetControllerReference(ctx.cluster, node, r.Scheme); err != nil {
				return err
			}
			node.ObjectMeta.Labels = ctx.cluster.GetLabels()

			node.Spec = ctx.cluster.Spec.Validator.Spec

			node.Default()

			node.Spec.StaticNodes = staticNodes

			node.Spec.Genesis.Istanbul.Validators = validators

			node.Spec.Coinbase = ethereumv1alpha1.EthereumAddress(address)

			return nil
		})

	}

	return nil
}

func (r *ClusterReconciler) getValidatorsAddress(ctx reconcileClusterRequestContext) (addresses []ethereumv1alpha1.EthereumAddress, err error) {

	bip44 := crypto.NewBIP44(ctx.cluster.Spec.Mnemonic)

	for i := 0; i < ctx.cluster.Spec.Validator.Number; i++ {

		address, err := bip44.DeriveAddress(r.getNodeAccountIndex(ctx, ethereumv1alpha1.NodeTypeValidator, i))
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, ethereumv1alpha1.EthereumAddress(address))
	}

	return
}

func (r *ClusterReconciler) staticNodes(ctx reconcileClusterRequestContext) (enodeURLs []ethereumv1alpha1.Enode, err error) {

	bip44 := crypto.NewBIP44(ctx.cluster.Spec.Mnemonic)

	for i := 0; i < ctx.cluster.Spec.Validator.Number; i++ {

		var publicKey string
		if publicKey, err = bip44.DerivePublicKey(r.getNodeAccountIndex(ctx, ethereumv1alpha1.NodeTypeValidator, i)); err != nil {
			return
		}

		enodeURLs = append(enodeURLs,
			ethereumv1alpha1.Enode(fmt.Sprintf("enode://%s@%s.%s.svc.cluster.local:%d?discport=0",
				publicKey,
				r.getNodeName(ctx, ethereumv1alpha1.NodeTypeValidator, i),
				ctx.cluster.Namespace,
				30303,
			)))
	}

	for i := 0; i < ctx.cluster.Spec.Member.Number; i++ {

		var publicKey string
		if publicKey, err = bip44.DerivePublicKey(r.getNodeAccountIndex(ctx, ethereumv1alpha1.NodeTypeMember, i)); err != nil {
			return
		}

		enodeURLs = append(enodeURLs,
			ethereumv1alpha1.Enode(fmt.Sprintf("enode://%s@%s.%s.svc.cluster.local:%d",
				publicKey,
				r.getNodeName(ctx, ethereumv1alpha1.NodeTypeMember, i),
				ctx.cluster.Namespace,
				30303,
			)))
	}

	return
}

func (r *ClusterReconciler) reconcileSecret(ctx reconcileClusterRequestContext) (err error) {

	// reconcile secrets of validators
	for i := 0; i < ctx.cluster.Spec.Validator.Number; i++ {

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, i),
				Namespace: ctx.cluster.Namespace,
			},
		}

		_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() (err error) {
			if err := ctrl.SetControllerReference(ctx.cluster, secret, r.Scheme); err != nil {
				return err
			}
			secret.ObjectMeta.Labels = ctx.cluster.GetLabels()

			var (
				accountIndex int
				privKey      string
			)
			bip44 := crypto.NewBIP44(ctx.cluster.Spec.Mnemonic)
			if accountIndex, err = strconv.Atoi(fmt.Sprintf(AccountIndexPrefixValidator, i)); err != nil {
				return
			}
			if privKey, err = bip44.DerivePrivateKey(accountIndex); err != nil {
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

func (r *ClusterReconciler) getNodeAddress(ctx reconcileClusterRequestContext, nodeType ethereumv1alpha1.NodeType, index int) (address string, err error) {
	var accountIndex int
	bip44 := crypto.NewBIP44(ctx.cluster.Spec.Mnemonic)
	if accountIndex, err = strconv.Atoi(fmt.Sprintf(AccountIndexPrefixValidator, index)); err != nil {
		return
	}
	if address, err = bip44.DeriveAddress(accountIndex); err != nil {
		return
	}
	return
}

func (r *ClusterReconciler) getNodeName(ctx reconcileClusterRequestContext, nodeType ethereumv1alpha1.NodeType, index int) string {
	if nodeType == ethereumv1alpha1.NodeTypeValidator {
		return fmt.Sprintf(NameFormatValidator, ctx.cluster.Name, index)
	}
	return fmt.Sprintf(NameFormatMember, ctx.cluster.Name, index)
}

func (r *ClusterReconciler) getNodeAccountIndex(ctx reconcileClusterRequestContext, nodeType ethereumv1alpha1.NodeType, index int) int {
	var indexFmt string

	if nodeType == ethereumv1alpha1.NodeTypeValidator {

		indexFmt = AccountIndexPrefixValidator
	} else {
		indexFmt = AccountIndexPrefixMember
	}

	accountIndex, _ := strconv.Atoi(fmt.Sprintf(indexFmt, index))

	return accountIndex
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
