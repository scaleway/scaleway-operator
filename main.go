/*


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

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	rdbv1alpha1 "github.com/scaleway/scaleway-operator/apis/rdb/v1alpha1"
	"github.com/scaleway/scaleway-operator/controllers"
	rdbcontroller "github.com/scaleway/scaleway-operator/controllers/rdb"
	rdbmanager "github.com/scaleway/scaleway-operator/pkg/manager/rdb"
	"github.com/scaleway/scaleway-operator/webhooks"
	rdbwebhook "github.com/scaleway/scaleway-operator/webhooks/rdb"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = rdbv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "e0607a04.scaleway.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	scwClient, err := scw.NewClient(scw.WithEnv())
	if err != nil {
		setupLog.Error(err, "unable to create scw client")
		os.Exit(1)
	}

	if err = (&rdbcontroller.RDBInstanceReconciler{
		ScalewayReconciler: &controllers.ScalewayReconciler{
			Client:   mgr.GetClient(),
			Log:      ctrl.Log.WithName("controllers").WithName("RDBInstance"),
			Recorder: mgr.GetEventRecorderFor("RDBInstance"),
			Scheme:   mgr.GetScheme(),
			ScalewayManager: &rdbmanager.InstanceManager{
				API:    rdb.NewAPI(scwClient),
				Client: mgr.GetClient(),
				Log:    ctrl.Log.WithName("manager").WithName("RDBInstance"),
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RDBInstance")
		os.Exit(1)
	}

	if err = (&rdbcontroller.RDBDatabaseReconciler{
		ScalewayReconciler: &controllers.ScalewayReconciler{
			Client:   mgr.GetClient(),
			Log:      ctrl.Log.WithName("controllers").WithName("RDBDatabase"),
			Recorder: mgr.GetEventRecorderFor("RDBDatabase"),
			Scheme:   mgr.GetScheme(),
			ScalewayManager: &rdbmanager.DatabaseManager{
				API:    rdb.NewAPI(scwClient),
				Client: mgr.GetClient(),
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RDBDatabase")
		os.Exit(1)
	}

	if err = (&rdbcontroller.RDBUserReconciler{
		ScalewayReconciler: &controllers.ScalewayReconciler{
			Client:   mgr.GetClient(),
			Log:      ctrl.Log.WithName("controllers").WithName("RDBUser"),
			Recorder: mgr.GetEventRecorderFor("RDBDatabase"),
			Scheme:   mgr.GetScheme(),
			ScalewayManager: &rdbmanager.DatabaseManager{
				API:    rdb.NewAPI(scwClient),
				Client: mgr.GetClient(),
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RDBUser")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&rdbwebhook.RDBInstanceValidator{
			Log: ctrl.Log.WithName("webhooks").WithName("RDBInstance"),
			ScalewayWebhook: &webhooks.ScalewayWebhook{
				ScalewayManager: &rdbmanager.InstanceManager{
					API:    rdb.NewAPI(scwClient),
					Client: mgr.GetClient(),
				},
			},
		}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RDBInstance")
			os.Exit(1)
		}

		if err = (&rdbwebhook.RDBDatabaseValidator{
			Log: ctrl.Log.WithName("webhooks").WithName("RDBDatabase"),
			ScalewayWebhook: &webhooks.ScalewayWebhook{
				ScalewayManager: &rdbmanager.DatabaseManager{
					API:    rdb.NewAPI(scwClient),
					Client: mgr.GetClient(),
				},
			},
		}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RDBDatabase")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
