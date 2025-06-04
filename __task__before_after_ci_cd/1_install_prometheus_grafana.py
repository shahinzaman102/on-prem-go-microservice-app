# Run the script using Python3 --> 

# cd py_file_path
# ex "py_file_path" --> 
# cd py_file_path : /home/shahinzaman/Documents/projects/on-prem-go-microservice-app/__task__before_after_ci_cd

# python3 10_install_prometheus_grafana.py

import subprocess

def run_command(command: list, description: str):
    """
    Executes a shell command and handles errors.
    Args:
        command: The command to execute as a list of strings.
        description: Description of the task being performed.
    """
    print(f"Starting: {description}...")
    try:
        result = subprocess.run(command, check=True, capture_output=True, text=True)
        print(result.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Error during {description}: {e.stderr}")
        raise e

def main():
    # Set `helm_chart_path` path as necessary --> 
    helm_chart_path = "/home/shahinzaman/Documents/projects/on-prem-go-microservice-app/config/helm-monitoring-chart"
    namespace = "monitoring"

    # Update Helm dependencies
    run_command(
        ["helm", "dependency", "update", helm_chart_path],
        "Updating Helm dependencies"
    )

    # Lint the Helm chart
    run_command(
        ["helm", "lint", helm_chart_path],
        "Linting Helm chart"
    )

    # Create Kubernetes namespace (ignore if already exists)
    try:
        run_command(
            ["kubectl", "create", "namespace", namespace],
            f"Creating Kubernetes namespace '{namespace}'"
        )
    except subprocess.CalledProcessError as e:
        if "already exists" in str(e.stderr):
            print(f"Namespace '{namespace}' already exists, continuing...")
        else:
            raise e

    # Install or upgrade Prometheus & Grafana with custom Helm chart
    run_command(
        [
            "helm", "upgrade", "--install", "monitoring", helm_chart_path,
            "-n", namespace
        ],
        "Installing or upgrading Prometheus & Grafana"
    )

    print("Prometheus and Grafana have been successfully installed/upgraded with the custom configuration.")

if __name__ == "__main__":
    main()
