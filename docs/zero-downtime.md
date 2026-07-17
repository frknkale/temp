# Zero-Downtime Deployment Approach

This document explains the strategy for achieving zero-downtime deployments for the Java application (Step 2.3.c).

## Strategy: Rolling Updates

Kubernetes supports zero-downtime deployments out of the box using **RollingUpdate** strategies for Deployments.

### Key Components

1. **`maxUnavailable: 0` and `maxSurge: 1`:**
   In the Deployment specification (`k8s/app-manifests/deployment.yaml`), the update strategy is explicitly configured:
   ```yaml
   strategy:
     type: RollingUpdate
     rollingUpdate:
       maxSurge: 1        # Allow 1 extra Pod during rollout
       maxUnavailable: 0  # All existing Pods must remain running
   ```
   - `maxSurge: 1`: Kubernetes spins up **one** extra Pod above the desired replica count before terminating any old Pods.
   - `maxUnavailable: 0`: **Zero** Pods are taken out of service during the rollout, guaranteeing no capacity loss.

2. **Readiness Probes:**
   The `readinessProbe` targets the Spring Boot Actuator endpoint:
   ```yaml
   readinessProbe:
     httpGet:
       path: /actuator/health/readiness
       port: 8080
     initialDelaySeconds: 10
     periodSeconds: 5
   ```
   - **Why this is critical:** Kubernetes will not add the new Pod to the Service endpoints (and thus Nginx will not route traffic to it) until the readiness probe passes. This ensures traffic is only sent to fully-started Pods.

3. **Liveness Probes:**
   ```yaml
   livenessProbe:
     httpGet:
       path: /actuator/health/liveness
       port: 8080
     initialDelaySeconds: 20
     periodSeconds: 15
   ```
   - Ensures Pods that enter a broken state post-startup are automatically restarted by the kubelet.

4. **Graceful Shutdown (SIGTERM):**
   When Kubernetes terminates old Pods, it sends a `SIGTERM` signal. The Spring Boot application handles this gracefully:
   ```properties
   # application.properties
   server.shutdown=graceful
   spring.lifecycle.timeout-per-shutdown-phase=30s
   ```
   Combined with `terminationGracePeriodSeconds: 30` in the Pod spec, this gives in-flight requests up to 30 seconds to complete before the Pod is killed.

5. **Topology Spread Constraints:**
   ```yaml
   topologySpreadConstraints:
   - maxSkew: 1
     topologyKey: kubernetes.io/hostname
     whenUnsatisfiable: DoNotSchedule
   ```
   - Ensures Pods are evenly distributed across worker nodes, so a node failure doesn't take down all replicas.

### Rollout Process
During deployment, Kubernetes will:
1. Create a new ReplicaSet with the updated image tag.
2. Spin up 1 new Pod (due to `maxSurge: 1`).
3. Wait for the `readinessProbe` of the new Pod to pass.
4. Add the new Pod to the Service endpoints.
5. Send `SIGTERM` to 1 old Pod (since `maxUnavailable: 0`, this only happens after the new Pod is Ready).
6. Repeat steps 2-5 until all 4 Pods are running the new version.

This ensures the user never experiences a dropped request or a 502 Bad Gateway error during the deployment pipeline.

### CI/CD Integration
The deploy pipeline (`Jenkinsfile.deploy`) uses Ansible to apply the manifests and then waits for the rollout to complete:
```bash
kubectl rollout status deployment/sample-java-app --namespace default --timeout=300s
```
If the rollout fails (e.g., the new image crashes), Kubernetes automatically stops the rollout, preserving the old Pods.
