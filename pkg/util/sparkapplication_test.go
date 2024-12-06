/*
Copyright 2024 The Kubeflow authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/spark-operator/api/v1beta2"
	"github.com/kubeflow/spark-operator/pkg/common"
	"github.com/kubeflow/spark-operator/pkg/util"
)

var _ = Describe("GetDriverPodName", func() {
	Context("SparkApplication without driver pod name field and driver pod name conf", func() {
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
		}

		It("Should return the default driver pod name", func() {
			Expect(util.GetDriverPodName(app)).To(Equal("test-app-driver"))
		})
	})

	Context("SparkApplication with only driver pod name field", func() {
		driverPodName := "test-app-driver-pod"
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				Driver: v1beta2.DriverSpec{
					PodName: &driverPodName,
				},
			},
		}

		It("Should return the driver pod name from driver spec", func() {
			Expect(util.GetDriverPodName(app)).To(Equal(driverPodName))
		})
	})

	Context("SparkApplication with only driver pod name conf", func() {
		driverPodName := "test-app-driver-pod"
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				SparkConf: map[string]string{
					common.SparkKubernetesDriverPodName: driverPodName,
				},
			},
		}

		It("Should return the driver name from spark conf", func() {
			Expect(util.GetDriverPodName(app)).To(Equal(driverPodName))
		})
	})

	Context("SparkApplication with both driver pod name field and driver pod name conf", func() {
		driverPodName1 := "test-app-driver-1"
		driverPodName2 := "test-app-driver-2"
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				SparkConf: map[string]string{
					common.SparkKubernetesDriverPodName: driverPodName1,
				},
				Driver: v1beta2.DriverSpec{
					PodName: &driverPodName2,
				},
			},
		}

		It("Should return the driver pod name from driver spec", func() {
			Expect(util.GetDriverPodName(app)).To(Equal(driverPodName2))
		})
	})
})

var _ = Describe("GetApplicationState", func() {
	Context("SparkApplication with completed state", func() {
		app := &v1beta2.SparkApplication{
			Status: v1beta2.SparkApplicationStatus{
				AppState: v1beta2.ApplicationState{
					State: v1beta2.ApplicationStateCompleted,
				},
			},
		}

		It("Should return completed state", func() {
			Expect(util.GetApplicationState(app)).To(Equal(v1beta2.ApplicationStateCompleted))
		})
	})
})

var _ = Describe("IsExpired", func() {
	Context("SparkApplication without TTL", func() {
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
		}

		It("Should return false", func() {
			Expect(util.IsExpired(app)).To(BeFalse())
		})
	})

	Context("SparkApplication not terminated with TTL", func() {
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				TimeToLiveSeconds: util.Int64Ptr(3600),
			},
		}

		It("Should return false", func() {
			Expect(util.IsExpired(app)).To(BeFalse())
		})
	})

	Context("SparkApplication terminated with TTL not expired", func() {
		now := time.Now()
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				TimeToLiveSeconds: util.Int64Ptr(3600),
			},
			Status: v1beta2.SparkApplicationStatus{
				TerminationTime: metav1.NewTime(now.Add(-30 * time.Minute)),
			},
		}

		It("Should return false", func() {
			Expect(util.IsExpired(app)).To(BeFalse())
		})
	})

	Context("SparkApplication terminated with TTL expired", func() {
		now := time.Now()
		app := &v1beta2.SparkApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: v1beta2.SparkApplicationSpec{
				TimeToLiveSeconds: util.Int64Ptr(3600),
			},
			Status: v1beta2.SparkApplicationStatus{
				TerminationTime: metav1.NewTime(now.Add(-2 * time.Hour)),
			},
		}

		It("Should return true", func() {
			Expect(util.IsExpired(app)).To(BeTrue())
		})
	})
})

var _ = Describe("IsDriverRunning", func() {
	Context("SparkApplication with completed state", func() {
		app := &v1beta2.SparkApplication{
			Status: v1beta2.SparkApplicationStatus{
				AppState: v1beta2.ApplicationState{
					State: v1beta2.ApplicationStateCompleted,
				},
			},
		}

		It("Should return false", func() {
			Expect(util.IsDriverRunning(app)).To(BeFalse())
		})
	})

	Context("SparkApplication with running state", func() {
		app := &v1beta2.SparkApplication{
			Status: v1beta2.SparkApplicationStatus{
				AppState: v1beta2.ApplicationState{
					State: v1beta2.ApplicationStateRunning,
				},
			},
		}

		It("Should return true", func() {
			Expect(util.IsDriverRunning(app)).To(BeTrue())
		})
	})
})

var _ = Describe("GetLocalVolumes", func() {
	Context("SparkApplication with local volumes", func() {
		volume1 := corev1.Volume{
			Name: "local-volume",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/tmp",
				},
			},
		}

		volume2 := corev1.Volume{
			Name: "spark-local-dir-1",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/spark-local-dir-1",
				},
			},
		}

		volume3 := corev1.Volume{
			Name: "spark-local-dir-2",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/spark-local-dir-2",
				},
			},
		}

		app := &v1beta2.SparkApplication{
			Spec: v1beta2.SparkApplicationSpec{
				Volumes: []corev1.Volume{
					volume1,
					volume2,
					volume3,
				},
			},
		}

		It("Should return volumes with the correct prefix", func() {
			volumes := util.GetLocalVolumes(app)
			expected := map[string]corev1.Volume{
				volume2.Name: volume2,
				volume3.Name: volume3,
			}
			Expect(volumes).To(Equal(expected))
		})
	})
})

var _ = Describe("GetDefaultUIServiceName", func() {
	app := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test-namespace",
		},
	}

	It("Should return the default UI service name", func() {
		Expect(util.GetDefaultUIServiceName(app)).To(Equal("test-app-ui-svc"))
	})

	appWithLongName := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-with-a-long-name-that-would-be-over-63-characters",
			Namespace: "test-namespace",
		},
	}

	It("Should truncate the app name so the service name is below 63 characters", func() {
		serviceName := util.GetDefaultUIServiceName(appWithLongName)
		Expect(len(serviceName)).To(BeNumerically("<=", 63))
		Expect(serviceName).To(HavePrefix(appWithLongName.Name[:47]))
		Expect(serviceName).To(HaveSuffix("-ui-svc"))
	})
})

var _ = Describe("GetDefaultUIIngressName", func() {
	app := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test-namespace",
		},
	}

	It("Should return the default UI ingress name", func() {
		Expect(util.GetDefaultUIIngressName(app)).To(Equal("test-app-ui-ingress"))
	})

	appWithLongName := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-with-a-long-name-that-would-be-over-63-characters",
			Namespace: "test-namespace",
		},
	}

	It("Should truncate the app name so the ingress name is below 63 characters", func() {
		serviceName := util.GetDefaultUIIngressName(appWithLongName)
		Expect(len(serviceName)).To(BeNumerically("<=", 63))
		Expect(serviceName).To(HavePrefix(appWithLongName.Name[:42]))
		Expect(serviceName).To(HaveSuffix("-ui-ingress"))
	})
})

var _ = Describe("IsDriverTerminated", func() {
	It("Should check whether driver is terminated", func() {
		Expect(util.IsDriverTerminated(v1beta2.DriverStatePending)).To(BeFalse())
		Expect(util.IsDriverTerminated(v1beta2.DriverStateRunning)).To(BeFalse())
		Expect(util.IsDriverTerminated(v1beta2.DriverStateCompleted)).To(BeTrue())
		Expect(util.IsDriverTerminated(v1beta2.DriverStateFailed)).To(BeTrue())
		Expect(util.IsDriverTerminated(v1beta2.DriverStateUnknown)).To(BeFalse())
	})
})

var _ = Describe("IsExecutorTerminated", func() {
	It("Should check whether executor is terminated", func() {
		Expect(util.IsExecutorTerminated(v1beta2.ExecutorStatePending)).To(BeFalse())
		Expect(util.IsExecutorTerminated(v1beta2.ExecutorStateRunning)).To(BeFalse())
		Expect(util.IsExecutorTerminated(v1beta2.ExecutorStateCompleted)).To(BeTrue())
		Expect(util.IsExecutorTerminated(v1beta2.ExecutorStateFailed)).To(BeTrue())
		Expect(util.IsExecutorTerminated(v1beta2.ExecutorStateUnknown)).To(BeFalse())
	})
})

var _ = Describe("DriverStateToApplicationState", func() {
	It("Should convert driver state to application state correctly", func() {
		Expect(util.DriverStateToApplicationState(v1beta2.DriverStatePending)).To(Equal(v1beta2.ApplicationStateSubmitted))
		Expect(util.DriverStateToApplicationState(v1beta2.DriverStateRunning)).To(Equal(v1beta2.ApplicationStateRunning))
		Expect(util.DriverStateToApplicationState(v1beta2.DriverStateCompleted)).To(Equal(v1beta2.ApplicationStateSucceeding))
		Expect(util.DriverStateToApplicationState(v1beta2.DriverStateFailed)).To(Equal(v1beta2.ApplicationStateFailing))
		Expect(util.DriverStateToApplicationState(v1beta2.DriverStateUnknown)).To(Equal(v1beta2.ApplicationStateUnknown))
	})
})

var _ = Describe("GetDriverState", func() {
	Context("With no MonitoredSidecars configured", func() {
		It("Should return the correct driver state when the pod is pending", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			}
			Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStatePending))
		})

		It("Should return the correct driver state when the pod is succeeded", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			}
			Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateCompleted))
		})

		It("Should return the correct driver state when the pod state is unknown", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodUnknown,
				},
			}
			Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateUnknown))
		})

		Context("The driver pod has the Spark container only", func() {
			It("Should return the correct driver state when the pod is running", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateRunning))
			})

			It("Should return the correct driver state when the pod is failed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateFailed))
			})
		})

		Context("The driver pod has a sidecar as well apart from the Spark container", func() {
			It("Should return the correct driver state when the pod is running and the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is running and the Spark container failed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is running and a sidecar container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateRunning))
			})

			It("Should return the correct driver state when the pod is running and a sidecar container failed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateRunning))
			})

			It("Should return the correct driver state when the pod is failed but the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is failed due to a Spark container failure", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
							{
								Name: "sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, nil, nil)).To(Equal(v1beta2.DriverStateFailed))
			})
		})
	})

	Context("With a single element MonitoredSidecars configured", func() {
		monitoredSidecar := "monitored-sidecar"
		Context("With FailOnMonitoredSidecarZeroExitCode being false", func() {
			failOnZeroExitCode := false
			It("Should return the correct driver state when the pod is failed due to a non-monitored sidecar but the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is failed due to a monitored sidecar but the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is failed due to a Spark container failure and there is a successful non-monitored sidecar", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is failed due to a Spark container failure and there is a successful monitored sidecar", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is running and a non-monitored sidecar container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateRunning))
			})

			It("Should return the correct driver state when the pod is running and a monitored sidecar container failed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is running and a monitored sidecar container completed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateRunning))
			})

		})

		Context("With FailOnMonitoredSidecarZeroExitCode being true", func() {
			failOnZeroExitCode := true
			It("Should return the correct driver state when the pod is failed due to a non-monitored sidecar but the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is failed due to a monitored sidecar but the Spark container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateCompleted))
			})

			It("Should return the correct driver state when the pod is failed due to a Spark container failure and there is a successful non-monitored sidecar", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is failed due to a Spark container failure and there is a successful monitored sidecar", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: common.SparkDriverContainerName,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is running and a non-monitored sidecar container completed successfully", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "non-monitored-sidecar",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateRunning))
			})

			It("Should return the correct driver state when the pod is running and a monitored sidecar container failed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 1,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})

			It("Should return the correct driver state when the pod is running and a monitored sidecar container completed", func() {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: monitoredSidecar,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										ExitCode: 0,
									},
								},
							},
						},
					},
				}
				Expect(util.GetDriverState(pod, &monitoredSidecar, &failOnZeroExitCode)).To(Equal(v1beta2.DriverStateFailed))
			})
		})

	})
})

var _ = Describe("GetDriverState2", func(){
	Context("The driver pod has no sidecar container", func() {
		paramFalse := false
		paramTrue := true
		paramEmpty := ""
		DescribeTable("Should return the correct driver state when the container is not terminated",
			func(podPhase corev1.PodPhase, monitoredSidecars *string, failOnMonitoredSidecarZeroExitCode *bool, expectedDriverState v1beta2.DriverState) {
				pod := &corev1.Pod{
					Status: corev1.PodStatus{
						Phase: podPhase,
					},
				}
				Expect(util.GetDriverState(pod, monitoredSidecars, failOnMonitoredSidecarZeroExitCode)).To(Equal(expectedDriverState))
			},
			Entry("PodPending, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=nil", corev1.PodPending, nil, nil, v1beta2.DriverStatePending),
			Entry("PodPending, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=false", corev1.PodPending, nil, &paramFalse, v1beta2.DriverStatePending),
			Entry("PodPending, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=true", corev1.PodPending, nil, &paramTrue, v1beta2.DriverStatePending),
			Entry("PodPending, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=nil", corev1.PodPending, &paramEmpty, nil, v1beta2.DriverStatePending),
			Entry("PodPending, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=false", corev1.PodPending, &paramEmpty, &paramFalse, v1beta2.DriverStatePending),
			Entry("PodPending, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=true", corev1.PodPending, &paramEmpty, &paramTrue, v1beta2.DriverStatePending),

			Entry("PodSucceeded, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=nil", corev1.PodSucceeded, nil, nil, v1beta2.DriverStateCompleted),
			Entry("PodSucceeded, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=false", corev1.PodSucceeded, nil, &paramFalse, v1beta2.DriverStateCompleted),
			Entry("PodSucceeded, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=true", corev1.PodSucceeded, nil, &paramTrue, v1beta2.DriverStateCompleted),
			Entry("PodSucceeded, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=nil", corev1.PodSucceeded, &paramEmpty, nil, v1beta2.DriverStateCompleted),
			Entry("PodSucceeded, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=false", corev1.PodSucceeded, &paramEmpty, &paramFalse, v1beta2.DriverStateCompleted),
			Entry("PodSucceeded, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=true", corev1.PodSucceeded, &paramEmpty, &paramTrue, v1beta2.DriverStateCompleted),

			Entry("PodUnknown, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=nil", corev1.PodUnknown, nil, nil, v1beta2.DriverStateUnknown),
			Entry("PodUnknown, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=false", corev1.PodUnknown, nil, &paramFalse, v1beta2.DriverStateUnknown),
			Entry("PodUnknown, monitoredSidecars=nil, failOnMonitoredSidecarZeroExitCode=true", corev1.PodUnknown, nil, &paramTrue, v1beta2.DriverStateUnknown),
			Entry("PodUnknown, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=nil", corev1.PodUnknown, &paramEmpty, nil, v1beta2.DriverStateUnknown),
			Entry("PodUnknown, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=false", corev1.PodUnknown, &paramEmpty, &paramFalse, v1beta2.DriverStateUnknown),
			Entry("PodUnknown, monitoredSidecars='', failOnMonitoredSidecarZeroExitCode=true", corev1.PodUnknown, &paramEmpty, &paramTrue, v1beta2.DriverStateUnknown),
		)
	})
})
