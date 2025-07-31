# Run the script using Python3 --> 

# cd py_file_path
# ex "py_file_path" --> 
# cd py_file_path : /home/shahinzaman/Documents/on-prem-go-microservice-app/__task__before_after_ci_cd

# python3 4_deploy_istio.py

import subprocess
import os
import shutil

HELM_VERSION = "v3.16.2"
HELM_ARCHIVE = f"helm-{HELM_VERSION}-linux-amd64.tar.gz"
HELM_URL = f"https://get.helm.sh/{HELM_ARCHIVE}"
HELM_BINARY_PATH = "/usr/local/bin/helm"

def run(cmd, check=True):
    print(f"👉 Running: {cmd}")
    subprocess.run(cmd, shell=True, check=check)

def helm_installed():
    try:
        subprocess.run(["helm", "version"], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        print("✅ Helm is already installed.")
        return True
    except subprocess.CalledProcessError:
        return False

def install_helm():
    print("🚀 Installing Helm...")
    run(f"curl -O {HELM_URL}")
    run(f"tar xvf {HELM_ARCHIVE}")
    run(f"sudo mv linux-amd64/helm {HELM_BINARY_PATH}")
    run(f"rm {HELM_ARCHIVE}")
    shutil.rmtree("linux-amd64")
    run("helm version")

def deploy_istio():
    print("🚀 Deploying Istio with automatic tracing...")
    run("helm repo add istio https://istio-release.storage.googleapis.com/charts")
    run("helm repo update")

    run("kubectl create namespace istio-system || true")

    run("helm install istio-base istio/base -n istio-system")
    run("helm install istiod istio/istiod -n istio-system "
        "--set global.proxy.autoInject=enabled "
        "--set meshConfig.enableTracing=true "
        "--set meshConfig.defaultConfig.tracing.sampling=100.0")

    run("helm install istio-ingress istio/gateway -n istio-system")

def label_namespace(namespace):
    print(f"🏷️ Labeling namespace '{namespace}' for sidecar injection...")
    run(f"kubectl label namespace {namespace} istio-injection=enabled --overwrite")

def restart_deployments(namespace):
    print(f"🔄 Restarting deployments in namespace '{namespace}' to inject sidecars...")
    run(f"kubectl delete pods --all -n {namespace}")

def main():
    if not helm_installed():
        install_helm()

    deploy_istio()

    # 👇 Replace 'default' with your actual app namespace if different
    app_namespace = "default"
    label_namespace(app_namespace)
    restart_deployments(app_namespace)

    print("✅ Istio has been deployed with infrastructure-level tracing!")

if __name__ == "__main__":
    main()
