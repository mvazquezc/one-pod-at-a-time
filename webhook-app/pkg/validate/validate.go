// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package validate

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
        "context"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/kubernetes"
)

// Validate validates
func Validate(body []byte, verbose bool) ([]byte, error) {
	if verbose {
		log.Printf("recv: %s\n", string(body))
	}

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

        config, err := rest.InClusterConfig()
        if err != nil {
                fmt.Errorf("Error getting in-cluster config")
                panic(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                fmt.Errorf("Error creating the clientset")
                panic(err)
        }

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}
	if ar != nil {

		// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		// set response options
		resp.Allowed = true // allow pods by default
		resp.UID = ar.UID

		// add some audit annotations, helpful to know why a object was reviewed
		resp.AuditAnnotations = map[string]string{
			"reviewedResourceRequestsAndLimits": "true",
		}
                podNamespace := pod.Namespace

		// Get pods not-ready in the namespace
                pods, err := clientset.CoreV1().Pods(podNamespace).List(context.TODO(), metav1.ListOptions{})
                if err != nil {
                        fmt.Errorf("Error listing pods")
                        panic(err)
                }
                var nonRunningPods []string
                for _, listedPod := range pods.Items {
                        podStatus := listedPod.Status.Phase
                        if (podStatus != "Running") {
                                nonRunningPods = append(nonRunningPods, listedPod.Name)
                        }
                }

		// If there is any non-valid container then reject the creation
		if (len(nonRunningPods) > 0 ) {
			resp.Allowed = false
			nonRunningPodNames := strings.Join(nonRunningPods, ", ")
			statusMessage := "The following non-running pods prevented the pod creation: " + nonRunningPodNames
			log.Print(statusMessage)
			resp.Result = &metav1.Status{
				Message: statusMessage,
				Status: "Failure",
			}
		} else {
			// Success
			statusMessage := "The pod is valid, pod creation can proceed"
			log.Print(statusMessage)
			resp.Result = &metav1.Status{
				Message: statusMessage,
				Status: "Success",
			}
		}


		admReview.Response = &resp
		// back into JSON so we can return the finished AdmissionReview w/ Response directly
		// w/o needing to convert things in the http handler
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err // untested section
		}
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody))
	}

	return responseBody, nil
}
