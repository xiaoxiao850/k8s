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

package controller

import (
	"context"
	nomalerrors "errors"
	"fmt"
	"strconv"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	distriinferv1 "github.com/pipeline-operator/api/v1"
	"github.com/pipeline-operator/internal/controller/utils"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const pipelineNamespace = "pipeline"

func (r *PipelineReconciler) getPVkey(pipeline distriinferv1.Pipeline) (pvkey types.NamespacedName) {
	return client.ObjectKey{
		Name:      fmt.Sprintf("%s-pv", pipeline.Name),
		Namespace: pipelineNamespace, //pv是集群范围资源
	}

}
func (r *PipelineReconciler) getPVCkey(pipeline distriinferv1.Pipeline) (pvckey types.NamespacedName) {
	return client.ObjectKey{
		Name:      fmt.Sprintf("%s-pvc", pipeline.Name),
		Namespace: pipelineNamespace,
	}

}
func (r *PipelineReconciler) getDeploymentskey(pipeline distriinferv1.Pipeline) (deploymentskey []types.NamespacedName) {
	for _, step := range pipeline.Spec.Steps {
		depkey := client.ObjectKey{
			Name:      step.Model,
			Namespace: pipelineNamespace,
		}
		deploymentskey = append(deploymentskey, depkey)
	}
	return deploymentskey

}
func (r *PipelineReconciler) getServiceskey(pipeline distriinferv1.Pipeline) (serviceskey []types.NamespacedName) {
	for _, step := range pipeline.Spec.Steps {
		svckey := client.ObjectKey{
			Name:      step.Model,
			Namespace: pipelineNamespace,
		}
		serviceskey = append(serviceskey, svckey)
	}
	return serviceskey

}

// +kubebuilder:rbac:groups=distri-infer.ndsl.cn,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=distri-infer.ndsl.cn,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=distri-infer.ndsl.cn,resources=pipelines/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=*,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=*,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=*,resources=persistentvolumes/status,verbs=get
// +kubebuilder:rbac:groups=*,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=*,resources=persistentvolumeclaims/status,verbs=get
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	fmt.Printf("处理Reconcile...\n")

	//获取pipeline
	pipeline := &distriinferv1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// fmt.Printf("pipeline.NamespacedName:%+v\n", req.NamespacedName)
	if len(pipeline.Status.DetailPhase.StepsPhase) == 0 {
		stepsPhase := make([]distriinferv1.StepPhase, len(pipeline.Spec.Steps))
		pipeline.Status.DetailPhase.StepsPhase = stepsPhase

	}

	pipeline.Status.StepsLength = len(pipeline.Spec.Steps)

	//处理pv pvc
	//查找同名pv-->exist,not found,
	fmt.Printf("处理pv...\n")
	pv := &corev1.PersistentVolume{}
	pvkey := r.getPVkey(*pipeline)
	// fmt.Printf("pvkey:%+v\n", pvkey)
	pvc := &corev1.PersistentVolumeClaim{}
	pvckey := r.getPVCkey(*pipeline)
	// fmt.Printf("pvckey:%+v\n", pvckey)

	newpv, newpvc := utils.NewStorageVolume(pipeline)
	// fmt.Printf("newpv:%+v\n", newpv)
	// fmt.Printf("newpvc:%+v\n", newpvc)

	if err := controllerutil.SetControllerReference(pipeline, newpv, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference failed!")
		return ctrl.Result{}, err
	}
	// for _, ownerRef := range newpv.OwnerReferences {
	// 	fmt.Printf("newpv OwnerReference: %+v\n", ownerRef)
	// }

	if err := controllerutil.SetControllerReference(pipeline, newpvc, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference failed!")
		return ctrl.Result{}, err
	}
	// fmt.Printf("pv.GetOwnerReference: %+v\n", newpv.GetOwnerReferences())
	// fmt.Printf("pvc.GetOwnerReference: %+v\n", newpvc.GetOwnerReferences())

	if err1 := r.Get(ctx, pvkey, pv); err1 != nil {
		if k8serrors.IsNotFound(err1) { //not found->create
			fmt.Printf("pvIsNotFound\n")

			if err := r.Create(ctx, newpv, &client.CreateOptions{}); err != nil {
				logger.Error(err, "create pv failed!")
				return ctrl.Result{}, err
			}
			fmt.Printf("CreateNewpv success\n")
			fmt.Printf("newpv.DetailPhase.PVPhase: %+v\n", newpv.Status.Phase)
			//成功创建，更新status
			pipeline.Status.DetailPhase.PVPhase = string(newpv.Status.Phase)

			//after Create pv,create pvc
			if err2 := r.Get(ctx, pvckey, pvc); err2 != nil { //pvc
				if k8serrors.IsNotFound(err2) { //not found->create
					fmt.Printf("pvcIsNotFound\n")

					if err := r.Create(ctx, newpvc, &client.CreateOptions{}); err != nil {
						logger.Error(err, "create pvc failed!")
						return ctrl.Result{}, err
					}
					fmt.Printf("CreateNewpvc success\n")
					fmt.Printf("newpvc.DetailPhase.PVCPhase: %+v\n", newpvc.Status.Phase)
					//成功创建，更新status
					pipeline.Status.DetailPhase.PVCPhase = string(newpvc.Status.Phase)

				}
			}

		}
	} else { //exist pv
		// 获取到了 PV 对象
		fmt.Printf("find PV success\n")
		fmt.Printf("PVStatus: %+v\n", pv.Status.Phase)
		//fmt.Printf("pc.DetailPhase.PVCPhase: %+v\n", pvc.Status.Phase)
		//更新status
		pipeline.Status.DetailPhase.PVPhase = string(pv.Status.Phase)
		// pvStatus := &corev1.PersistentVolumeStatus{}
		// 判断 PV 的状态是否为 Released，是的话先删除再重建
		if pv.Status.Phase == corev1.VolumeReleased {
			if err := r.Delete(ctx, newpv, &client.DeleteOptions{}); err != nil {
				logger.Error(err, "Delete pv failed!")
				fmt.Printf("Delete pv failed!\n")
				return ctrl.Result{}, err
			}

			if err := r.Create(ctx, newpv, &client.CreateOptions{}); err != nil {
				logger.Error(err, "create pv failed!")
				fmt.Printf("create pv failed!\n")
				return ctrl.Result{}, err
			}
			fmt.Printf("newpv.Status.Phase: %+v\n", newpv.Status.Phase)
			//成功创建，更新status
			pipeline.Status.DetailPhase.PVPhase = string(newpv.Status.Phase)
		}

		if err2 := r.Get(ctx, pvckey, pvc); err2 != nil { //pvc
			fmt.Printf("pvcIsNotFound\n")
			if k8serrors.IsNotFound(err2) { //not found->create
				fmt.Printf("CreateNewpvc\n")
				if err := r.Create(ctx, newpvc, &client.CreateOptions{}); err != nil {
					logger.Error(err, "create pvc failed!")
					return ctrl.Result{}, err
				}
				fmt.Printf("newpvc.Status.Phase: %+v\n", newpvc.Status.Phase)
				//成功创建，更新status
				pipeline.Status.DetailPhase.PVCPhase = string(newpvc.Status.Phase)
			}
		} else { //exist pv pvc
			// logger.Info("skip update")
			// fmt.Printf("exist pv .OwnerReferences: %+v, pvc .OwnerReferences: %+v\n", pv.OwnerReferences, pvc.OwnerReferences)

			fmt.Printf("find PVC success\n")

			//存在，更新status
			fmt.Printf("PVStatus: %+v\n", pv.Status.Phase)
			fmt.Printf("PVCStatus: %+v\n", pvc.Status.Phase)
			//更新status
			pipeline.Status.DetailPhase.PVPhase = string(pv.Status.Phase)
			pipeline.Status.DetailPhase.PVCPhase = string(pvc.Status.Phase)

		}
		//更新部分 逻辑还需要补充
		// if err := r.Update(ctx, newpv); err != nil { //pv做了更新，pvc也需要更新
		// 	return ctrl.Result{}, err
		// }
		// if err := r.Update(ctx, newpvc); err != nil {
		// 	return ctrl.Result{}, err
		// }

	}

	//处理depolyment
	//查找同名deployment-->exist,not found,
	fmt.Printf("处理deployments ...\n")
	// deployment := &appv1.Deployment{}
	newDeps := *utils.NewDeployments(pipeline)
	createDepFlag := 0

	// var childPipeline_deps []*appv1.Deployment
	// if err := r.List(ctx, &childPipeline_deps, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
	// 	log.Error(err, "unable to list child Deps")
	// 	return ctrl.Result{}, err
	// }

	//key
	deploymentsKey := r.getDeploymentskey(*pipeline)
	// deploys := []appv1.Deployment{}
	for i := 0; i < len(pipeline.Spec.Steps); i++ {
		depkey := deploymentsKey[i]
		// print
		fmt.Printf("deployment %d key: %+v\n", i, depkey)
		deployment := &appv1.Deployment{}
		// deploys = append(deploys, *deployment)
		if err := r.Get(ctx, depkey, deployment); err != nil {
			if k8serrors.IsNotFound(err) { //not found->create
				//存在一个deployment没找到就全都重新创建
				createDepFlag = 1
				fmt.Printf("exist deployment IsNotFound\n")
				break

			}
		} else { //exist
			// fmt.Println("deployment already exists")
			// print
			// fmt.Printf("exist deployment %d :.OwnerReferences: %+v\n", i, deployment.OwnerReferences)
			// fmt.Printf("exist deployment.Status: %+v\n", deployment.Status)
			fmt.Printf("find Deployment %d success\n", i)

			//存在，更新status
			// fmt.Printf("ReadyReplicas: %v, Replicas: %v\n", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
			fmt.Printf("deploymentStatus: %s\n", fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas))
			// 存在，更新status
			pipeline.Status.DetailPhase.StepsPhase[i].DeploymentPhase = fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas)

			// if err := r.Update(ctx, &newDeps[i]); err != nil {
			// 	return ctrl.Result{}, err
			// }

		}

	}

	if createDepFlag != 0 {

		for i := 0; i < len(pipeline.Spec.Steps); i++ {
			// print newDeps[i]
			// fmt.Printf("create deployment: %+v\n", &newDeps[i])

			if err := controllerutil.SetControllerReference(pipeline, &newDeps[i], r.Scheme); err != nil {
				// logger.Error(err, "SetControllerReference failed!")
				fmt.Println(err, "SetControllerReference failed!")

				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, &newDeps[i], &client.CreateOptions{}); err != nil {
				if !k8serrors.IsAlreadyExists(err) {
					logger.Error(err, "create deployment failed!")
					return ctrl.Result{}, err
				} else {
					fmt.Printf("deployment %d IsAlreadyExists \n", i)

				}

			}
			fmt.Printf("createNewDeployment success\n")
			//更新status
			fmt.Printf("newDeps deploymentStatus: %s\n", string(newDeps[i].Status.ReadyReplicas)+"/"+string(newDeps[i].Status.Replicas))

			//成功创建，更新status
			pipeline.Status.DetailPhase.StepsPhase[i].DeploymentPhase = string(newDeps[i].Status.ReadyReplicas) + "/" + string(newDeps[i].Status.Replicas)

		}
	}

	//处理service
	fmt.Printf("处理services ...\n")

	newServices := *utils.NewServices(pipeline)
	createSerFlag := 0
	//key
	serviceskey := r.getServiceskey(*pipeline)
	for i := 0; i < len(pipeline.Spec.Steps); i++ {
		servicekey := serviceskey[i]
		// print
		// fmt.Printf("service %d key: %+v\n", i, &servicekey)
		service := &corev1.Service{}
		if err := r.Get(ctx, servicekey, service); err != nil {
			if k8serrors.IsNotFound(err) { //not found->create
				//存在一个service没找到就全都重新创建
				createSerFlag = 1
				break

			}
		} else { //exist
			//exist
			// fmt.Println("Services already exists")
			// print
			fmt.Printf("find Service %d success\n", i)
			// if err := r.Update(ctx, &newServices[i]); err != nil {
			// 	return ctrl.Result{}, err
			// }

		}

	}
	if createSerFlag != 0 {
		for i := 0; i < len(pipeline.Spec.Steps); i++ {
			if err := controllerutil.SetControllerReference(pipeline, &newServices[i], r.Scheme); err != nil {
				logger.Error(err, "SetControllerReference services failed!")
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, &newServices[i], &client.CreateOptions{}); err != nil {
				if !k8serrors.IsAlreadyExists(err) {
					logger.Error(err, "create service failed!")
					return ctrl.Result{}, err
				}

			}
			fmt.Printf("createNewService success\n")

		}
	}

	//上述都成功创建，更新status
	isAvailable, err := r.checkPipelineAvailable(*pipeline)
	if err != nil {
		fmt.Printf("Error checking pipeline availability: %v\n", err)
	}
	if isAvailable {
		fmt.Println("Pipeline is available.")
		pipeline.Status.Phase = string(distriinferv1.PipelineAvailable) // available  unavailable
		//update status
		if err := r.Status().Update(ctx, pipeline); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		fmt.Println("Pipeline is not available.")
		pipeline.Status.Phase = string(distriinferv1.PipelineUnAvailable) // available  unavailable
		//update status
		if err := r.Status().Update(ctx, pipeline); err != nil {
			return ctrl.Result{}, err
		}
	}

	fmt.Printf("r.Status：%+v\n", pipeline.Status)
	fmt.Printf("处理Reconcile 完成\n")

	return ctrl.Result{}, nil
}

// check Pipeline is Available
func (r *PipelineReconciler) checkPipelineAvailable(pipeline distriinferv1.Pipeline) (bool, error) {
	if pipeline.Status.DetailPhase.PVCPhase == "Bound" && pipeline.Status.DetailPhase.PVPhase == "Bound" {
		for i := 0; i < len(pipeline.Spec.Steps); i++ {
			if err := checkStepAvailability(pipeline, i); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

// check Pipeline'steps is Available
func checkStepAvailability(pipeline distriinferv1.Pipeline, stepIndex int) error {
	deploymentPhase := pipeline.Status.DetailPhase.StepsPhase[stepIndex].DeploymentPhase
	results := strings.Split(deploymentPhase, "/")

	if len(results) != 2 {
		return nomalerrors.New("invalid format for DeploymentPhase")
	}

	readyReplicasStr := results[0]
	replicasStr := results[1]

	readyReplicas, err1 := strconv.Atoi(readyReplicasStr)
	replicas, err2 := strconv.Atoi(replicasStr)

	if err1 != nil || err2 != nil {
		return nomalerrors.New("error converting strings to integers")
	}

	if readyReplicas != replicas {
		return nomalerrors.New("ready replicas not equal to total replicas")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&distriinferv1.Pipeline{}).
		Owns(&corev1.PersistentVolume{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
