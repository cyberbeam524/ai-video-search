import grpc
# from proto import feature_extractor_pb2
# from proto import feature_extractor_pb2_grpc
from proto import features_pb2
from proto import features_pb2_grpc
from concurrent import futures
import torch
from torch.nn.parallel import DistributedDataParallel as DDP
import torch.distributed as dist
from torchvision import models, transforms
from PIL import Image
import cv2
import os
import numpy as np
import torch.multiprocessing as mp
from redis import Redis
import json
import subprocess
import glob
from annoy import AnnoyIndex

# Initialize process group
def init_process_group(rank, world_size):
    dist.init_process_group(backend='gloo', init_method='file:///tmp/sharedfile', world_size=world_size, rank=rank)
    
    if rank < torch.cuda.device_count():
        print(f"Using device: {torch.cuda.get_device_name(rank)}")
    else:
        print(f"Invalid device ID for rank {rank}, using CPU instead.")

# Initialize Redis client
def push_features_to_queue(features, timestamps, video_path):
    redis_client = Redis(host='localhost', port=6379, db=0)  # Ensure separate connection per process
    feature_data = {
        "video_path": video_path,
        "features": features.tolist(),  # Convert numpy array to list
        "timestamps": timestamps  # Store the timestamps directly
    }
    redis_client.rpush("features_queue", json.dumps(feature_data))
    print(f"Pushed features and timestamps for {video_path} to Redis queue.")

# Main function for processing and pushing features to Redis
def process_and_push_features(video_path):
    chunk_paths = split_video_into_chunks(video_path)
    world_size = torch.cuda.device_count() or 1  # Default to CPU if no GPU

    distributed_chunks = distribute_chunks(chunk_paths, world_size)
    features, timestamps = parallel_process_chunks(distributed_chunks, world_size)

    combined_features, combined_timestamps = combine_features_and_timestamps(video_path, features, timestamps)

    # Push the combined features and timestamps to the Redis queue for processing
    push_features_to_queue(combined_features, combined_timestamps, video_path)

    return "success"

# Distribute video chunks across GPUs
def distribute_chunks(chunk_paths, num_gpus):
    """Distribute video chunks evenly across available GPUs."""
    chunk_batches = [[] for _ in range(num_gpus)]
    for i, chunk_path in enumerate(chunk_paths):
        chunk_batches[i % num_gpus].append(chunk_path)
    return chunk_batches

# Process video chunks in parallel across GPUs
def parallel_process_chunks(chunk_batches, world_size):
    """Process chunks in parallel on multiple GPUs and collect results in a shared list."""
    print(f"Processing {len(chunk_batches)} chunk batches in parallel.")

    # Create a multiprocessing Manager for sharing the list between processes
    with mp.Manager() as manager:
        shared_features = manager.list()  # A shared list for storing the features
        shared_timestamps = manager.list()  # A shared list for storing the timestamps
        
        # Start parallel processing using torch.multiprocessing.spawn
        torch.multiprocessing.spawn(process_chunks_on_gpu, args=(chunk_batches, world_size, shared_features, shared_timestamps), nprocs=world_size, join=True)
        
        # After processing, return the shared list (which will have the features and timestamps)
        return list(shared_features), list(shared_timestamps)

# Initialize the model with DDP
def initialize_model(rank, model):
    model = torch.nn.Sequential(*(list(model.children())[:-1]))
    device = setup_device(rank)
    model.to(device)
    model = DDP(model, device_ids=[rank] if torch.cuda.is_available() else None)
    model.eval()  # Ensure model is in evaluation mode
    print(f"Model initialized on device: {device}")
    return model, device

# Process chunks on a specific GPU (or CPU)
def process_chunks_on_gpu(rank, chunk_batches, world_size, shared_features, shared_timestamps):
    """Process a batch of chunks on a specific GPU (or CPU if no GPU)."""
    init_process_group(rank, world_size)
    print(f"Processing on GPU {rank}...")

    device = setup_device(rank)
    model = models.resnet50(pretrained=True)
    model, device = initialize_model(rank, model)

    for chunk_path in chunk_batches[rank]:
        print(f"Processing chunk: {chunk_path} on GPU {rank}")
        if os.path.exists(chunk_path):
            features, timestamps = extract_chunk_features(chunk_path, model, device)
            if features is not None and timestamps is not None:
                # Store the features and timestamps in the shared lists
                shared_features.append(features)
                shared_timestamps.append(timestamps)
            else:
                print(f"Skipping chunk {chunk_path} due to extraction failure.")
        else:
            print(f"Skipping invalid chunk {chunk_path}")

    print(f"GPU {rank} finished processing.")

# Setup device for distributed training
def setup_device(rank):
    if torch.cuda.is_available():
        torch.cuda.set_device(rank)
        return torch.device(f'cuda:{rank}')
    else:
        return torch.device('cpu')

# Extract features ensuring 2048-dimensional vectors
def extract_features(image, model, device):
    transform = transforms.Compose([
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
    ])
    
    img_tensor = transform(image).unsqueeze(0).to(device)  # Add batch dimension and move to device
    
    with torch.no_grad():
        features = model(img_tensor)  # Extract features
        features = features.view(-1)  # Flatten the features to a 1D tensor
    
    # print(f"features.size(0): {features.size(0)}")
    assert features.size(0) == 2048, f"Error: Expected feature size of 2048, got {features.size(0)}"
    
    return features.cpu().numpy()

# Split video into chunks using FFmpeg
def split_video_into_chunks(video_path, chunk_duration=60):
    print(f"Splitting video into {chunk_duration}-second chunks...")
    output_pattern = f"{video_path}_chunk_%03d.mp4"
    cmd = ['ffmpeg', '-i', video_path, '-c', 'copy', '-map', '0', '-segment_time', str(chunk_duration), '-f', 'segment', output_pattern]
    result = subprocess.run(cmd, capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"Error chunking video: {result.stderr}")
        return []
    
    chunk_paths = sorted(glob.glob(f"{video_path}_chunk_*.mp4"))
    print(f"Video split into {len(chunk_paths)} chunks.")
    return chunk_paths

# Extract features from video chunks
def extract_chunk_features(chunk_path, model, device):
    print(f"Extracting features from chunk: {chunk_path}")
    
    cap = cv2.VideoCapture(chunk_path)
    if not cap.isOpened():
        print(f"Failed to open video file: {chunk_path}")
        return None, None

    frame_count = 0
    features_list = []
    timestamps = []

    while True:
        ret, frame = cap.read()
        if not ret:
            break  # End of video or failure to read frame

        timestamp = cap.get(cv2.CAP_PROP_POS_MSEC) / 1000.0  # Get timestamp in seconds
        timestamps.append(timestamp)

        img = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
        pil_img = Image.fromarray(img)

        # Extract features using the ResNet50 model
        try:
            features = extract_features(pil_img, model, device)  # Use extract_features function
            features_list.append(features)
        except AssertionError as e:
            print(f"Error during feature extraction: {e}")
            return None, None

        frame_count += 1
        if frame_count >= 5:  # Limit to the first 5 frames
            break

    cap.release()

    if not features_list:
        print(f"No valid frames extracted from {chunk_path}")
        return None, None

    combined_features = np.stack(features_list, axis=0)
    return combined_features, timestamps

# Combine features and timestamps and store them as (video_path, timestamp) tuples
def combine_features_and_timestamps(video_path, features_list, timestamps_list):
    print(f"Combining features and timestamps from {len(features_list)} chunks for video {video_path}...")

    valid_features = [features for features in features_list if features is not None]
    valid_timestamps = [timestamps for timestamps in timestamps_list if timestamps is not None]

    if len(valid_features) == 0:
        print("No valid features to combine.")
        return None, None

    # Combine features and timestamps
    combined_features = np.concatenate(valid_features, axis=0)
    combined_timestamps = [(video_path, ts) for timestamps in valid_timestamps for ts in timestamps]

    print(f"Combined features and timestamps saved for video {video_path}.")
    return combined_features, combined_timestamps

# gRPC service
class FeatureExtractorServicer(features_pb2_grpc.FeatureExtractorServicer):
    def ExtractFeatures(self, request, context):
        video_path = request.video_path
        print(f"Received request to extract features from video: {video_path}")

        try:
            # Process video and push features to Redis
            result = process_and_push_features(video_path)

            return features_pb2.FeatureResponse(status=result)

        except Exception as e:
            print(f"Error during GPU-based feature extraction: {str(e)}")
            return features_pb2.FeatureResponse(status="failure")


# FeatureSearcher Servicer class
class FeatureSearcherServicer(features_pb2_grpc.FeatureSearcherServicer):
    def __init__(self):
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        self.model = models.resnet50(pretrained=True).to(self.device)
        self.model = torch.nn.Sequential(*(list(self.model.children())[:-1]))  # Remove the final layer
        self.model.eval()

    def extract_features(self, image_path):
        image = Image.open(image_path).convert('RGB')
        transform = transforms.Compose([
            transforms.Resize(256),
            transforms.CenterCrop(224),
            transforms.ToTensor(),
        ])
        img_tensor = transform(image).unsqueeze(0).to(self.device)
        with torch.no_grad():
            features = self.model(img_tensor).cpu().numpy().flatten()
        return features

    def search_image_across_videos(self, request, context):
        image_path = request.image_path
        print(f"Received request for image: {image_path}")

        # Feature extraction
        query_features = self.extract_features(image_path)

        # Annoy index and timestamp loading
        dim = 2048
        annoy_index = AnnoyIndex(dim, 'angular')
        annoy_index.load('video_features.ann')

        timestamps = np.load('timestamps.npy', allow_pickle=True)

        # Annoy search
        best_match_idx, distances = annoy_index.get_nns_by_vector(query_features, 1, include_distances=True)

        best_match_idx = best_match_idx[0]
        best_match_timestamp = timestamps[best_match_idx][1]
        video_path = timestamps[best_match_idx][0]
        distance = distances[0]
        print(f"vvv: {video_path}")
        return features_pb2.SearchImageResponse(video_path=video_path, timestamp=best_match_timestamp, distance=distance)



# Main function to start the gRPC service
def serve(rank, world_size):
    init_process_group(rank, world_size)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    features_pb2_grpc.add_FeatureExtractorServicer_to_server(FeatureExtractorServicer(), server)
    features_pb2_grpc.add_FeatureSearcherServicer_to_server(FeatureSearcherServicer(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    print(f"gRPC server running on rank {rank}...")
    server.wait_for_termination()

def main(rank, world_size):
    serve(rank, world_size)

if __name__ == '__main__':
    world_size = torch.cuda.device_count() or 1  # Fallback to CPU if no GPU
    torch.multiprocessing.spawn(main, args=(world_size,), nprocs=world_size, join=True)
