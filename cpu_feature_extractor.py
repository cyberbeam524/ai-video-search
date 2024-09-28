import os
import sys
import torch
from torchvision import models, transforms
from PIL import Image

# Define the transform for the input image
transform = transforms.Compose([
    transforms.Resize(256),
    transforms.CenterCrop(224),
    transforms.ToTensor()
])

# Load the pre-trained model (ResNet-50)
try:
    model = models.resnet50(pretrained=True)
    model.eval()  # Set the model to evaluation mode
    print("ResNet-50 model loaded successfully.")
except Exception as e:
    print(f"Error loading ResNet-50 model: {e}")
    sys.exit(1)

def extract_features(image_path):
    """Extract features from an image."""
    try:
        image = Image.open(image_path)
        image = transform(image).unsqueeze(0)  # Add batch dimension
        with torch.no_grad():
            features = model(image)
        return features
    except Exception as e:
        print(f"Error extracting features from {image_path}: {e}")
        sys.exit(1)

def process_frames(frame_dir):
    """Process each frame in the directory."""
    if not os.path.exists(frame_dir):
        print(f"Frames directory not found: {frame_dir}")
        sys.exit(1)

    if not os.listdir(frame_dir):
        print("No frames found in frames directory.")
        sys.exit(1)
    
    feature_dir = os.path.join(frame_dir, "features")
    os.makedirs(feature_dir, exist_ok=True)
    
    # Iterate over frames in the frame directory
    for frame in os.listdir(frame_dir):
        if frame.endswith(".jpg"):
            frame_path = os.path.join(frame_dir, frame)
            print(f"Processing {frame_path} on CPU...")
            
            # Extract features
            features = extract_features(frame_path)
            
            # Save the features to a file
            feature_path = os.path.join(feature_dir, f"{os.path.basename(frame)}.features")
            with open(feature_path, "w") as f:
                f.write(str(features.cpu().numpy()))
    
    print("Finished extracting features from all frames.")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 cpu_feature_extractor.py <video_path>")
        sys.exit(1)

    video_path = sys.argv[1]
    frame_dir = "./frames"
    
    # Process frames and extract features
    process_frames(frame_dir)
