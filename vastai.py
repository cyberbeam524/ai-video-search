import requests
import time

API_KEY = 'your_vast_ai_api_key'
INSTANCE_ID = 'your_instance_id'

def get_instance_status(instance_id):
    url = f"https://vast.ai/api/v0/instances/{instance_id}/"
    headers = {'Authorization': f"ApiKey {API_KEY}"}
    response = requests.get(url, headers=headers)
    return response.json()

def stop_instance(instance_id):
    url = f"https://vast.ai/api/v0/instances/{instance_id}/stop"
    headers = {'Authorization': f"ApiKey {API_KEY}"}
    response = requests.post(url, headers=headers)
    if response.status_code == 200:
        print(f"Instance {instance_id} stopped successfully.")
    else:
        print(f"Failed to stop instance {instance_id}: {response.text}")

def main():
    while True:
        status = get_instance_status(INSTANCE_ID)
        # Check for a condition to stop the instance, such as a low utilization rate
        if status['state'] == 'running' and should_stop_based_on_your_criteria(status):
            stop_instance(INSTANCE_ID)
            break
        time.sleep(600)  # Check every 10 minutes

def should_stop_based_on_your_criteria(status):
    # Define your logic here, e.g., stop after running for 2 hours
    return status['running_duration'] > 7200

if __name__ == "__main__":
    main()
