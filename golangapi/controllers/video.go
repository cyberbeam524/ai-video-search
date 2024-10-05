package controllers

import (
	"bytes"
	"context"
	"encoding/json"

	// "encoding/json"
	"fmt"
	// "io"
	// "io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	// "video_search_project/config"
	// "video_search_project/models"

	"github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "video_search_project/proto"
	"video_search_project/proto"
	pb "video_search_project/proto" // Generated gRPC code

	"github.com/go-redis/redis/v8" // Redis for real-time notifications
	"google.golang.org/grpc"
	"gopkg.in/gomail.v2" // Email package
)

var redisClient *redis.Client

// gRPC Client for communicating with Python service
func startGRPCClient(videoPath string) (string, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewFeatureExtractorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the ExtractFeatures RPC from the gRPC server
	request := &pb.FeatureRequest{VideoPath: videoPath}
	resp, err := client.ExtractFeatures(ctx, request)
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

// Calculate video size using the FileHeader, not the File interface
func calculateVideoSize(file *multipart.FileHeader) int64 {
	return file.Size
}

// Split video into chunks using FFmpeg
func splitVideoIntoChunks(videoPath string, chunkDuration int) ([]string, error) {
	// Use FFmpeg to split the video into chunks
	outputPattern := videoPath + "_chunk_%03d.mp4"
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-c", "copy", "-map", "0", "-segment_time", fmt.Sprintf("00:%02d:00", chunkDuration), "-f", "segment", outputPattern)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Find all chunks created
	matches, err := filepath.Glob(videoPath + "_chunk_*.mp4")
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// Process video locally (CPU-based)
func processLocally(videoPath string) error {
	fmt.Printf("Processing %s locally (CPU)...\n", videoPath)

	// Extract frames from the video
	err := extractFrames(videoPath)
	if err != nil {
		return fmt.Errorf("failed to extract frames: %v", err)
	}

	// Call Python script for feature extraction on CPU
	err = runCPUPythonScript(videoPath)
	if err != nil {
		return fmt.Errorf("failed to extract features on CPU: %v", err)
	}

	fmt.Printf("Finished processing %s on CPU\n", videoPath)
	return nil
}

func extractFrames(videoPath string) error {
	// Ensure the output directory exists
	if err := os.MkdirAll("./frames", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create frames directory: %v", err)
	}

	// Check if the input video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file not found at path: %s", videoPath)
	}

	// Execute the ffmpeg command to extract frames
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1", "./frames/frame_%04d.jpg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error extracting frames: %v, output: %s", err, string(output))
	}

	fmt.Println("Frames extracted successfully.")
	return nil
}

// Run Python script for feature extraction on CPU
func runCPUPythonScript(videoPath string) error {
	// Call the Python script that processes frames and extracts features using CPU
	cmd := exec.Command("python3", "cpu_feature_extractor.py", videoPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running CPU feature extraction: %v", err)
	}
	return nil
}

// Upload video handler with improved error handling and notifications
func UploadHandler(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file is required"})
		return
	}

	// Save video locally
	videoPath := "./videos/" + file.Filename
	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video"})
		return
	}

	// Calculate the video size
	videoSize := calculateVideoSize(file)

	// Set a workload threshold (e.g., 500MB)
	// workloadThreshold := int64(500 * 1024 * 1024) // 500MB
	workloadThreshold := int64(500) // 500MB

	if videoSize < workloadThreshold {
		// Process on CPU
		err := processLocally(videoPath)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process video locally"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "Processed on CPU"})
	} else {
		// Split video into chunks and process using GPU
		chunks, err := splitVideoIntoChunks(videoPath, 10) // Split into 10-minute chunks
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to split video"})
			return
		}

		// Process each chunk on GPU
		for _, chunk := range chunks {
			go func(chunkPath string) {
				status, err := startGRPCClient(chunkPath)
				if err != nil {
					log.Println("Failed to extract features using GPU for chunk", chunkPath)
					notifyUser("user@example.com", fmt.Sprintf("Error processing video chunk: %s", chunkPath))
					return
				}
				log.Println("Processed chunk", chunkPath, "with status:", status)
				notifyUser("user@example.com", "Feature extraction completed successfully.")
			}(chunk)
		}

		c.JSON(http.StatusOK, gin.H{"status": "Processing on GPU in parallel"})
	}
}



func UploadImageAndSearchHandler(c *gin.Context) {
    // Step 1: Handle image file upload
    file, err := c.FormFile("image") // Expecting the image file in form field "image"
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image file"})
        return
    }

    // Step 2: Save the uploaded image file to a temporary location
    imagePath := "./uploaded_images/" + file.Filename
    if err := c.SaveUploadedFile(file, imagePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
        return
    }

    // Step 3: Call the gRPC Python service to search for the image in the video index
    videoPath, timestamp, distance, err := searchImageAcrossVideosGRPC(imagePath)
    if err != nil {
		fmt.Println(err);
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search video for the image"})
        return
    }

    // Step 4: Return the result back to the client
    c.JSON(http.StatusOK, gin.H{
        "video_path": videoPath,
        "timestamp":  timestamp,
        "distance":   distance,
    })
}

// searchImageAcrossVideosGRPC calls the Python gRPC server to search for the image across videos
func searchImageAcrossVideosGRPC(imagePath string) (string, float32, float32, error) {
    // Step 5: Dial the Python gRPC server
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        return "", 0, 0, fmt.Errorf("failed to connect to gRPC server: %v", err)
    }
    defer conn.Close()

    client := pb.NewFeatureSearcherClient(conn)

    // Step 6: Create the request and call the SearchImageAcrossVideos method
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()

    req := &pb.SearchImageRequest{ImagePath: imagePath}
    res, err := client.SearchImageAcrossVideos(ctx, req)
    if err != nil {
        return "", 0, 0, fmt.Errorf("error calling SearchImageAcrossVideos: %v", err)
    }

    return res.VideoPath, res.Timestamp, res.Distance, nil
}


// Struct to hold the response from Flask API
type SearchResponse struct {
    VideoPath string  `json:"video_path"`
    Timestamp float64 `json:"timestamp"`
    Distance  float64 `json:"distance"`
}

// Define a slice to hold the list of search results
type SearchResults []SearchResponse

// Handler function to call the Flask API
func CallFlaskAPI(c *gin.Context) {
	// Step 1: Handle image file upload
	file, err := c.FormFile("image") // Expecting the image file in form field "image"
	if err != nil {
		log.Println("Error uploading file:", err) // Log the error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image file"})
		return
	}

	// Step 2: Save the uploaded image file to a temporary location
	imagePath := "./uploaded_images/" + file.Filename
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		return
	}

    // Prepare the request body as JSON
    requestBody, err := json.Marshal(map[string]string{
        "image_path": imagePath,
    })
    if err != nil {
        log.Fatalf("Error encoding JSON: %v", err)
    }

    // Make a POST request to the Flask API
    url := "http://127.0.0.1:5000:/search" // Flask server URL
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        log.Fatalf("Error sending request: %v", err)
    }
    defer resp.Body.Close()
	// Decode the response from Flask into a list of search results
	var searchResults SearchResults
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	// Send the list of results back to the client (Go API)
	c.JSON(http.StatusOK, gin.H{
		"results": searchResults,
	})
}


// RunFeatureExtractionHandler is the new handler that mimics the Python run() functionality
func RunFeatureExtractionHandler(c *gin.Context) {
    // videoPath := c.Query("video_path") // Assuming the video path is passed as a query parameter
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file is required"})
		return
	}

	// Save video locally
	videoPath := "./videos/" + file.Filename
	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video"})
		return
	}

    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gRPC server", "details": err.Error()})
        return
    }
    defer conn.Close()

    // Create the gRPC client stub
    client := proto.NewFeatureExtractorClient(conn)

    // Set up context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Send a feature extraction request
    request := &proto.FeatureRequest{VideoPath: videoPath}
    response, err := client.ExtractFeatures(ctx, request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract features", "details": err.Error()})
        return
    }

    // Return the status from the gRPC server response
    c.JSON(http.StatusOK, gin.H{"status": response.Status})
}


// Function to send email notification
func notifyUser(email, message string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@example.com")
	email = "maaruni505@gmail.com"
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Feature Extraction Completed")
	m.SetBody("text/plain", message)

	d := gomail.NewDialer("smtp.example.com", 587, "your-username", "your-password")
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Failed to send email:", err)
	}
}

func checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg not found: %v", err)
	}
	fmt.Printf("FFmpeg version: %s\n", string(output))
	return nil
}

// Initialize Redis for WebSocket-style notification
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}
