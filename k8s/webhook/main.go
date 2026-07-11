package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateResources(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
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
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		log.Printf("could not unmarshal pod: %v", err)
		http.Error(w, "could not unmarshal pod", http.StatusBadRequest)
		return
	}

	allowed := true
	message := "All containers have resource requests specified."

	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests.Cpu().IsZero() || container.Resources.Requests.Memory().IsZero() {
			allowed = false
			message = fmt.Sprintf("Container %s is missing CPU or Memory requests", container.Name)
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
	// The TLS certificate paths are hardcoded and injected via Kubernetes secrets
	err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
