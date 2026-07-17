package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// validateResources validates that all containers in a Deployment's pod spec
// have CPU and memory resource requests defined.
// Assignment Step 2.4: "Write a custom validation webhook that fails if a
// deployment does not specify required CPU and memory resource requests."
func validateResources(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	var admissionReviewReq admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)
		return
	}

	req := admissionReviewReq.Request

	// Unmarshal the Deployment object
	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		log.Printf("could not unmarshal deployment: %v", err)
		http.Error(w, "could not unmarshal deployment", http.StatusBadRequest)
		return
	}

	allowed := true
	message := "All containers have CPU and memory resource requests specified."

	// Check each container in the pod template spec
	for _, container := range deployment.Spec.Template.Spec.Containers {
		cpuReq := container.Resources.Requests.Cpu()
		memReq := container.Resources.Requests.Memory()

		if cpuReq == nil || cpuReq.IsZero() || memReq == nil || memReq.IsZero() {
			allowed = false
			message = fmt.Sprintf(
				"Deployment '%s': container '%s' is missing required CPU or Memory resource requests",
				deployment.Name, container.Name,
			)
			break
		}
	}

	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}

	admissionReviewRes := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: admissionResponse,
	}

	respBytes, err := json.Marshal(admissionReviewRes)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func main() {
	http.HandleFunc("/validate", validateResources)

	log.Println("Starting validation webhook server on port 8443...")
	// TLS certificates are injected via cert-manager and mounted as a Kubernetes secret
	err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
