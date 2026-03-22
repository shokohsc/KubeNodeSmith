package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubenodesmithv1alpha1 "github.com/StealthBadger747/KubeNodeSmith/api/v1alpha1"
)

func TestPodMatchesPoolTreatsEmptyValueMachineTemplateLabelAsTargeting(t *testing.T) {
	pool := &kubenodesmithv1alpha1.NodeSmithPool{
		Spec: kubenodesmithv1alpha1.NodeSmithPoolSpec{
			PoolLabelKey: "topology.kubenodesmith.io/pool",
			MachineTemplate: kubenodesmithv1alpha1.MachineTemplate{
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
				},
			},
		},
	}
	poolLabels := buildPoolLabelSet(pool)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "worker-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"node-role.kubernetes.io/worker": "",
			},
		},
	}

	matches, requires := podMatchesPool(pod, pool, poolLabels)
	if !matches {
		t.Fatalf("expected pod to match pool labels")
	}
	if !requires {
		t.Fatalf("expected empty-value machine template label to mark pod as pool-targeting")
	}
}

func TestPodMatchesPoolTreatsEmptyValueAffinityLabelAsTargeting(t *testing.T) {
	pool := &kubenodesmithv1alpha1.NodeSmithPool{
		Spec: kubenodesmithv1alpha1.NodeSmithPoolSpec{
			PoolLabelKey: "topology.kubenodesmith.io/pool",
			MachineTemplate: kubenodesmithv1alpha1.MachineTemplate{
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
				},
			},
		},
	}
	poolLabels := buildPoolLabelSet(pool)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "worker-affinity-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "node-role.kubernetes.io/worker",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{""},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// This exercises the podMatchesPool -> nodeSelectorTermMatches affinity path.
	matches, requires := podMatchesPool(pod, pool, poolLabels)
	if !matches {
		t.Fatalf("expected affinity-constrained pod to match pool labels")
	}
	if !requires {
		t.Fatalf("expected empty-value affinity label to mark pod as pool-targeting")
	}
}
