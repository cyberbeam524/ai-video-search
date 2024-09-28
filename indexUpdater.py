from redis import Redis
import numpy as np
import json
from annoy import AnnoyIndex
import os
import time

# Initialize Redis client
redis_client = Redis(host='localhost', port=6379, db=0)

# Rebuild Annoy index function
def rebuild_annoy_index(features_list, num_trees=10, annoy_index_path="video_features.ann"):
    dim = 2048  # Expected dimension for ResNet50 features

    # Initialize Annoy index
    annoy_index = AnnoyIndex(dim, 'angular')

    # Add all features to the index
    for idx, vector in enumerate(features_list):
        annoy_index.add_item(idx, vector)

    annoy_index.build(num_trees)
    annoy_index.save(annoy_index_path)
    print(f"Annoy index saved at {annoy_index_path}")


# Function to append new timestamps to the timestamps.npy file
def append_timestamps(new_timestamps, timestamps_path="timestamps.npy"):
    if os.path.exists(timestamps_path):
        # Load existing timestamps and append new ones
        existing_timestamps = np.load(timestamps_path, allow_pickle=True)
        combined_timestamps = np.concatenate([existing_timestamps, new_timestamps], axis=0)
    else:
        # No existing timestamps, use new ones directly
        combined_timestamps = new_timestamps

    # Save the updated timestamps
    np.save(timestamps_path, combined_timestamps)
    print(f"Timestamps updated and saved to {timestamps_path}")


# Function to listen to Redis queue and update Annoy index and timestamps
def listen_to_queue_and_update_index():
    features_list = []
    timestamps_list = []

    while True:
        # Block until a new item is received
        _, feature_data = redis_client.blpop("features_queue")

        # Deserialize feature data from JSON
        feature_data = json.loads(feature_data)
        video_path = feature_data["video_path"]
        new_features = np.array(feature_data["features"])
        new_timestamps = feature_data["timestamps"]

        print(f"Received features and timestamps for video: {video_path}")

        # Add the new features and timestamps to the lists for the index and timestamps file
        features_list.append(new_features)
        timestamps_list.extend(new_timestamps)

        # Rebuild the Annoy index
        print("Rebuilding Annoy index with new features...")
        rebuild_annoy_index(np.concatenate(features_list))

        # Append the new timestamps to the timestamps.npy file
        append_timestamps(timestamps_list)

        # Optionally, save the combined features to disk for future use
        np.save("combined_features.npy", np.concatenate(features_list))



# Start listening to the queue
if __name__ == "__main__":
    listen_to_queue_and_update_index()
