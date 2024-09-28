'use client'
import { useState } from 'react'

export default function UploadModal({ isOpen, onClose }) {
  const [files, setFiles] = useState([])

  const handleUpload = (event) => {
    setFiles([...event.target.files])
  }

  return (
    <div className={`modal ${isOpen ? 'block' : 'hidden'}`}>
      <div className="modal-content">
        <h2>Upload Videos</h2>
        <input type="file" multiple onChange={handleUpload} />
        <button className="btn" onClick={onClose}>Close</button>
      </div>
    </div>
  )
}

