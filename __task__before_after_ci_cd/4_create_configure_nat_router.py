# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 4_create_configure_nat_router.py

import subprocess

def create_nat_router_and_config(project_id, region, network_name, router_name, nat_name):
    try:
        # Step 1: Create the NAT router
        print(f"Creating NAT router '{router_name}'...")
        create_router_command = [
            "gcloud", "compute", "routers", "create", router_name,
            "--project", project_id,
            "--region", region,
            "--network", network_name
        ]
        subprocess.run(create_router_command, check=True)
        print(f"Router '{router_name}' created successfully.")

        # Step 2: Configure NAT
        print(f"Configuring NAT '{nat_name}'...")
        create_nat_command = [
            "gcloud", "compute", "routers", "nats", "create", nat_name,
            "--project", project_id,
            "--router", router_name,
            "--region", region,
            "--auto-allocate-nat-external-ips",
            "--nat-all-subnet-ip-ranges",
            "--enable-logging",
            "--log-filter", "ERRORS_ONLY"
        ]
        subprocess.run(create_nat_command, check=True)
        print(f"NAT '{nat_name}' configured successfully.")

    except subprocess.CalledProcessError as e:
        print(f"Error occurred while creating NAT router or configuration: {e}")

# Variables extracted from terraform.tfvars
PROJECT_ID = "go-microservice-app-449402"
REGION = "us-central1"
NETWORK_NAME = "custom-vpc-network"
ROUTER_NAME = "nat-router" # change the name as necessary
NAT_NAME = "nat-config" # change the name as necessary

# Run the function
create_nat_router_and_config(PROJECT_ID, REGION, NETWORK_NAME, ROUTER_NAME, NAT_NAME)
