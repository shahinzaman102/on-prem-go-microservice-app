# Run the script using Python3 --> 
# ***(with the authenticated service account which has owner permission) --> 

# cd /mnt/c/Users/shahi/.gcp
# python3 -m venv cloud-monitoring-env
# source cloud-monitoring-env/bin/activate
# pip3 install google-cloud-monitoring
# gcloud auth application-default login
# python3 8_setup_notify_channel.py
# deactivate

from google.cloud import monitoring_v3

def get_existing_notification_channel(project_id, display_name):
    """
    Check if the notification channel already exists.

    Args:
        project_id (str): Google Cloud project ID.
        display_name (str): Display name for the notification channel.

    Returns:
        str: Name of the existing notification channel or None if not found.
    """
    client = monitoring_v3.NotificationChannelServiceClient()
    project_name = f"projects/{project_id}"

    # List all notification channels and search for the matching display name
    channels = client.list_notification_channels(name=project_name)
    for channel in channels:
        if channel.display_name == display_name:
            print(f"Notification channel '{display_name}' already exists: {channel.name}")
            return channel.name

    return None  # If no existing channel is found


def create_email_notification_channel(project_id, email, display_name):
    """
    Create an email notification channel in Google Cloud Monitoring.
    
    Args:
        project_id (str): Google Cloud project ID.
        email (str): Email address for notifications.
        display_name (str): Display name for the notification channel.
    
    Returns:
        str: Name of the created notification channel.
    """
    client = monitoring_v3.NotificationChannelServiceClient()
    project_name = f"projects/{project_id}"

    # Define the email notification channel
    notification_channel = monitoring_v3.NotificationChannel(
        type_="email",
        display_name=display_name,
        labels={"email_address": email},
    )

    # Create the notification channel
    response = client.create_notification_channel(
        request={"name": project_name, "notification_channel": notification_channel}
    )
    print(f"Notification channel created: {response.name}")
    return response.name

if __name__ == "__main__":
    # First notification channel details
    PROJECT_ID = "go-microservice-app-449402"
    EMAIL_1 = "shahinzaman102@gmail.com"  # change it as necessary
    DISPLAY_NAME_1 = "System Handling Alerts"  # change it as necessary

    # Second notification channel details
    EMAIL_2 = "www.zaman19710@gmail.com"  # change it as necessary
    DISPLAY_NAME_2 = "Event Handling Alerts"  # change it as necessary

    # Step 1: Check and create the first notification channel
    notification_channel_name_1 = get_existing_notification_channel(PROJECT_ID, DISPLAY_NAME_1)
    if not notification_channel_name_1:
        notification_channel_name_1 = create_email_notification_channel(PROJECT_ID, EMAIL_1, DISPLAY_NAME_1)

    # Step 2: Check and create the second notification channel
    notification_channel_name_2 = get_existing_notification_channel(PROJECT_ID, DISPLAY_NAME_2)
    if not notification_channel_name_2:
        notification_channel_name_2 = create_email_notification_channel(PROJECT_ID, EMAIL_2, DISPLAY_NAME_2)

    # Print final status for both channels
    if notification_channel_name_1:
        print(f"Notification channel '{notification_channel_name_1}' is ready to use.")
    else:
        print("No new notification channel was created for the first channel.")

    if notification_channel_name_2:
        print(f"Notification channel '{notification_channel_name_2}' is ready to use.")
    else:
        print("No new notification channel was created for the second channel.")
