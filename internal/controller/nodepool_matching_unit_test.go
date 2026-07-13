package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func TestDetermineNodeCapacityIgnoresClaimsFromOtherPools(t *testing.T) {
	claims := &kubenodesmithv1alpha1.NodeSmithClaimList{Items: []kubenodesmithv1alpha1.NodeSmithClaim{
		{
			Spec: kubenodesmithv1alpha1.NodeSmithClaimSpec{
				PoolRef: "other-pool",
				Requirements: &kubenodesmithv1alpha1.NodeSmithClaimRequirements{
					CPUCores:  40,
					MemoryMiB: 81920,
				},
			},
		},
	}}
	pods := []corev1.Pod{{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("16"),
				corev1.ResourceMemory: resource.MustParse("40Gi"),
			}},
		}}},
	}}

	capacity := determineNodeCapacity("mayastor-pool", nil, claims, pods)
	if capacity.cpuMilli != 16000 || capacity.memBytes != 40*1024*1024*1024 {
		t.Fatalf("expected pending workload capacity (16 CPU, 40 GiB), got (%d CPU millicores, %d bytes)", capacity.cpuMilli, capacity.memBytes)
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
