package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// AnomalyDetectedTotal counts total anomalies detected
	AnomalyDetectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_remediator_anomalies_detected_total",
			Help: "Total number of anomalies detected by the AnomalyRemediator",
		},
		[]string{"namespace", "anomalyaction", "source"},
	)

	// ActionableAnomalyTotal counts anomalies that are actionable (i.e., not in cooldown and not pending approval).
	// This is useful for demo parity with remediation metrics.
	ActionableAnomalyTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_remediator_actionable_anomalies_total",
			Help: "Total number of actionable anomalies (past cooldown / not pending approval) detected by the AnomalyRemediator",
		},
		[]string{"namespace", "anomalyaction", "source"},
	)

	// RemediationExecutedTotal counts total remediations executed
	RemediationExecutedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_remediator_remediations_executed_total",
			Help: "Total number of remediations executed by the AnomalyRemediator",
		},
		[]string{"namespace", "anomalyaction", "remediation_type"},
	)

	// RemediationDurationSeconds tracks remediation execution time
	RemediationDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "anomaly_remediator_remediation_duration_seconds",
			Help:    "Time taken to execute remediation actions",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "anomalyaction", "remediation_type"},
	)

	// PodRestartsTotal counts pod restarts performed
	PodRestartsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_remediator_pod_restarts_total",
			Help: "Total number of pods restarted by the AnomalyRemediator",
		},
		[]string{"namespace", "anomalyaction", "pod_name"},
	)
)

func init() {
	// Register custom metrics with controller-runtime's metrics registry
	metrics.Registry.MustRegister(
		AnomalyDetectedTotal,
		ActionableAnomalyTotal,
		RemediationExecutedTotal,
		RemediationDurationSeconds,
		PodRestartsTotal,
	)
}
