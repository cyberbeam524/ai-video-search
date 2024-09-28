"use client"

import { useState } from "react";
import Image from 'next/image';
export default function SearchPage() {
    const baseUrl = "http://localhost:8080/videos/";
    const [showModal, setShowModal] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [isUploadComplete, setIsUploadComplete] = useState(false);
  const [selectedImage, setSelectedImage] = useState(null);
  const [selectedImageBlob, setSelectedImageBlob] = useState(null);
  const [results, setResults] = useState([
    {
        "distance": 0.1790960431098938,
        "timestamp": 0.0,
        "video_path": "./videos/MRT East-West line disruption_ Expert shares insights into what could have happened - CNA (360p, h264, youtube).mp4"
    },
    {
        "distance": 0.2687002122402191,
        "timestamp": 0.16683333333333333,
        "video_path": "./videos/video2.mp4"
    }
]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);


  const [sortOption, setSortOption] = useState("Price low to high");

  const handleImageChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      const imageFile = e.target.files[0];
      setSelectedImage(URL.createObjectURL(imageFile));
      setSelectedImageBlob(imageFile);
    }
  };

  const handleSliderChange = (e, videoElement, result) => {
    const time = e.target.value;
    videoElement.currentTime = time;
  };

const handleSearch = async (e) => {
  e.preventDefault();

  if (!selectedImage) {
    console.log('Please select an image to search.');
    return;
  }

  // Retrieve JWT token from localStorage
  const token = localStorage.getItem('token');

  if (!token) {
    console.log('You must be logged in to perform this action.', token);
    return;
  }

  // Check if selectedImage is correctly set
  console.log("Selected Image:", selectedImageBlob);

  // Create form data to send the image to the Go API
  const formData = new FormData();
  formData.append('image', selectedImageBlob);

  setIsLoading(true);
  setError(null);

  try {
    // Call the Go API
    const response = await fetch('http://localhost:8080/search', {
      method: 'POST',
      body: formData,
      headers: {
        'Authorization': `Bearer ${token}`,  // Add the JWT token to the request header
        // Do NOT manually set the 'Content-Type' header, fetch will handle this automatically
      },
    });

    if (!response.ok) {
      // Check for a response error and throw
      const errorText = await response.text();
      console.error("Response error:", errorText);
      throw new Error('Failed to search for videos.');
    }

    const data = await response.json();
    console.log("data.results", data.results);
    setResults(data.results);
  } catch (err) {
    console.error("Search error:", err.message);
    setError(err.message);
  } finally {
    setIsLoading(false);
  }
};


  const handleSortChange = (e) => {
    setSortOption(e.target.value);
    // Implement sorting logic based on selected value
  };

    // Show the modal for uploading a video
    const handleUploadVideoClick = () => {
        setShowModal(true);
      };
    
      // Handle the file upload process with progress
      const handleFileUpload = (e) => {
        const file = e.target.files[0];
        const formData = new FormData();
        formData.append("video", file);
    
        const xhr = new XMLHttpRequest();
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) {
            const percentage = Math.floor((event.loaded / event.total) * 100);
            setUploadProgress(percentage);
          }
        };
    
        xhr.onload = () => {
          setIsUploadComplete(true);
          setTimeout(() => {
            setShowModal(false);
            setUploadProgress(0);
            setIsUploadComplete(false);
          }, 2000);
        };
    
        // xhr.open("POST", "http://localhost:8080/video/upload2", true);
        // xhr.send(formData);

          // Add JWT token from localStorage or sessionStorage
        const token = localStorage.getItem("token"); // Or use sessionStorage
        if (token) {
            xhr.open("POST", "http://localhost:8080/video/upload2", true);
            xhr.setRequestHeader("Authorization", `Bearer ${token}`);  // Add token to request header
            xhr.send(formData);
        } else {
            console.error("No token found. Please log in.");
        }

      };

  return (
    <div className="min-h-screen flex">

<aside className="w-64 bg-white drop-shadow-2xl">
        <div className="p-6">
          <Image className="rounded-full shadow-lg" src="/img/image3.jpg" alt="Logo" width={100} height={100} />
          <ul className="mt-8">
            <li className="mb-6">
              <a href="/billing" className="text-gray-700 hover:text-blue-600">Billing</a>
            </li>
            <li className="mb-6">
              <a href="/videos" className="text-gray-700 hover:text-blue-600">Uploaded Videos</a>
            </li>
            <li className="mb-6">
              <a href="/account" className="text-gray-700 hover:text-blue-600">Account</a>
            </li>

            <li className="mb-6">
              <button onClick={handleUploadVideoClick} className="text-white bold hover:text-gray-600 bg-blue-900 p-5 shadow-lg rounded-lg mt-4">Upload Video</button>
            </li>
          </ul>
        </div>
      </aside>


      {/* Modal for uploading video */}
      {showModal && (
        <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-1/3">
            <h2 className="text-xl font-bold mb-4">Upload Video</h2>
            <input type="file" accept="video/*" onChange={handleFileUpload} className="mb-4" />
            <div className="w-full bg-gray-200 rounded-full h-4 mb-4">
              <div className={`h-4 rounded-full ${uploadProgress === 100 ? 'bg-green-500' : 'bg-blue-500'}`} style={{ width: `${uploadProgress}%` }}></div>
            </div>
            {isUploadComplete ? (
              <p className="text-green-600 font-bold">Upload Complete! ✅</p>
            ) : (
              <p>Uploading... {uploadProgress}%</p>
            )}
          </div>
        </div>
      )}


      {/* Side Panel */}
      <aside className="w-1/4 p-6 border-r bg-gray-50 shadow-lg">
        <h2 className="text-xl font-bold mb-4">Filters</h2>

        {/* Image Preview */}
        {selectedImage && (
          <div className="mb-4">
            <h3 className="text-lg mb-2">Uploaded Image</h3>
            <img src={selectedImage} alt="Selected" className="w-full" />
          </div>
        )}

        {/* Filters */}
        <div className="mb-4">
          <label className="block mb-2 font-medium">Price</label>
          <input type="range" className="w-full" />
        </div>

        <div className="mb-4">
          <label className="block mb-2 font-medium">Currency</label>
          <select className="w-full p-2 border rounded">
            <option>USD</option>
            <option>ETH</option>
            <option>BTC</option>
          </select>
        </div>

        <div className="mb-4">
          <label className="block mb-2 font-medium">Traits</label>
          <input type="text" placeholder="Search traits" className="w-full p-2 border rounded" />
        </div>

        <button className="w-full bg-blue-600 text-white py-2 mt-4 rounded" onClick={handleSearch}>
          Apply Filters
        </button>
      </aside>

      {/* Main Content */}
      {/* <main className="w-3/4 p-6 bg-gray-100"> */}
      <main className="flex-1 flex flex-col p-5 bg-gray-200">



        {/* Search Form */}
        <form onSubmit={handleSearch} className="mb-8 flex gap-4 items-center mt-10 ml-10">
          <input
            type="file"
            accept="image/*"
            onChange={handleImageChange}
            className="file-input file-input-bordered w-full max-w-xs"
          />
          <button type="submit" className="btn bg-orange-600 text-white">
            Search
          </button>
        </form>

        {/* Sorting */}
        <div className="flex justify-between mb-4">
          <div>Showing {results.length} results</div>
          <div>
            <label className="mr-2 font-medium">Sort by:</label>
            <select value={sortOption} onChange={handleSortChange} className="p-2 border rounded">
              <option value="Price low to high">Price low to high</option>
              <option value="Price high to low">Price high to low</option>
              <option value="Recently listed">Recently listed</option>
              <option value="Best offer">Best offer</option>
              <option value="Recently created">Recently created</option>
            </select>
          </div>
        </div>

        {/* Search Results */}
        <div className="grid grid-cols-3 gap-4">
          {results.length > 0 ? (
            results.map((result, index) => (
              <div key={index} className="p-4 border rounded-lg shadow-lg bg-white">

                {/* Video Preview */}
                <video
                  id={`video-${index}`}
                  width="100%"
                  controls
                  src={result.video_path.replace("./videos/", baseUrl)}
                  type="video/mp4"
                  className="mb-4"
                >
                  Your browser does not support the video tag.
                </video>

                {/* Timestamp Slider */}
                <label className="block mb-2 font-medium">
                  Timestamp: {result.timestamp}s
                </label>
                <input
                  type="range"
                  min="0"
                  max={Math.floor(result.video_duration)} // Assuming the API returns video duration
                  step="1"
                  defaultValue={Math.floor(result.timestamp)}
                  onChange={(e) => handleSliderChange(e, document.getElementById(`video-${index}`), result)}
                  className="w-full"
                />


                <p className="font-medium">Video: {result.video_path}</p>
                <p className="bold">Timestamp: {result.timestamp}</p>
                <p>Distance: {result.distance}</p>
              </div>
            ))
          ) : (
            <p>No results found</p>
          )}
        </div>
      </main>
    </div>
  );
}
