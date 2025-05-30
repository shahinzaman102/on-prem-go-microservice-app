# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 6_create_artifact_registry_repository.py

import subprocess

# Define the environment variables
GAR_REPO_NAME = "go-microservices-repo"
GCP_REGION = "us-central1"

# Create the gcloud command to create the Artifact Registry repository
gcloud_command = [
    "gcloud", "artifacts", "repositories", "create", GAR_REPO_NAME,
    "--repository-format=docker", "--location", GCP_REGION
]

# Run the command
try:
    print(f"Creating Artifact Registry repository: {GAR_REPO_NAME} in region: {GCP_REGION}")
    result = subprocess.run(gcloud_command, check=True, capture_output=True, text=True)
    print("Repository created successfully!")
    print(result.stdout)
except subprocess.CalledProcessError as e:
    print(f"Error while creating repository: {e}")
    print(e.stderr)
