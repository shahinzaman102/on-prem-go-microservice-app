# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 -m venv cloud-networking-env
# source cloud-networking-env/bin/activate
# pip3 install google-cloud-compute
# gcloud auth application-default login
# python3 3_create_vpc_network_subnetwork.py
# deactivate

import os
import time
from google.auth import default
from google.cloud import compute_v1

def create_network(project_id, network_name):
    networks_client = compute_v1.NetworksClient()
    try:
        print(f"Checking if network '{network_name}' exists...")
        networks_client.get(project=project_id, network=network_name)
        print(f"Network '{network_name}' already exists.")
    except Exception:
        print(f"Network '{network_name}' does not exist. Creating it...")
        network = compute_v1.Network(
            name=network_name,
            auto_create_subnetworks=False,
            description="Custom VPC for GKE"
        )
        operation = networks_client.insert(project=project_id, network_resource=network)
        print(f"Network creation initiated. Operation: {operation.name}")
        wait_for_global_operation(project_id, operation.name)

def create_subnetwork(project_id, region, network_name, subnetwork_name, cidr_range):
    subnetworks_client = compute_v1.SubnetworksClient()
    try:
        print(f"Checking if subnetwork '{subnetwork_name}' exists in region '{region}'...")
        subnetworks_client.get(project=project_id, region=region, subnetwork=subnetwork_name)
        print(f"Subnetwork '{subnetwork_name}' already exists.")
    except Exception:
        print(f"Subnetwork '{subnetwork_name}' does not exist. Creating it...")
        subnetwork = compute_v1.Subnetwork(
            name=subnetwork_name,
            ip_cidr_range=cidr_range,
            region=region,
            network=f"projects/{project_id}/global/networks/{network_name}"
        )
        operation = subnetworks_client.insert(project=project_id, region=region, subnetwork_resource=subnetwork)
        print(f"Subnetwork creation initiated. Operation: {operation.name}")
        wait_for_region_operation(project_id, region, operation.name)

def wait_for_global_operation(project_id, operation_name, timeout=300):
    operation_client = compute_v1.GlobalOperationsClient()
    start_time = time.time()
    while time.time() - start_time < timeout:
        result = operation_client.get(operation=operation_name, project=project_id)
        if result.status == compute_v1.Operation.Status.DONE:
            print("Global operation completed.")
            if result.error:
                raise Exception(f"Operation failed: {result.error}")
            return
        print("Waiting for global operation to complete...")
        time.sleep(10)
    raise TimeoutError("Timed out waiting for global operation to complete.")

def wait_for_region_operation(project_id, region, operation_name, timeout=300):
    operation_client = compute_v1.RegionOperationsClient()
    start_time = time.time()
    while time.time() - start_time < timeout:
        result = operation_client.get(operation=operation_name, project=project_id, region=region)
        if result.status == compute_v1.Operation.Status.DONE:
            print("Regional operation completed.")
            if result.error:
                raise Exception(f"Operation failed: {result.error}")
            return
        print("Waiting for regional operation to complete...")
        time.sleep(10)
    raise TimeoutError("Timed out waiting for regional operation to complete.")

if __name__ == "__main__":
    # Use Application Default Credentials (ADC)
    credentials, project = default()

    # Set these variables accordingly
    REGION = os.environ.get("REGION", "us-central1")
    NETWORK_NAME = os.environ.get("NETWORK_NAME", "custom-vpc-network")
    SUBNETWORK_NAME = os.environ.get("SUBNETWORK_NAME", "custom-subnetwork")
    CIDR_RANGE = os.environ.get("CIDR_RANGE", "10.0.0.0/16")

    # Create the network and subnetwork
    create_network(project, NETWORK_NAME)
    create_subnetwork(project, REGION, NETWORK_NAME, SUBNETWORK_NAME, CIDR_RANGE)
