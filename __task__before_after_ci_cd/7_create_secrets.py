# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 7_create_secrets.py

import subprocess
import tempfile
import os

def run_command(command):
    try:
        # Run the command and capture both stdout and stderr
        result = subprocess.run(command, shell=True, check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # If the command is successful, it will return with a zero exit code
        if result.returncode == 0:
            return result.stdout.strip() or "Command executed successfully."  # If stdout is empty, return a success message
        else:
            return f"Command failed with return code {result.returncode}"
    except subprocess.CalledProcessError as e:
        # Handle error case and display stderr if command fails
        print(f"Error occurred: {e.stderr.strip()}")
        return None

def create_secrets():
    secrets = [
        {"name": "postgres-user", "value": "postgres"},
        {"name": "postgres-password", "value": "password"},
        {"name": "postgres-db", "value": "users"},
        {"name": "mongo-user", "value": "admin"},
        {"name": "mongo-password", "value": "password"},
        {"name": "mailer-username", "value": "johnsmith"},
        {"name": "mailer-password", "value": "password"},
        {"name": "user-email", "value": "admin@example.com"},
        {"name": "user-password", "value": "$2a$12$1zGLuYDDNvATh4RA4avbKuheAMpb1svexSzrQm7up.bnpwQHs0jNe"},
        {"name": "pgadmin-user", "value": "admin@example.com"},
        {"name": "pgadmin-password", "value": "admin"},
    ]

    for secret in secrets:
        print(f"Checking if secret exists: {secret['name']}...")

        # Check if the secret exists
        check_command = f"gcloud secrets describe {secret['name']}"
        check_result = run_command(check_command)
        
        if check_result and "Not Found" not in check_result:
            print(f"Secret '{secret['name']}' already exists, updating with new value.")
            # Delete the existing secret
            delete_command = f"gcloud secrets delete {secret['name']} --quiet"
            delete_result = run_command(delete_command)
            if delete_result:
                print(f"Secret '{secret['name']}' deleted successfully.")
            else:
                print(f"Failed to delete secret '{secret['name']}'.")
        
        # Create or recreate the secret with the new value
        print(f"Creating secret: {secret['name']}...")

        # Write the secret value to a temporary file
        with tempfile.NamedTemporaryFile(delete=False) as temp_file:
            temp_file.write(secret['value'].encode('utf-8'))
            temp_file_path = temp_file.name

        try:
            # Try creating the secret
            command = f"gcloud secrets create {secret['name']} --replication-policy='automatic' --data-file={temp_file_path}"
            result = run_command(command)
            if result:
                print(f"Secret '{secret['name']}' created successfully.")
            else:
                print(f"Failed to create secret '{secret['name']}'.")

        finally:
            # Clean up the temporary file
            if os.path.exists(temp_file_path):
                os.remove(temp_file_path)

if __name__ == "__main__":
    create_secrets()
