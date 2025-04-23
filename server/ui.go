package server

import (
	"encoding/json"
	"net/http"
	"strings"

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

// uiHomeWithTab renders the main UI page and sets the initial tab
func (s *Server) uiHomeWithTab(c *gin.Context, tab string) {
	c.Header("Content-Type", "text/html")
	// Inject a JS variable to select the tab
	tabScript := `<script>window._initialTab = '` + tab + `';</script>`
	// Inject full config for client-side usage
	cfgJSON, _ := json.Marshal(s.config)
	configScript := `<script>window._config = ` + string(cfgJSON) + `;</script>`
	// Insert the script just before </head>
	html := homeTemplate
	headEnd := "</head>"
	if idx := strings.Index(html, headEnd); idx != -1 {
		html = html[:idx] + configScript + tabScript + html[idx:]
	}
	c.String(http.StatusOK, html)
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
        /* Horizontal sub-tabs for Files view */
        .horizontal-tabs { display: flex; justify-content: space-between; border-bottom: 1px solid #e0e0e0; margin-bottom: 1rem; }
        .tab-group { display: flex; gap: 0.5rem; }
        .tab-btn { background: none; border: none; padding: 0.5rem 1rem; font-size: 1rem; cursor: pointer; color: #4444ff; }
        .tab-btn.active { border-bottom: 2px solid #4444ff; color: #222244; }

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

        /* Sidebar styles */
        .sidebar {
            position: fixed;
            top: 0;
            left: 0;
            height: 100vh;
            width: 220px;
            background: linear-gradient(180deg, #4444ff 0%, #222244 100%);
            color: #fff;
            display: flex;
            flex-direction: column;
            z-index: 1000;
            box-shadow: 2px 0 12px rgba(44,44,100,0.08);
            border-right: 1px solid #333366;
        }
        .sidebar .logo {
            font-size: 1.5rem;
            font-weight: bold;
            padding: 1.5rem 1rem 1rem 1.5rem;
            border-bottom: 1px solid #333366;
            letter-spacing: 1px;
            background: rgba(255,255,255,0.04);
        }
        .sidebar .nav {
            flex: 1;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            padding: 2rem 0 1rem 0;
        }
        .sidebar .nav button {
            background: none;
            border: none;
            color: #fff;
            text-align: left;
            padding: 0.9rem 2rem;
            font-size: 1.08rem;
            cursor: pointer;
            border-radius: 8px 20px 20px 8px;
            transition: background 0.18s, color 0.18s;
            display: flex;
            align-items: center;
            gap: 0.9em;
            font-weight: 500;
        }
        .sidebar .nav button.active, .sidebar .nav button:hover {
            background: linear-gradient(90deg, #fff 0%, #e0e7ff 100%);
            color: #4444ff;
        }
        .sidebar .nav .icon {
            font-size: 1.3em;
            width: 1.7em;
            text-align: center;
        }
        .sidebar .spacer { flex: 1; }
        .sidebar .footer {
            padding: 1.2rem 1.5rem;
            font-size: 0.95rem;
            color: #b3b3ff;
            border-top: 1px solid #333366;
            background: rgba(255,255,255,0.03);
        }
        .main-with-sidebar { margin-left: 220px; padding: 2rem 1rem 1rem 1rem; }
        @media (max-width: 900px) {
            .sidebar { display: none; }
            .main-with-sidebar { margin-left: 0; }
        }
        /* Bottom bar for mobile */
        .bottombar {
            display: none;
            position: fixed;
            left: 0; right: 0; bottom: 0;
            background: linear-gradient(90deg, #4444ff 0%, #222244 100%);
            color: #fff;
            height: 60px;
            z-index: 1000;
            box-shadow: 0 -2px 12px rgba(44,44,100,0.10);
            border-top: 1px solid #333366;
            justify-content: space-around;
            align-items: center;
        }
        .bottombar .nav-btn {
            background: none;
            border: none;
            color: #fff;
            font-size: 1.15rem;
            flex: 1;
            height: 100%;
            cursor: pointer;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 0.2em;
            border-radius: 8px 8px 0 0;
            transition: background 0.18s, color 0.18s;
        }
        .bottombar .nav-btn.active, .bottombar .nav-btn:hover {
            background: #fff;
            color: #4444ff;
        }
        .bottombar .icon {
            font-size: 1.3em;
        }
        @media (max-width: 900px) {
            .bottombar { display: flex; }
        }

        /* File browser styles */
        .file-browser-list { list-style: none; padding: 0; margin: 1rem 0; }
        .file-browser-list li { display: flex; align-items: center; justify-content: space-between; padding: 0.75rem; border-bottom: 1px solid #eee; }
        .file-browser-list li:hover { background: #fafafa; }
        .file-button-group button { margin-left: 0.5rem; padding: 0.3rem 0.6rem; font-size: 0.9rem; border: none; border-radius: 4px; background: #4444ff; color: #fff; cursor: pointer; }
        .file-button-group button:hover { background: #3333dd; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="logo">noplacelike</div>
        <div class="nav">
            <button id="tab-home" onclick="showTab('home')"><span class="icon">🏠</span> Home</button>
            <button id="tab-files" onclick="showTab('files')"><span class="icon">📁</span> Files</button>
            <button id="tab-audio" onclick="showTab('audio')"><span class="icon">🎵</span> Audio</button>
            <button id="tab-others" onclick="showTab('others')"><span class="icon">✨</span> Others</button>
        </div>
        <div class="spacer"></div>
        <div class="footer">v0.1.0</div>
    </div>
    <div class="bottombar">
        <button id="tab-home-mobile" class="nav-btn" onclick="showTab('home')"><span class="icon">🏠</span><span style="font-size:0.85em;">Home</span></button>
        <button id="tab-files-mobile" class="nav-btn" onclick="showTab('files')"><span class="icon">📁</span><span style="font-size:0.85em;">Files</span></button>
        <button id="tab-audio-mobile" class="nav-btn" onclick="showTab('audio')"><span class="icon">🎵</span><span style="font-size:0.85em;">Audio</span></button>
        <button id="tab-others-mobile" class="nav-btn" onclick="showTab('others')"><span class="icon">✨</span><span style="font-size:0.85em;">Others</span></button>
    </div>
    <main class="main-with-sidebar">
        <div id="tab-content-home">
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Live Clipboard Sync</h3>
                <div style="display:flex;align-items:center;gap:1em;">
                    <label style="display:flex;align-items:center;gap:0.5em;">
                        <input type="checkbox" id="liveClipboardToggle" onchange="toggleLiveClipboard()">
                        <span>Enable live clipboard sync</span>
                    </label>
                    <span id="liveClipboardStatus" style="color:#4444ff;font-weight:500;">OFF</span>
                </div>
                <div style="margin-top:1em;color:#888;font-size:0.95em;">
                    When enabled, your clipboard will automatically sync with the server in real time.
                </div>
                <div id="clipboard-controls" style="margin-top:2em;">
                    <div class="card" style="margin-bottom:1em;">
                        <h4 style="margin-bottom:0.5em;">Clipboard Controls</h4>
                        <div style="display:flex;gap:0.5em;align-items:center;flex-wrap:wrap;">
                            <button class="button" onclick="sendClipboardToServer()">Send clipboard to others</button>
                            <button class="button" onclick="receiveClipboardFromServer()">Receive clipboard from server</button>
                            <span id="clipboardSyncStatus" style="color:#888;font-size:0.95em;"></span>
                        </div>
                        <div style="margin-top:1em;">
                            <textarea id="manualClipboardInput" class="textarea" placeholder="Type here to send to other devices (does not affect your clipboard unless live sync is ON)"></textarea>
                            <button class="button" style="margin-top:0.5em;" onclick="sendManualClipboard()">Send to others</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div id="tab-content-clipboard" style="display:none;">
            <!-- Clipboard Card -->
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Clipboard Sharing</h3>
                <textarea id="clipboard" class="textarea" placeholder="Paste text here to share..."></textarea>
                <button onclick="shareClipboard()" class="button">Share Clipboard</button>
            </div>
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Server Clipboard</h3>
                <div id="serverClipboard" class="textarea" style="overflow:auto; background:#f0f0f0;"></div>
                <button onclick="fetchServerClipboard()" class="button" style="margin-top:0.5rem;">Fetch Server Clipboard</button>
            </div>
        </div>
        <div id="tab-content-files" style="display:none;">
            <div class="horizontal-tabs">
                <div class="tab-group"><button id="subtab-manager" class="tab-btn" onclick="showFileSubTab('manager')">Manager</button></div>
                <div class="tab-group"><button id="subtab-sharing" class="tab-btn" onclick="showFileSubTab('sharing')">Sharing</button></div>
            </div>
            <div id="filesub-manager">
                <div class="file-browser">
                    <div class="path" id="file-browser-path"></div>
                    <ul id="file-browser-list" class="file-browser-list"></ul>
                    <div id="file-browser-content"></div>
                </div>
            </div>
            <div id="filesub-sharing" style="display:none;">
                <div class="card">
                    <h3>File Sharing</h3>
                    <div class="upload-area">
                        <input type="file" id="fileInput" style="display:none;" multiple onchange="uploadFiles()">
                        <button onclick="document.getElementById('fileInput').click()" class="button">Select Files</button>
                        <p style="color:#666;">or drag and drop files here</p>
                    </div>
                </div>
                <div class="card">
                    <h3>Shared Files</h3>
                    <div id="fileList" class="file-list"></div>
                </div>
            </div>
        </div>
        <div id="tab-content-audio" style="display:none;">
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Audio Streaming</h3>
                <audio id="audioStream" controls style="width:100%;"></audio>
                <div id="audioFiles" style="margin-top: 1rem;"></div>
            </div>
        </div>
        <div id="tab-content-others" style="display:none;">
            <div class="card">
                <h3 style="font-size: 1.2rem; margin-bottom: 1rem;">Other Features</h3>
                <p>More functionalities coming soon.</p>
            </div>
        </div>
    </main>
    <script>
        // Tab switching logic
        function showTab(tab) {
            ['home','clipboard','files','audio','others'].forEach(function(t) {
                var content = document.getElementById('tab-content-' + t);
                if (content) content.style.display = (t === tab) ? '' : 'none';
                var btn = document.getElementById('tab-' + t);
                if (btn) btn.classList.toggle('active', t === tab);
                var btnMobile = document.getElementById('tab-' + t + '-mobile');
                if (btnMobile) btnMobile.classList.toggle('active', t === tab);
            });
        }
        // Default tab
        if (window._initialTab) {
            showTab(window._initialTab);
        } else {
            showTab('home');
        }

        // Files sub-tab logic
        function showFileSubTab(tab) {
            ['manager','sharing'].forEach(function(t){
                document.getElementById('filesub-'+t).style.display = (t===tab?'':'none');
                document.getElementById('subtab-'+t).classList.toggle('active', t===tab);
            });
        }
        // Default to Manager view in Files
        // Only call when Files tab is active
        document.getElementById('tab-files').addEventListener('click', function(){ showFileSubTab('manager'); loadFileBrowser('/'); });
        document.getElementById('tab-files-mobile').addEventListener('click', function(){ showFileSubTab('manager'); loadFileBrowser('/'); });
        showFileSubTab('manager');

        // File browser logic
        var currentPath = '/';
        function loadFileBrowser(path) {
            if (!path) path = '/';
            currentPath = path;
            document.getElementById('file-browser-path').textContent = path;
            document.getElementById('file-browser-content').innerHTML = '';
            fetch('/api/v1/filesystem/list?path=' + encodeURIComponent(path))
                .then(function(res) { return res.json(); })
                .then(function(data) {
                    var ul = document.getElementById('file-browser-list');
                    ul.innerHTML = '';
                    if (path !== '/') {
                        var upLi = document.createElement('li');
                        upLi.innerHTML = '<span class="icon">⬆️</span> <button class="folder-link" onclick="loadFileBrowser(\'' + parentDir(path) + '\')">.. (Up)</button>';
                        ul.appendChild(upLi);
                    }
                    (data.directories || []).forEach(function(dir) {
                        var li = document.createElement('li');
                        li.innerHTML = '<span class="icon">📁</span> <button class="folder-link" onclick="loadFileBrowser(\'' + joinPath(path, dir) + '\')">' + dir + '</button>';
                        ul.appendChild(li);
                    });
                    (data.files || []).forEach(function(file) {
                        var li = document.createElement('li');
                        // Use downloadPath to download with full filesystem path
                        var buttons = '<button onclick="viewFile(\'' + joinPath(path, file.name) + '\')" class="button small">View</button>' +
                                      '<button onclick="downloadPath(\'' + joinPath(path, file.name) + '\')" class="button small">Download</button>';
                        if (file.name.match(/\.(mp3|wav|ogg|webm|m4a)$/i)) {
                            buttons += '<button onclick="playFile(\'' + joinPath(path, file.name) + '\')" class="button small">Play</button>';
                        }
                        li.innerHTML = '<span class="icon">📄 ' + file.name + '</span><span class="file-button-group">' + buttons + '</span>';
                        ul.appendChild(li);
                    });
                });
        }
        function parentDir(path) {
            if (path === '/' || !path) return '/';
            var parts = path.split('/').filter(Boolean);
            parts.pop();
            return '/' + parts.join('/');
        }
        function joinPath(base, name) {
            if (base.endsWith('/')) return base + name;
            return base + '/' + name;
        }
        function viewFile(path) {
            fetch('/api/v1/filesystem/content?path=' + encodeURIComponent(path))
                .then(function(res) { return res.json(); })
                .then(function(data) {
                    var contentDiv = document.getElementById('file-browser-content');
                    if (data.contentType && data.contentType.startsWith('text/')) {
                        contentDiv.innerHTML = '<div class="file-content">' + escapeHtml(data.content) + '</div>';
                    } else if (data.contentType && data.contentType.startsWith('image/')) {
                        contentDiv.innerHTML = '<img src="/api/v1/filesystem/serve?path=' + encodeURIComponent(path) + '" style="max-width:100%;border-radius:6px;" />';
                    } else {
                        // Use download query param to trigger file attachment
                        contentDiv.innerHTML = '<a href="/api/v1/filesystem/serve?path=' + encodeURIComponent(path) + '&download=true" download>Download file</a>';
                    }
                });
        }
        function escapeHtml(text) {
            return text.replace(/[&<>"']/g, function(m) {
                return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#39;'}[m];
            });
        }
        // Load root directory on tab open
        document.getElementById('tab-files').addEventListener('click', function() { loadFileBrowser('/'); });
        document.getElementById('tab-files-mobile').addEventListener('click', function() { loadFileBrowser('/'); });

        // Fetch and display files
        async function updateFileList() {
            try {
                const response = await fetch('/api/files');
                const data = await response.json();
                const fileList = document.getElementById('fileList');
                fileList.innerHTML = data.files.map(function(file) { return "<div class=\"file-item\"><span>" + file + "</span><button onclick=\"downloadFile('" + file + "')\" class=\"link-button\">Download</button></div>"; }).join('');
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

        // Live Clipboard Sync logic
        let liveClipboardEnabled = false;
        let clipboardSyncInterval = null;
        function toggleLiveClipboard() {
            liveClipboardEnabled = document.getElementById('liveClipboardToggle').checked;
            document.getElementById('liveClipboardStatus').textContent = liveClipboardEnabled ? 'ON' : 'OFF';
            if (liveClipboardEnabled) {
                // Request clipboard permissions on user gesture
                if (navigator.permissions) {
                    navigator.permissions.query({name: 'clipboard-read'}).then(function(result) {
                        if (result.state === 'denied') {
                            setClipboardSyncStatus('Clipboard read permission denied. Live sync will not work.');
                            document.getElementById('liveClipboardToggle').checked = false;
                            liveClipboardEnabled = false;
                            return;
                        }
                    });
                    navigator.permissions.query({name: 'clipboard-write'}).then(function(result) {
                        if (result.state === 'denied') {
                            setClipboardSyncStatus('Clipboard write permission denied. Live sync will not work.');
                            document.getElementById('liveClipboardToggle').checked = false;
                            liveClipboardEnabled = false;
                            return;
                        }
                    });
                }
                clipboardSyncInterval = setInterval(syncClipboardWithServer, 1500);
            } else {
                if (clipboardSyncInterval) clearInterval(clipboardSyncInterval);
            }
        }

        async function syncClipboardWithServer() {
            // Try to read from system clipboard (if allowed)
            if (navigator.clipboard && window.isSecureContext) {
                try {
                    const text = await navigator.clipboard.readText();
                    // Send to server if changed
                    await fetch('/api/clipboard', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({text})
                    });
                } catch (e) {
                    // Permission denied or not available
                }
            }
            // Optionally, fetch server clipboard and update local clipboard
            // Uncomment below to pull from server as well:
            // try {
            //     const res = await fetch('/api/clipboard');
            //     const data = await res.json();
            //     if (navigator.clipboard && window.isSecureContext) {
            //         await navigator.clipboard.writeText(data.text || '');
            //     }
            // } catch (e) {}
        }

        // Clipboard advanced controls
        async function sendClipboardToServer() {
            if (navigator.clipboard && window.isSecureContext) {
                try {
                    const text = await navigator.clipboard.readText();
                    await fetch('/api/clipboard', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({text})
                    });
                    setClipboardSyncStatus('Clipboard sent to all devices.');
                } catch (e) {
                    setClipboardSyncStatus('Failed to read clipboard.');
                }
            }
        }

        async function receiveClipboardFromServer() {
            try {
                const res = await fetch('/api/clipboard');
                const data = await res.json();
                if (navigator.clipboard && window.isSecureContext) {
                    await navigator.clipboard.writeText(data.text || '');
                    setClipboardSyncStatus('Clipboard received from server.');
                } else {
                    setClipboardSyncStatus('Clipboard API not available.');
                }
            } catch (e) {
                setClipboardSyncStatus('Failed to fetch clipboard from server.');
            }
        }

        async function sendManualClipboard() {
            const text = document.getElementById('manualClipboardInput').value;
            if (!text) return setClipboardSyncStatus('No text to send.');
            await fetch('/api/clipboard', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({text})
            });
            setClipboardSyncStatus('Manual text sent to all devices.');
        }

        function setClipboardSyncStatus(msg) {
            document.getElementById('clipboardSyncStatus').textContent = msg;
            setTimeout(() => { document.getElementById('clipboardSyncStatus').textContent = ''; }, 3000);
        }

        // Initialize
        updateFileList();
        fetchAudioFiles();

        // Attach file tab actions using configured root path
        var defaultRoot = (window._config && window._config.allowedPaths && window._config.allowedPaths.length) ? window._config.allowedPaths[0] : '/';
        // Only call when Files tab is active
        document.getElementById('tab-files').addEventListener('click', function(){ showFileSubTab('manager'); loadFileBrowser(defaultRoot); });
        document.getElementById('tab-files-mobile').addEventListener('click', function(){ showFileSubTab('manager'); loadFileBrowser(defaultRoot); });
        // Default to Manager view in Files
        showFileSubTab('manager');
        // Initial file browser load if Files tab active
        if (window._initialTab === 'files') {
            loadFileBrowser(defaultRoot);
        }

        // Play audio file inline in content area
        function playFile(path) {
            var contentDiv = document.getElementById('file-browser-content');
            contentDiv.innerHTML = '<audio controls style="width:100%; margin-top:1rem;" src="/api/v1/filesystem/serve?path=' + encodeURI(path) + '"></audio>';
        }

        // Download a file at given filesystem path via FileSystemAPI
        function downloadPath(path) {
            window.open('/api/v1/filesystem/serve?path=' + encodeURI(path) + '&download=true', '_blank');
        }
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
