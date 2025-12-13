package main

import (
	"context"
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	aiopsv1alpha1 "github.com/prophet-aiops/autonomous-agent/api/v1alpha1"
	"github.com/prophet-aiops/autonomous-agent/controllers"
	"github.com/prophet-aiops/autonomous-agent/mcp-server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(aiopsv1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var mcpPort int
	var mcpTLSEnabled bool
	var mcpTLSPort int
	var mcpTLSCertFile string
	var mcpTLSKeyFile string
	var mcpTLSClientCAFile string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.IntVar(&mcpPort, "mcp-port", 8082, "The port for MCP server.")
	flag.BoolVar(&mcpTLSEnabled, "mcp-tls-enabled", false, "Enable HTTPS for the MCP server.")
	flag.IntVar(&mcpTLSPort, "mcp-tls-port", 8443, "The port for MCP server HTTPS listener.")
	flag.StringVar(&mcpTLSCertFile, "mcp-tls-cert-file", "", "Path to MCP server TLS certificate file (PEM).")
	flag.StringVar(&mcpTLSKeyFile, "mcp-tls-key-file", "", "Path to MCP server TLS private key file (PEM).")
	flag.StringVar(&mcpTLSClientCAFile, "mcp-tls-client-ca-file", "", "Optional path to client CA bundle (PEM) to enable mTLS.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "autonomous-agent.prophet.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Register field indexers for cache-backed client queries.
	// This is required for client.MatchingFields{"spec.nodeName": <node>} lookups on Pods.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(rawObj client.Object) []string {
		pod, ok := rawObj.(*corev1.Pod)
		if !ok {
			return nil
		}
		if pod.Spec.NodeName == "" {
			return nil
		}
		return []string{pod.Spec.NodeName}
	}); err != nil {
		setupLog.Error(err, "unable to set up field indexer", "field", "spec.nodeName", "object", "Pod")
		os.Exit(1)
	}

	// Initialize MCP server
	mcpSrv := mcpserver.NewMCPServer(mgr.GetClient(), nil)
	go func() {
		if err := mcpSrv.Start(ctrl.SetupSignalHandler(), mcpPort); err != nil {
			setupLog.Error(err, "unable to start MCP server")
		}
	}()
	if mcpTLSEnabled {
		go func() {
			if err := mcpSrv.StartTLS(ctrl.SetupSignalHandler(), mcpTLSPort, mcpserver.TLSOptions{
				CertFile:     mcpTLSCertFile,
				KeyFile:      mcpTLSKeyFile,
				ClientCAFile: mcpTLSClientCAFile,
			}); err != nil {
				setupLog.Error(err, "unable to start MCP server HTTPS")
			}
		}()
	}

	// Initialize action executor
	actionExecutor := controllers.NewActionExecutor(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("ActionExecutor"),
	)

	if err = (&controllers.AutonomousActionReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		Log:            ctrl.Log.WithName("controllers").WithName("AutonomousAction"),
		MCPServer:      mcpSrv,
		ActionExecutor: actionExecutor,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AutonomousAction")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
