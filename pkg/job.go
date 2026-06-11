package pkg

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"github.com/aws/smithy-go/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	// appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

func ToKubeblockInstallJob(saName string) *batchv1.Job {
	shellStr := `
kubectl apply -f $KO_DATA_PATH/crds-kubeblocks/1.0.1/ --server-side
kubectl apply -f $KO_DATA_PATH/crds-kubeblocks/snapshot/ --server-side
helm upgrade kubeblocks $KO_DATA_PATH/charts/kubeblocks-1.0.1.tgz -n kb-system --create-namespace --install
`
	cmd := []string{"/bin/sh", "-c", shellStr}
	jobName := "kubeblock-job-" + RandomString(8)

	labels := map[string]string{
		"job":       jobName,
		"kubeblock": "install",
	}
	annotations := map[string]string{}
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "kubeblocks-install",
					Image: SelfImage(),
					// Env:             ,
					Command:         cmd,
					ImagePullPolicy: corev1.PullIfNotPresent,
				},
			},
			ServiceAccountName: saName,
		},
	}
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.Int32(300),
			Template:                pod,
		},
	}
	return job
}
