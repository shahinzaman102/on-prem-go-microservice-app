# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 -m venv cloud-firewall_rules-env
# source cloud-firewall_rules-env/bin/activate
# pip3 install google-cloud-compute
# gcloud auth application-default login
# python3 5_create_vpc_firewall_rules.py
# deactivate

from google.cloud import compute_v1
from google.api_core.extended_operation import ExtendedOperation

def wait_for_extended_operation(operation: ExtendedOperation, verbose_name: str = "operation", timeout: int = 300):
    """Wait for the extended operation to complete."""
    result = operation.result(timeout=timeout)

    if operation.error_code:
        print(f"Error during {verbose_name}: {operation.error_code}: {operation.error_message}")
        raise operation.exception() or RuntimeError(operation.error_message)

    if operation.warnings:
        print(f"Warnings during {verbose_name}:")
        for warning in operation.warnings:
            print(f" - {warning.code}: {warning.message}")

    return result

def create_firewall_rule(
    project_id: str,
    rule_name: str,
    network: str,
    direction: str,
    priority: int,
    source_ranges: list,
    target_tags: list,
    allowed_protocols_ports: list,
):
    """
    Creates a firewall rule with the specified parameters.
    Args:
        project_id: The project ID or number for the Cloud project.
        rule_name: The name of the firewall rule to create.
        network: URL of the network for the firewall rule.
        direction: Direction of traffic to which this firewall applies (INGRESS or EGRESS).
        priority: The priority of the rule (lower values take precedence).
        source_ranges: A list of source IP ranges in CIDR format.
        target_tags: A list of tags for the rule.
        allowed_protocols_ports: A list of tuples, each containing protocol and a list of ports.
    """
    firewall_rule = compute_v1.Firewall()
    firewall_rule.name = rule_name
    firewall_rule.direction = direction
    firewall_rule.priority = priority
    firewall_rule.network = network
    firewall_rule.source_ranges = source_ranges
    firewall_rule.target_tags = target_tags

    # Construct the list of Allowed rules
    allowed_list = []
    for proto, ports in allowed_protocols_ports:
        allowed = compute_v1.Allowed()
        allowed.I_p_protocol = proto  # Correct field name from the documentation
        allowed.ports = ports if ports else []  # Only include ports if provided
        allowed_list.append(allowed)
    firewall_rule.allowed = allowed_list

    firewall_client = compute_v1.FirewallsClient()

    try:
        operation = firewall_client.insert(project=project_id, firewall_resource=firewall_rule)
        wait_for_extended_operation(operation, f"create firewall rule {rule_name}")
        print(f"Firewall rule '{rule_name}' created successfully.")
    except Exception as e:
        if "already exists" in str(e):
            print(f"Firewall rule '{rule_name}' already exists. Skipping...")
        else:
            print(f"Error creating firewall rule '{rule_name}': {e}")
            raise

if __name__ == "__main__":
    # Define project, network, and subnetwork details
    project_id = "go-microservice-app-449402"  # Replace with your project ID
    network_name = "custom-vpc-network"  # Replace with your network name
    subnetwork_name = "custom-subnetwork"  # Replace with your subnetwork name
    subnet_cidr = "10.0.0.0/16"  # Subnet CIDR block

    # Construct network and subnetwork URLs
    network_url = f"projects/{project_id}/global/networks/{network_name}"
    subnetwork_url = f"projects/{project_id}/regions/us-central1/subnetworks/{subnetwork_name}"  # Modify region as needed

    # Custom Default Firewall rules -->
    # --------------------------------------------------------------
    create_firewall_rule(
        project_id=project_id,
        rule_name="custom-allow-icmp",
        network=network_url,
        direction="INGRESS",
        priority=65534,
        source_ranges=["0.0.0.0/0"],  # Open to all
        target_tags=["custom"],
        allowed_protocols_ports=[("icmp", [])],  # Allow ICMP
    )

    create_firewall_rule(
        project_id=project_id,
        rule_name="custom-allow-internal",
        network=network_url,
        direction="INGRESS",
        priority=65534,
        source_ranges=["10.128.0.0/9"],  # Internal CIDR range
        target_tags=["custom"],
        allowed_protocols_ports=[("tcp", ["0-65535"]), ("udp", ["0-65535"]), ("icmp", [])],  # Allow all TCP, UDP, and ICMP
    )

    create_firewall_rule(
        project_id=project_id,
        rule_name="custom-allow-rdp",
        network=network_url,
        direction="INGRESS",
        priority=65534,
        source_ranges=["0.0.0.0/0"],  # Open to all
        target_tags=["custom"],
        allowed_protocols_ports=[("tcp", ["3389"])],  # Allow RDP on port 3389
    )

    create_firewall_rule(
        project_id=project_id,
        rule_name="custom-allow-ssh",
        network=network_url,
        direction="INGRESS",
        priority=65534,
        source_ranges=["0.0.0.0/0"],  # Open to all
        target_tags=["custom"],
        allowed_protocols_ports=[("tcp", ["22"])],  # Allow SSH on port 22
    )

    # Firewall rule for Master to worker traffic -->
    # --------------------------------------------------------------
    create_firewall_rule(
        project_id=project_id,
        rule_name="allow-master-to-worker-8443",
        network=network_url,
        direction="INGRESS",
        priority=1000,
        source_ranges=["10.0.0.0/16"],  # Replace with master CIDR block
        target_tags=["no-external-ip"],
        allowed_protocols_ports=[("tcp", ["8443"])],  # Allow TCP traffic on port 8443
    )
