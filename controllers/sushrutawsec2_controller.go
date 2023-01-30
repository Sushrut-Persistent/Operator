/*
Copyright 2023.

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

package aws

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	corev1 "k8s.io/api/core/v1"

	awsv1 "github.com/Sushrut-Persistent/OperatorPOC.git/apis/aws/v1"
)

// SushrutAWSEC2Reconciler reconciles a SushrutAWSEC2 object
type SushrutAWSEC2Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aws.sushrut.com,resources=sushrutawsec2s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aws.sushrut.com,resources=sushrutawsec2s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aws.sushrut.com,resources=sushrutawsec2s/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/finalizers;jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SushrutAWSEC2 object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

const AWSEC2Finalizer = "aws.sushrut.com/finalizer"

func (rec *SushrutAWSEC2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logs := log.FromContext(ctx)
	logs.Info("Reconciling EC2 CRs")

	// TODO(user): your logic here
	awsInstance := &awsv1.AWSEC2{}	//Fetch Instance
	logs.Info("Name: " + req.NamespacedName.Name)

	err := rec.Client.Get(ctx, req.NamespacedName, awsInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logs.Info("awsInstance resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logs.Error("Failed to get awsEC2", err)
		return ctrl.Result{}, err
	}

	// Search what 'found' is
	found := &batchv1.Job{}
	err = rec.Client.Get(ctx, types.NamespacedName{Name: awsInstance.Name + "create", Namespace: awsInstance.Namespace}, found)
	logs.Info("BatchV1.Job Info: ", found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		job := rec.JobForAWSEC2(awsInstance, "create")
		logs.Info("Creating a new Job\njob.Namespace: ", job.Namespace, "\njob.Name: ", job.Name)
		err = rec.Client.Create(ctx, job)
		if err != nil {
			logs.Error("Failed to create new Job with namespace: ", job.Namespace, " and name: ", job.Name " because ", err)
			return ctrl.Result{}, err
		}
		// job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logs.Error("Failed to get job because ", err)
		return ctrl.Result{}, err
	}

	instanceToBeDeleted := awsInstance.GetDeletionTimestamp() != nil
	if instanceToBeDeleted {
		if controllerutil.ContainsFinalizer(awsInstance, AWSEC2Finalizer) {
			// Run finalization logic for AWSEC2Finalizer. If the finalization logic fails, don't remove the finalizer so that we can retry during the next reconciliation.
			logs.Info(awsInstance.Name, " is marked for deletion")
			if err := rec.finalizeAWSEC2(ctx, awsInstance); err != nil {
				return ctrl.Result{}, err
			}

			// Remove AWSEC2Finalizer
			//Once all finalizers have been removed, the object will be deleted.
			controllerutil.RemoveFinalizer(awsInstance, AWSEC2Finalizer)
			err := rec.Client.Update(ctx, awsInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			logs.Info("Finalizer for ", awsInstance.Name, " removed")
		}
		return ctrl.Result{}, nil
	}

	// If not marked for deletion, add finalizer for this CR
	if !controllerutil.ContainsFinalizer(awsInstance, AWSEC2Finalizer) {
		logs.Info(awsInstance.Name, " finalizer added again")
		controllerutil.AddFinalizer(awsInstance, AWSEC2Finalizer)
		err = rec.Client.Update(ctx, awsInstance)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (rec *AWSEC2Reconciler) finalizeAWSEC2(ctx context.Context, awsInstance *awsv1.AWSEC2) error {
	// Add the cleanup steps that the operator needs to do before the CR can be deleted. 
	//Examples of finalizers include performing backups and deleting resources that are not owned by this CR, like a PVC.
	logs := log.FromContext(ctx)
	logs.Info("Successfully finalized AWSEC2")

	found := &batchv1.Job{}
	err := rec.Client.Get(ctx, types.NamespacedName{Name: awsInstance.Name + "delete", Namespace: awsInstance.Namespace}, found)
	logs.Info("BatchV1.Job Info: ", found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		job := rec.JobForAWSEC2(awsInstance, "delete")
		logs.Info("Creating a new Job\njob.Namespace: ", job.Namespace, "\njob.Name: ", job.Name)
		err = rec.Client.Create(ctx, job)
		if err != nil {
			logs.Error("Failed to create new Job with namespace: ", job.Namespace, " and name: ", job.Name " because ", err)
			return err
		}
		// job created successfully - return and requeue
		return nil
	} else if err != nil {
		logs.Error("Failed to get job because ", err)
		return err
	}
	return nil
}

func AWSEC2Labels(v *awsv1.AWSEC2, tier string) map[string]string {
	return map[string]string{
		"app":       "AWSEC2",
		"AWSEC2_cr": v.Name,
		"tier":      tier,
	}
}

// Job Spec
func (rec *AWSEC2Reconciler) JobForAWSEC2(awsInstance *awsv1.AWSEC2, command string) *batchv1.Job {
	jobName := awsInstance.Name + command
	job := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: awsInstance.Namespace,
			// Check if tier should be awsEC2 or awsInstance
			Labels:    AWSEC2Labels(awsInstance, "awsEC2"),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  awsInstance.Name,
						Image: awsInstance.Spec.Image,
						Env: []corev1.EnvVar{
							{
								Name:  "ec2_command",
								Value: command,
							},
							{
								Name: "AWS_ACCESS_KEY_ID",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-access-key-id",
									},
								},
							},
							{
								Name: "AWS_SECRET_ACCESS_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-secret-access-key",
									},
								},
							},
							{
								Name: "AWS_DEFAULT_REGION",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-default-region",
									},
								},
							}},
					}},
					// RestartPolicy: "OnFailure",
					//ImagePullPolicy
				},
			},
		},
	}
	// Set awsEC2 instance as the owner and controller
	ctrl.SetControllerReference(awsInstance, job, rec.Scheme)
	return job
}

// SetupWithManager sets up the controller with the Manager.
func (rec *SushrutAWSEC2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&awsv1.SushrutAWSEC2{}).
		Owns(&batchv1.Job{}).
		Complete(rec)
}
