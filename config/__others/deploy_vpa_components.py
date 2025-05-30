import subprocess

def run_kubectl_command(command):
    """Runs a kubectl command and handles errors."""
    try:
        print(f"Executing: {command}")
        subprocess.run(command, check=True, shell=True)
        print("Success!")
    except subprocess.CalledProcessError as e:
        print(f"Error executing {command}: {e}")


def deploy_vpa_components():
    """Deploys Kubernetes Vertical Pod Autoscaler (VPA) components."""
    vpa_manifests = [
        "https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml",
        "https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/vpa-rbac.yaml",
        "https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/updater-deployment.yaml",
        "https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/recommender-deployment.yaml",
        "https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/admission-controller-deployment.yaml",
    ]
    
    for manifest in vpa_manifests:
        run_kubectl_command(f"kubectl apply -f {manifest}")

if __name__ == "__main__":
    deploy_vpa_components()
