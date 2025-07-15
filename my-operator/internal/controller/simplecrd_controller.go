package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	tutosv1alpha1 "dasmlab.org/my-operator/api/v1alpha1"
)

// SimpleCRDReconciler reconciles a SimpleCRD object
type SimpleCRDReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tutos.dasmlab.org,resources=simplecrds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tutos.dasmlab.org,resources=simplecrds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tutos.dasmlab.org,resources=simplecrds/finalizers,verbs=update
func (r *SimpleCRDReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// DASM - I prefer logrus, cause the key/pair of zap is constrainting/debugging nightmare
	_ = logf.FromContext(ctx)
	log.Infof("INIT - SimpleCRD Reconciler called......")

	// We need an Instace of the API Struct to work with and map values/requests to, we also want to make sure it hasnt been deleted
	simpleCrd := &tutosv1alpha1.SimpleCRD{}

	// From the Request (the NamespacedName of the CRD), we will fetch the whole CRD and bind it to our instance of the Variable
	err := r.Get(ctx, req.NamespacedName, simpleCrd)
	if err != nil {
		// if this catches, it was deleted from the cluster before we were able to reconcile
		if apierrors.IsNotFound(err) {
			log.Errorf("SimpleCRD Resource not found or already deleted: %v", err)
			// We exit, since there is "nothing" to reconcile
			return ctrl.Result{}, nil
		}
		// if this catches, something wrong with cluster, check K8s logs, we will retry
		log.Errorf("Unable to Fetch the SimpleCRD from cluster: %v", err)
		return ctrl.Result{}, err
	}

	/// THESE ARE FROM THE REQUESTED CRD (Ephemeral - changes to the REQ are NOT persistent until your write them!)
	/// DASM - i use a "scope" pkg to help with this, when doing "lots of API/CRDs"

	// Add a Log to indicate a CRD was deployed and we are reconcilizing - req.Name is the CRD "name" found on the cluster
	log.Infof("Reconciling a new CRD that was detected - CRD Name (From Cluster): %v", simpleCrd.Name)

	// Lets see what NumPods is set to now?
	log.Infof("CRD: %v - Current Value of NumPods is: %v", simpleCrd.Name, simpleCrd.Spec.NumPods)

	// Lets see when LastChecked "was" (before we set it below)
	log.Infof("CRD: %v - Current Value of LastChecked is: %v", simpleCrd.Name, simpleCrd.Spec.LastChecked)

	// Lets see what iteration we are on
	log.Infof("CRD: %v - Current Iteraion is: %v", simpleCrd.Name, simpleCrd.Spec.NumIter)

	//// OPERATOR LOGIC SECTION ////
	///  We want to do the following in our logic
	///     - we want to count the number of Pods on the current cluster that we are on and update the numPods variable on the CRD on the cluster
	///     - we want to fetch out the Current UTC time (from the internet) and write that as a UTC string to the LastChecked with that value
	///     - we want to increment the Iteration counter and if:
	/// 		- we are less than 10, we set Status to Waiting and throw reconcilation error
	///		- we are at 10, we set Status to Completed and exit, ending reconcilation

	// Get a Client to the cluster we are on (uses InClusterConfig via ServiceAccount RBAC)
	client, err := GetK8sClient()
	if err != nil {
		log.Errorf("Failed to get K8s client, check RBAC/Cluster Configuration !!: %v", err)
		return ctrl.Result{}, err
	}

	// Get the Total Number of Pods from the Cluster
	numPods, err := CountAllPods(ctx, client)
	if err != nil {
		log.Errorf("CountAllPods had and error on CRD %v with error: %v", simpleCrd.Name, err)
		return ctrl.Result{}, err
	}

	log.Info("CountAllPods Returned:", numPods)

	// Assign to our Variable (in our CRD)
	simpleCrd.Spec.NumPods = numPods

	// Get the Current UTC from External Internet Call
	currentTime, err := GetCurrentUTCTime()
	if err != nil {
		log.Errorf("Error fetching time from Internet Source: %v", err)
		return ctrl.Result{}, err
	} else {
		log.Infof("SUCCESS - Current UTC Time is: %v", currentTime)
	}

	// Assign to our Variable (in our CRD)
	simpleCrd.Spec.LastChecked = currentTime

	// Increment our NumIter and handle
	simpleCrd.Spec.NumIter += 1

	if simpleCrd.Spec.NumIter < 10 {
		// Update status to Waiting
		simpleCrd.Status.Waiting = true

		// Persist the Status Updates (waiting/Completed) - WRITE TO CLUSTER!
		if err := client.Status().Update(ctx, simpleCrd); err != nil {
			log.Errorf("Failed to Update (Write) Status of CRD on cluster - Waiting status: %v", err)
			return ctrl.Result{}, err
		}

		// Persist the Spec Updated (NumPods, LastChecked, NumIter) - WRITE TO CLUSTER!
		if err := client.Update(ctx, simpleCrd); err != nil {
			log.Errorf("Failed to Update (Write) Spec of CRD on Cliuster - NumPods/LastCheck/NumIter: %v", err)
			return ctrl.Result{}, err
		}

		// Return error to trigger another "loop" until we reach 10 times (this is just a drawn out demo, you shouldnt really use "counters" in a operator

		return ctrl.Result{}, errors.New("NOT AT 10!! - Retrying as we havent reached 10 iterations!")
	}

	// if NumIter == 10, mark completed and waiting=false and exit
	simpleCrd.Status.Completed = true
	simpleCrd.Status.Waiting = false
	if err := client.Status().Update(ctx, simpleCrd); err != nil {
		log.Errorf("Failed to Update CRD Spec (final increment): %v", err)
		return ctrl.Result{}, err
	}

	// Persis the Spec Update
	if err := client.Update(ctx, simpleCrd); err != nil {
		log.Errorf("Failed to update spec (final increment): %v", err)
		return ctrl.Result{}, err
	}

	log.Infof("Completed After 10 interations - NumIter value: %v", simpleCrd.Spec.NumIter)

	// We are "reconciled" and done!
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleCRDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tutosv1alpha1.SimpleCRD{}).
		Named("simplecrd").
		Complete(r)
}

/// THIS IS BAD PRACTICE, when doing "professional" operators, you should create your own libraries and logic in my-operator/pkg directory and import libraries.  This example is just for simplicity and to keep things in a single file for this tutorial - DONT DO THIS IN PRACTICE! :)

// GetK8sClient returns a controller-runtime client using in-cluster config.
func GetK8sClient() (client.Client, error) {
	// Get the Kubernetes config from inside the cluster
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	// Create a new client using that config
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CountAllPods - This counts all the pods on a cluster, this assumes the service account has the permissions to the API and appropriate Verbs, returns total count (integer)
func CountAllPods(ctx context.Context, c client.Client) (int, error) {
	podList := &corev1.PodList{}
	// List all pods in all namespaces ("")
	if err := c.List(ctx, podList, &client.ListOptions{}); err != nil {
		return 0, err
	}
	return len(podList.Items), nil
}

// / GetCurrentUTC Time - fetches current UTC time from worldtimeapi.org and returns it as a string.
func GetCurrentUTCTime() (string, error) {
	const url = "http://worldtimeapi.org/api/timezone/Etc/UTC"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch time from API")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		UTCDateTime string `json:"utc_datetime"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Optional: parse and reformat as you like
	parsed, err := time.Parse(time.RFC3339, result.UTCDateTime)
	if err != nil {
		// fallback to raw API string if parse fails
		return result.UTCDateTime, nil
	}

	// Return formatted string
	return parsed.UTC().Format("2006-01-02 15:04:05 UTC"), nil
}
