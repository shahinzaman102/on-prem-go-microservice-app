# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 1_enable_apis_accounts_roles.py

import subprocess

def create_service_account(service_account_name, project_id):
    """
    Creates a service account in a specific project.

    Args:
        service_account_name (str): Name of the service account to create.
        project_id (str): GCP project ID.
    """
    command = [
        "gcloud", "iam", "service-accounts", "create", service_account_name,
        "--project", project_id,
        "--display-name", service_account_name.replace("-", " ").title()  # Just for a better display name
    ]
    try:
        print(f"Creating service account {service_account_name}...")
        subprocess.run(command, check=True)
        print(f"Successfully created service account {service_account_name}.")
    except subprocess.CalledProcessError as e:
        print(f"Failed to create service account {service_account_name}. Error: {e}")

def assign_roles(service_account_email, roles, project_id):
    """
    Assigns roles to a service account in a specific project.

    Args:
        service_account_email (str): Email of the service account.
        roles (list): List of roles to assign.
        project_id (str): GCP project ID.
    """
    for role in roles:
        command = [
            "gcloud", "projects", "add-iam-policy-binding", project_id,
            "--member", f"serviceAccount:{service_account_email}",
            "--role", role
        ]
        try:
            print(f"Assigning role {role} to {service_account_email}...")
            subprocess.run(command, check=True)
            print(f"Successfully assigned role {role}.")
        except subprocess.CalledProcessError as e:
            print(f"Failed to assign role {role}. Error: {e}")

def enable_apis(project_id, apis):
    """
    Enables the specified APIs for the project.

    Args:
        project_id (str): GCP project ID.
        apis (list): List of API names to enable.
    """
    for api in apis:
        command = [
            "gcloud", "services", "enable", api,
            "--project", project_id
        ]
        try:
            print(f"Enabling API {api}...")
            subprocess.run(command, check=True)
            print(f"Successfully enabled API {api}.")
        except subprocess.CalledProcessError as e:
            print(f"Failed to enable API {api}. Error: {e}")

def main():
    # Replace with your project ID
    project_id = "go-microservice-app-449402"

    # Service account names
    terraform_service_account_name = "terraform-service-account"
    logging_monitoring_account_name = "logging-monitoring-service-acc"

    # Full email addresses for service accounts
    terraform_service_account = f"{terraform_service_account_name}@{project_id}.iam.gserviceaccount.com"
    logging_monitoring_account = f"{logging_monitoring_account_name}@{project_id}.iam.gserviceaccount.com"

    # Roles for terraform-service-account
    terraform_roles = [
        "roles/resourcemanager.projectIamAdmin",
        "roles/serviceusage.serviceUsageAdmin",
        "roles/iam.serviceAccountTokenCreator",
        "roles/iam.serviceAccountUser",
        "roles/iam.workloadIdentityPoolAdmin",
        "roles/monitoring.viewer",
        "roles/storage.admin",
        "roles/iam.serviceAccountAdmin",
        "roles/secretmanager.secretAccessor",
        "roles/container.admin",
        "roles/compute.admin",
        "roles/logging.viewer",
        "roles/monitoring.alertPolicyEditor",
        "roles/logging.configWriter",
        "roles/artifactregistry.reader",
        "roles/artifactregistry.writer",
    ]

    # Roles for logging-monitoring-account
    logging_monitoring_roles = [
        "roles/logging.logWriter",
        "roles/logging.viewer",
        "roles/monitoring.metricWriter",
    ]

    # APIs to enable
    required_apis = [
        "cloudresourcemanager.googleapis.com",
        "container.googleapis.com",
        "logging.googleapis.com",
        "monitoring.googleapis.com",
        "iam.googleapis.com",
        "iamcredentials.googleapis.com",
        "recommender.googleapis.com",
    ]

    # Enable required APIs
    enable_apis(project_id, required_apis)

    # Create service accounts
    create_service_account(terraform_service_account_name, project_id)
    create_service_account(logging_monitoring_account_name, project_id)

    # Assign roles to service accounts
    assign_roles(terraform_service_account, terraform_roles, project_id)
    assign_roles(logging_monitoring_account, logging_monitoring_roles, project_id)

if __name__ == "__main__":
    main()
