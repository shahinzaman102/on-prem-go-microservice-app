# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 2_configure_wif.py

import subprocess

# Function to run gcloud commands
def run_gcloud_command(command):
    """
    Runs a gcloud command and prints the output or error.
    """
    try:
        print(f"Running command: {' '.join(command)}")
        result = subprocess.run(command, capture_output=True, text=True, check=True)
        print("Command output:\n", result.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Error occurred: {e.stderr}")

# Assign roles to the service account
def assign_roles_to_service_account(service_account_email, project_id):
    """
    Assigns necessary roles to a service account.
    """
    commands = [
        ["gcloud", "iam", "service-accounts", "add-iam-policy-binding", 
         service_account_email, "--role", "roles/iam.serviceAccountTokenCreator", 
         "--member", f"serviceAccount:{service_account_email}"]
    ]
    
    for command in commands:
        run_gcloud_command(command)

# Create a Workload Identity Pool
def create_workload_identity_pool(pool_name):
    """
    Creates a Workload Identity Pool.
    """
    command = [
        "gcloud", "iam", "workload-identity-pools", "create", pool_name,
        "--location", "global", "--display-name", f"{pool_name} Pool"
    ]
    run_gcloud_command(command)

# Create a Workload Identity Provider for GitLab OIDC
def create_workload_identity_provider(pool_name, provider_name, issuer_uri, attribute_condition, allowed_audiences):
    """
    Creates a Workload Identity Provider for GitLab OIDC.
    """
    command = [
        "gcloud", "iam", "workload-identity-pools", "providers", "create-oidc", provider_name,
        "--location", "global", "--workload-identity-pool", pool_name,
        "--issuer-uri", issuer_uri, "--display-name", f"{provider_name} OIDC Provider",
        "--attribute-mapping", "attribute.guest_access=assertion.guest_access,attribute.planner_access=assertion.planner_access,attribute.reporter_access=assertion.reporter_access,attribute.developer_access=assertion.developer_access,attribute.maintainer_access=assertion.maintainer_access,attribute.owner_access=assertion.owner_access,attribute.namespace_id=assertion.namespace_id,attribute.namespace_path=assertion.namespace_path,attribute.project_id=assertion.project_id,attribute.project_path=assertion.project_path,attribute.user_id=assertion.user_id,attribute.user_login=assertion.user_login,attribute.user_email=assertion.user_email,attribute.user_access_level=assertion.user_access_level,google.subject=assertion.sub",
        "--attribute-condition", attribute_condition,
        "--allowed-audiences", allowed_audiences
    ]
    run_gcloud_command(command)

# Map GitLab CI Identity to GCP Service Account
def map_gitlab_identity_to_gcp_service_account(service_account_email, pool_name, project_path, branch_name, project_number):
    """
    Maps a GitLab CI identity to the GCP service account.
    """
    command = [
        "gcloud", "iam", "service-accounts", "add-iam-policy-binding", 
        service_account_email, "--role", "roles/iam.serviceAccountTokenCreator",
        "--member", f"principal://iam.googleapis.com/projects/{project_number}/locations/global/workloadIdentityPools/{pool_name}/subject/project_path:{project_path}:ref_type:branch:ref:{branch_name}"
    ]
    run_gcloud_command(command)

# Main function to run all tasks
def main():
    # Replace with your project configuration
    project_id = "go-microservice-app-449402"
    project_number = "501649434946"
    service_account_name = "terraform-service-account"
    service_account_email = f"{service_account_name}@{project_id}.iam.gserviceaccount.com"

    # Workload Identity Pool and Provider settings
    pool_name = "gitlab-pool-1" # change the name as necessary
    provider_name = "gitlab-provider-1" # change the name as necessary
    issuer_uri = "https://gitlab.com"
    allowed_audiences = "https://gitlab.com"

    # GitLab-specific settings
    namespace_id = "64250114"
    project_path = "shahinzaman102/go-microservice-app" # change the name as necessary
    branch_name = "main" # change the name as necessary
    attribute_condition = f"assertion.namespace_id=='{namespace_id}' && assertion.project_path=='{project_path}'"

    # Assign roles to the service account
    assign_roles_to_service_account(service_account_email, project_id)
    
    # Create the Workload Identity Pool
    create_workload_identity_pool(pool_name)
    
    # Create the Workload Identity Provider
    create_workload_identity_provider(pool_name, provider_name, issuer_uri, attribute_condition, allowed_audiences)
    
    # Map GitLab CI Identity to GCP Service Account
    map_gitlab_identity_to_gcp_service_account(service_account_email, pool_name, project_path, branch_name, project_number)

if __name__ == "__main__":
    main()
