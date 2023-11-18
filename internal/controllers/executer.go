package controllers

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/mohammadne/sanjagh/internal/api/v1alpha1"
)

// executer reconciles a Executer object
type executer struct {
	client.Client
	scheme *runtime.Scheme
	logger *zap.Logger
}

func NewExecuter(client client.Client, scheme *runtime.Scheme, lg *zap.Logger) *executer {
	return &executer{Client: client, scheme: scheme, logger: lg.Named("executer-controller")}
}

//+kubebuilder:rbac:groups=apps.mohammadne.me,resources=executers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.mohammadne.me,resources=executers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.mohammadne.me,resources=executers/finalizers,verbs=update

const executerFinalizer = "apps.mohammadne.me/finalizer"

func (r *executer) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.Named("reconcile")

	// Fetch the Executer instance
	// The purpose is check if the Custom Resource for the Kind Executer
	// is applied on the cluster if not we return nil to stop the reconciliation
	executer := &appsv1alpha1.Executer{}
	if err := r.Get(ctx, req.NamespacedName, executer); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("Executer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Error("Failed to get Executer resource", zap.Error(err))
		return ctrl.Result{}, err
	}

	if executer.GetDeletionTimestamp() != nil {
		if !controllerutil.ContainsFinalizer(executer, executerFinalizer) {
			return ctrl.Result{}, nil
		}

		// do finalizer removal stuffs here...

		log.Info("Removing Finalizer for Executer after successfully perform the operations")
		if ok := controllerutil.RemoveFinalizer(executer, executerFinalizer); !ok {
			err := errors.New("Failed to remove finalizer from the Executer")
			log.Error("Requeue the reconcile loop", zap.Error(err))
			return ctrl.Result{Requeue: true}, nil
		}

		if err := r.Update(ctx, executer); err != nil {
			log.Error("Failed to update the finalizer after removing finalizer from Executer", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	result, err := r.ReconcileDeployment(ctx, req, executer)
	if err != nil || !result.IsZero() {
		return result, err
	}

	if !controllerutil.ContainsFinalizer(executer, executerFinalizer) {
		log.Info("Adding Finalizer for Executer")
		if ok := controllerutil.AddFinalizer(executer, executerFinalizer); !ok {
			err := errors.New("Failed to add finalizer into the Executer")
			log.Error("Requeue the reconcile loop", zap.Error(err))
			return ctrl.Result{Requeue: true}, nil
		}

		// do finalizer stuffs here...

		if err := r.Update(ctx, executer); err != nil {
			log.Error("Failed to update the finalizer after adding finalizer to the Executer", zap.Error(err))
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *executer) ReconcileDeployment(ctx context.Context, req ctrl.Request, executer *appsv1alpha1.Executer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// create desired deployment and add the ownerReference to it
	desiredDeployment := deploymentTemplate(executer)
	if err := ctrl.SetControllerReference(executer, desiredDeployment, r.scheme); err != nil {
		log.Error(err, "Failed to set reference", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	foundDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, foundDeployment); err != nil && apierrors.IsNotFound(err) {
		executer.Status.Phase = appsv1alpha1.PhaseCreating
		if err := r.Status().Update(ctx, executer); err != nil {
			log.Error(err, "Failed to update deployment state", "NamespacedName", req.NamespacedName.String())
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Deployment", "NamespacedName", req.NamespacedName.String())
		if err = r.Create(ctx, desiredDeployment); err != nil {
			executer.Status.Phase = appsv1alpha1.PhaseFailed
			if err := r.Status().Update(ctx, executer); err != nil {
				log.Error(err, "Failed to update deployment state", "NamespacedName", req.NamespacedName.String())
				return ctrl.Result{}, err
			}

			log.Error(err, "Failed to create new Deployment", "NamespacedName", req.NamespacedName.String())
			return ctrl.Result{}, err
		}

		// We will requeue the reconciliation so that we can ensure the state and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update existing deployment spec
	if *foundDeployment.Spec.Replicas != *desiredDeployment.Spec.Replicas {
		executer.Status.Phase = appsv1alpha1.PhaseUpdating
		if err := r.Status().Update(ctx, executer); err != nil {
			log.Error(err, "Failed to update deployment state", "NamespacedName", req.NamespacedName.String())
			return ctrl.Result{}, err
		}

		log.Info("Updating executer's deployment replicas", "found", foundDeployment.Spec.Replicas, "desired", desiredDeployment.Spec.Replicas)
		foundDeployment.Spec.Replicas = desiredDeployment.Spec.Replicas
		if err := r.Update(ctx, foundDeployment); err != nil {
			if strings.Contains(err.Error(), genericregistry.OptimisticLockErrorMsg) {
				return reconcile.Result{RequeueAfter: time.Millisecond * 500}, nil
			}

			executer.Status.Phase = appsv1alpha1.PhaseFailed
			if err := r.Status().Update(ctx, executer); err != nil {
				log.Error(err, "Failed to update deployment state", "NamespacedName", req.NamespacedName.String())
				return ctrl.Result{}, err
			}

			log.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		}
	}

	if executer.Status.Phase != appsv1alpha1.PhaseCreated {
		executer.Status.Phase = appsv1alpha1.PhaseCreated
		if err := r.Status().Update(ctx, executer); err != nil {
			log.Error(err, "Failed to update deployment state", "NamespacedName", req.NamespacedName.String())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func labels(executer *appsv1alpha1.Executer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "Executer",
		"app.kubernetes.io/instance":   executer.Name,
		"app.kubernetes.io/part-of":    "sanjagh",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

func deploymentTemplate(executer *appsv1alpha1.Executer) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      executer.Name,
			Namespace: executer.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &executer.Spec.Replication,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(executer),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels(executer),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            executer.Name,
							Image:           executer.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         executer.Spec.Commands,
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *executer) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Executer{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
