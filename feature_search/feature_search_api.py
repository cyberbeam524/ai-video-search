from flask import Flask, request, jsonify
import torch
from torchvision import models, transforms
from PIL import Image
from annoy import AnnoyIndex
import numpy as np
from collections import OrderedDict
from decimal import Decimal

app = Flask(__name__)

device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
model = models.resnet50(pretrained=True).to(device)
model = torch.nn.Sequential(*(list(model.children())[:-1]))  # Remove the final classification layer
model.eval()

def search_image_across_videos(query_image_path, annoy_index_path, timestamps_path, model, device, num_neighbors=50):
    """Search for an image across all videos using Annoy index and return the top 20 distinct video results."""
    dim = 2048  # Dimensionality of the ResNet50 features
    annoy_index = AnnoyIndex(dim, 'angular')
    annoy_index.load(annoy_index_path)

    # Load the timestamps (video_path, timestamp) tuples
    timestamps = np.load(timestamps_path, allow_pickle=True)

    # Open and preprocess the query image
    query_image = Image.open(query_image_path).convert('RGB')
    transform = transforms.Compose([
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
    ])
    img_tensor = transform(query_image).unsqueeze(0).to(device)

    # Extract features from the query image
    with torch.no_grad():
        features = model(img_tensor).cpu().numpy().flatten()

    # Get the top num_neighbors results from the Annoy index
    best_match_indices, distances = annoy_index.get_nns_by_vector(features, num_neighbors, include_distances=True)

    # Create a list of (video_path, timestamp, distance) tuples
    results = []
    for idx, dist in zip(best_match_indices, distances):
        video_path, timestamp = timestamps[idx]
        results.append((video_path, timestamp, dist))

    # Sort the results by distance (i.e., confidence), lowest to highest
    results.sort(key=lambda x: x[2])

    # Use an ordered dictionary to maintain distinct videos, preserving order
    distinct_videos = OrderedDict()
    for video_path, timestamp, dist in results:
        if video_path not in distinct_videos:
            distinct_videos[video_path] = (timestamp, dist)
        if len(distinct_videos) == 20:  # Limit to the top 20 distinct videos
            break

    # Convert the dictionary to a list of top 20 results
    top_20_results = [(video_path, timestamp, dist) for video_path, (timestamp, dist) in distinct_videos.items()]

    return top_20_results

@app.route('/search', methods=['POST'])
def search_image():
    image_path = request.json['image_path']

    # Get the top 20 distinct results
    top_20_results = search_image_across_videos(image_path, 'video_features.ann', 'timestamps.npy', model, device)

    # Format the response
    response = [
        {
            'video_path': video_path,
            'timestamp': float(Decimal(timestamp)),
            'distance': distance
        }
        for video_path, timestamp, distance in top_20_results
    ]

    return jsonify(response)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
