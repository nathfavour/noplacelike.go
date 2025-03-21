package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// uiHome renders the main UI page
func (s *Server) uiHome(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, homeTemplate)
}

// adminPanel renders the admin UI
func (s *Server) adminPanel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, adminTemplate)
}

// HTML templates for UI components
const homeTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>noplacelike</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        /* Reset and base styles */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: system-ui, -apple-system, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.5;
        }

        /* Layout */
        .navbar {
            background: white;
            padding: 1rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 1rem;
        }

        .grid {
            display: grid;
            gap: 1rem;
            margin: 1rem 0;
        }

        @media (min-width: 768px) {
            .grid { grid-template-columns: 1fr 1fr; }
        }

        /* Cards */
        .card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }

        /* Form elements */
        .textarea, .upload-area input { width: 100%; }
        .textarea {
            height: 8rem;
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin: 0.5rem 0;
            font-family: inherit;
        }

        .button {
            background: #4444ff;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 1rem;
        }

        .button:hover {
            background: #3333dd;
        }

        /* File upload area */
        .upload-area {
            border: 2px dashed #ddd;
            border-radius: 4px;
            padding: 2rem;
            text-align: center;
        }

        /* File list */
        .file-list {
            margin-top: 1rem;
        }

        .file-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem 0;
            border-bottom: 1px solid #eee;
        }

        .file-item:last-child {
            border-bottom: none;
        }

        .link-button {
            color: #4444ff;
            text-decoration: none;
            cursor: pointer;
        }

        .link-button:hover {
            text-decoration: underline;
        }

        .scrollable { max-height: 300px; overflow-y: auto; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 0.5rem; border: 1px solid #ddd; text-align: left; }
    </style>
</head>
<body>
    <nav class="navbar">
        <div class="container">
            <h1 style="font-size: 1.5rem; font-weight: 600;">noplacelike</h1>
        </div>
    </nav>

    <main class="container">
        <div class="grid">
            <!-- Clipboard Card -->
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Clipboard Sharing</h3>
                <textarea id="clipboard" class="textarea" 
                        placeholder="Paste text here to share..."></textarea>
                <button onclick="shareClipboard()" class="button">
                    Share Clipboard
                </button>
            </div>

            <!-- File Sharing Card -->
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">File Sharing</h3>
                <div class="upload-area">
                    <input type="file" id="fileInput" style="display: none;" multiple onchange="uploadFiles()">
                    <button onclick="document.getElementById('fileInput').click()" 
                            class="button">
                        Select Files
                    </button>
                    <p style="margin-top: 0.5rem; color: #666;">
                        or drag and drop files here
                    </p>
                </div>
            </div>

            <!-- Server Clipboard Card -->
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Server Clipboard</h3>
                <div id="serverClipboard" class="textarea" style="overflow:auto; background:#f0f0f0;"></div>
                <button onclick="fetchServerClipboard()" class="button" style="margin-top:0.5rem;">Fetch Server Clipboard</button>
            </div>

            <!-- Audio Streaming Card -->
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Audio Streaming</h3>
                <!-- Streaming Source Selection Section with clear button -->
                <div style="margin-bottom:1rem; border:1px solid #ddd; border-radius:4px; padding:1rem;">
                    <h4>Streaming Source Selection</h4>
                    <!-- New text input for absolute directory path -->
                    <div style="margin-bottom:0.5rem;">
                        <input type="text" id="streamDirInput" placeholder="Enter absolute directory path" style="width:70%; padding:0.3rem;">
                        <button class="button" onclick="submitStreamDir()" style="margin-left:0.5rem;">Add Directory</button>
                    </div>
                    <button class="button" onclick="clearStreamingDirectories()">Clear Streaming Directories</button>
                    <div id="selectedDirs" style="margin-top:0.5rem; font-size:0.9rem; color:#555;"></div>
                </div>
                <!-- Existing audio player and audio files listing -->
                <audio id="audioStream" controls style="width:100%;"></audio>
                <div id="audioFiles" style="margin-top: 1rem;"></div>
            </div>
        </div>

        <!-- File List -->
        <div class="card">
            <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Shared Files</h3>
            <div id="fileList" class="file-list">
                <!-- Files will be listed here dynamically -->
            </div>
        </div>
    </main>

    <script>
        // Fetch and display files
        async function updateFileList() {
            try {
                const response = await fetch('/api/files');
                const data = await response.json();
                const fileList = document.getElementById('fileList');
                fileList.innerHTML = data.files.map(file => "<div class=\"file-item\"><span>" + file + "</span><button onclick=\"downloadFile('" + file + "')\" class=\"link-button\">Download</button></div>").join('');
            } catch (error) {
                console.error('Error updating file list:', error);
            }
        }

        // Share clipboard content
        async function shareClipboard() {
            const text = document.getElementById('clipboard').value;
            try {
                await fetch('/api/clipboard', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({text})
                });
                alert('Clipboard shared successfully!');
            } catch (error) {
                alert('Failed to share clipboard: ' + error.message);
            }
        }

        // Trigger share immediately when text is pasted
        document.getElementById('clipboard').addEventListener('paste', () => {
            // Delay slightly to capture pasted content
            setTimeout(shareClipboard, 50);
        });

        // Fetch server clipboard content
        async function fetchServerClipboard() {
            try {
                const response = await fetch('/api/clipboard');
                const data = await response.json();
                document.getElementById('serverClipboard').textContent = data.text || '';
            } catch (error) {
                alert('Failed to fetch server clipboard: ' + error.message);
            }
        }

        async function uploadFiles() {
            const input = document.getElementById('fileInput');
            const files = input.files;
            if (!files.length) return;
            for (let file of files) {
                const formData = new FormData();
                formData.append('file', file);
                try {
                    const res = await fetch('/api/files', {
                        method: 'POST',
                        body: formData
                    });
                    const result = await res.json();
                    if (res.ok) {
                        console.log('Uploaded:', result.filename);
                    } else {
                        alert(result.error || 'Upload failed');
                    }
                } catch (error) {
                    console.error('Upload error:', error);
                }
            }
            input.value = '';
            updateFileList();
        }

        function downloadFile(filename) {
            window.open('/api/files/' + filename, '_blank');
        }

        // Updated function to list files grouped by streaming directory
        async function fetchAudioFiles() {
            try {
                const res = await fetch('/stream/list');
                const data = await res.json();
                const container = document.getElementById('audioFiles');
                let html = '';
                // data.files is an object: {folder1: [files], folder2: [files], ...}
                for (const [dir, files] of Object.entries(data.files)) {
                    html += "<h5>Directory: " + dir + "</h5>";
                    if (files && files.length) {
                        html += "<table><tr><th>File</th><th>Action</th></tr>";
                        files.forEach(file => {
                            html += "<tr><td>" + file + "</td><td><button class=\"button\" onclick=\"streamAudio('" + file + "')\">Stream</button></td></tr>";
                        });
                        html += "</table>";
                    } else {
                        html += "<p>No files in this directory.</p>";
                    }
                }
                container.innerHTML = html;
            } catch (error) {
                console.error('Error fetching audio files:', error);
            }
        }

        // Set the audio player source to the streaming endpoint for the selected file.
        function streamAudio(fileName) {
            const audio = document.getElementById('audioStream');
            audio.src = '/stream/play?file=' + encodeURIComponent(fileName);
            audio.play();
        }

        // Submit directory from text input
        async function submitStreamDir() {
            const input = document.getElementById('streamDirInput');
            const dir = input.value.trim();
            if (!dir) return;
            await addDirectoryAPI(dir);
            input.value = '';
            updateSelectedDirs(dir, true);
            fetchAudioFiles();
        }

        // Show newly added directories in the UI
        function updateSelectedDirs(newDir, append) {
            const display = document.getElementById('selectedDirs');
            let current = display.textContent || "";
            if (append) {
                if (current === "" || current.includes("No streaming directories")) {
                    display.textContent = "Selected: " + newDir;
                } else {
                    display.textContent = current + ", " + newDir;
                }
            }
        }

        // Clear all streaming directories
        async function clearStreamingDirectories() {
            try {
                // Get current directories via admin GET endpoint:
                const res = await fetch('/admin/dirs');
                const data = await res.json();
                const dirs = data.dirs || [];
                // Delete each directory from the config
                for (let dir of dirs) {
                    await fetch('/admin/dirs', {
                        method: 'DELETE',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({dir})
                    });
                }
                document.getElementById('selectedDirs').textContent = "No streaming directories selected.";
                fetchAudioFiles(); // Refresh audio files list after clearing
            } catch (error) {
                console.error('Error clearing streaming directories:', error);
            }
        }

        // Add a directory via the admin API
        async function addDirectoryAPI(dir) {
            try {
                const res = await fetch('/admin/dirs', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({dir})
                });
                const data = await res.json();
                if(data.status !== 'success') {
                    console.error("Error adding directory: " + (data.error || 'Unknown error'));
                }
            } catch(e) {
                console.error(e);
            }
        }

        // Initialize
        updateFileList();
        fetchAudioFiles();
    </script>
</body>
</html>`

const adminTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>noplacelike Admin</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        /* Modern Admin UI Styles */
        :root {
            --primary: #4444ff;
            --bg-dark: #1a1a1a;
            --text-light: #ffffff;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: system-ui, -apple-system, sans-serif;
            background: #f5f5f5;
            color: #333;
        }

        .admin-header {
            background: var(--bg-dark);
            color: var(--text-light);
            padding: 1rem;
            position: fixed;
            width: 100%;
            top: 0;
            z-index: 100;
        }

        .main-content {
            margin-top: 60px;
            padding: 2rem;
        }

        .section {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .scroll-container {
            max-height: 300px;
            overflow-y: auto;
            border: 1px solid #eee;
            border-radius: 4px;
            padding: 1rem;
            margin: 1rem 0;
        }

        .dir-table {
            width: 100%;
            border-collapse: collapse;
        }

        .dir-table th, .dir-table td {
            padding: 0.75rem;
            text-align: left;
            border-bottom: 1px solid #eee;
        }

        .button {
            background: var(--primary);
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
        }

        .button:hover { opacity: 0.9; }

        .input-group {
            display: flex;
            gap: 0.5rem;
            margin: 1rem 0;
        }

        input[type="text"] {
            flex: 1;
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <header class="admin-header">
        <h1>noplacelike Server Administration</h1>
    </header>
    <main class="main-content">
        <section class="section">
            <h2>Audio Streaming Directories</h2>
            <div class="input-group">
                <input type="text" id="newDir" placeholder="Enter directory path">
                <button class="button" onclick="addDirectory()">Add Directory</button>
            </div>
            <div class="scroll-container">
                <table class="dir-table">
                    <thead>
                        <tr>
                            <th>Directory Path</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="dirList">
                        <!-- Directories will be listed here -->
                    </tbody>
                </table>
            </div>
        </section>
        <section class="section">
            <h2>Current Audio Files</h2>
            <div class="scroll-container" id="audioFilesList">
                <!-- Audio files will be listed here -->
            </div>
        </section>
    </main>

    <script>
        async function loadDirectories() {
            try {
                const res = await fetch('/admin/dirs');
                const data = await res.json();
                const tbody = document.getElementById('dirList');
                tbody.innerHTML = data.dirs.map(dir => "<tr><td>" + dir + "</td><td>" + checkDirStatus(dir) + "</td><td><button class=\"button\" onclick=\"removeDirectory('" + dir + "')\">Remove</button></td></tr>").join('');
            } catch (error) {
                console.error('Error loading directories:', error);
            }
        }

        function checkDirStatus(dir) {
            return 'Active'; // You can implement actual status checking
        }

        async function addDirectory() {
            const input = document.getElementById('newDir');
            const dir = input.value.trim();
            if (!dir) return;
            try {
                const res = await fetch('/admin/dirs', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({dir})
                });
                const data = await res.json();
                if (data.status === 'success') {
                    input.value = '';
                    loadDirectories();
                } else {
                    alert(data.error || 'Failed to add directory');
                }
            } catch (error) {
                alert('Error adding directory: ' + error.message);
            }
        }

        async function removeDirectory(dir) {
            try {
                const res = await fetch('/admin/dirs', {
                    method: 'DELETE',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({dir})
                });
                const data = await res.json();
                if (data.status === 'success') {
                    loadDirectories();
                } else {
                    alert(data.error || 'Failed to remove directory');
                }
            } catch (error) {
                alert('Error removing directory: ' + error.message);
            }
        }

        // Initialize
        loadDirectories();
    </script>
</body>
</html>`