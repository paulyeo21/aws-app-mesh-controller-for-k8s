package controllers

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobReconciler struct {
	K8sClient client.Client
}

func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	job := &batchv1.Job{}
	err := r.K8sClient.Get(ctx, req.NamespacedName, job)
	if err != nil {
		return ctrl.Result{}, err
	}

	fmt.Printf("job: %v\n", job)

	pods := &corev1.PodList{}
	err = r.K8sClient.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabels(job.Spec.Template.Labels))
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, p := range pods.Items {
		fmt.Printf("pod meta: %v\n", p.ObjectMeta)
		if len(p.Status.ContainerStatuses) == 0 {
			continue
		}

		exitPod := true

		for _, c := range p.Status.ContainerStatuses {
			fmt.Printf("container name: %v\n", c.Name)
			fmt.Printf("container: %v\n", c)

			if c.Name == "envoy" {
				fmt.Println("found envoy")
				continue
			}

			if c.State.Terminated == nil {
				exitPod = false
				break
			}
		}

		if exitPod {
			fmt.Println("Terminating pod")
			r.K8sClient.Delete(ctx, &p)
		}
	}

	return ctrl.Result{}, nil
}
