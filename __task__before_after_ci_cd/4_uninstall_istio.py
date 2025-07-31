# Run the script using Python3 --> 

# cd py_file_path
# ex "py_file_path" --> 
# cd py_file_path : /home/shahinzaman/Documents/on-prem-go-microservice-app/__task__before_after_ci_cd

# python3 4_uninstall_istio.py

import subprocess

def run(cmd, check=True):
    print(f"👉 Running: {cmd}")
    subprocess.run(cmd, shell=True, check=check)

def uninstall_istio():
    print("🧹 Uninstalling Istio components...")

    # Uninstall Istio components
    run("helm uninstall istio-ingress -n istio-system || true")
    run("helm uninstall istiod -n istio-system || true")
    run("helm uninstall istio-base -n istio-system || true")

    # Delete Istio namespace
    run("kubectl delete namespace istio-system || true")

    # Optional: remove label from your app namespace
    app_namespace = "default"
    print(f"🏷️ Removing sidecar injection label from namespace '{app_namespace}'...")
    run(f"kubectl label namespace {app_namespace} istio-injection- || true")

    print("✅ Istio has been uninstalled.")

if __name__ == "__main__":
    uninstall_istio()
