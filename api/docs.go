package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// APIEndpoint represents a documented API endpoint
type APIEndpoint struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Description string                 `json:"description"`
	Parameters  map[string]string      `json:"parameters,omitempty"`
	RequestBody map[string]interface{} `json:"requestBody,omitempty"`
	Response    map[string]interface{} `json:"response,omitempty"`
	Example     string                 `json:"example,omitempty"`
}

// APICategory groups endpoints by functionality
type APICategory struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

var apiDocs []APICategory

// InitDocs initializes the API documentation
func InitDocs() {
	// Clipboard operations
	apiDocs = append(apiDocs, APICategory{
		Name:        "Clipboard",
		Description: "Access and manipulate clipboard across devices",
		Endpoints: []APIEndpoint{
			{
				Path:        "/api/v1/clipboard",
				Method:      "GET",
				Description: "Retrieve the current clipboard content",
				Response: map[string]interface{}{
					"text": "clipboard content",
				},
				Example: "curl -X GET http://localhost:8080/api/v1/clipboard",
			},
			{
				Path:        "/api/v1/clipboard",
				Method:      "POST",
				Description: "Set new clipboard content",
				RequestBody: map[string]interface{}{
					"text": "content to set",
				},
				Response: map[string]interface{}{
					"status": "success",
				},
				Example: "curl -X POST -H \"Content-Type: application/json\" -d '{\"text\":\"Hello from API\"}' http://localhost:8080/api/v1/clipboard",
			},
			{
				Path:        "/api/v1/clipboard/history",
				Method:      "GET",
				Description: "Get clipboard history",
				Response: map[string]interface{}{
					"history": []map[string]interface{}{
						{"text": "content1", "timestamp": "2023-06-01T12:00:00Z"},
						{"text": "content2", "timestamp": "2023-06-01T12:05:00Z"},
					},
				},
				Example: "curl -X GET http://localhost:8080/api/v1/clipboard/history",
			},
		},
	})

	// File operations
	apiDocs = append(apiDocs, APICategory{
		Name:        "Files",
		Description: "Access and manage files",
		Endpoints: []APIEndpoint{
			{
				Path:        "/api/v1/files",
				Method:      "GET",
				Description: "List all available files in the uploads directory",
				Response: map[string]interface{}{
					"files": []string{"file1.txt", "image.jpg"},
				},
				Example: "curl -X GET http://localhost:8080/api/v1/files",
			},
			{
				Path:        "/api/v1/files",
				Method:      "POST",
				Description: "Upload a file",
				Parameters: map[string]string{
					"file": "File to upload (form-data)",
				},
				Response: map[string]interface{}{
					"status":   "success",
					"filename": "uploaded.txt",
				},
				Example: "curl -X POST -F \"file=@local_file.txt\" http://localhost:8080/api/v1/files",
			},
			{
				Path:        "/api/v1/files/:filename",
				Method:      "GET",
				Description: "Download a file",
				Parameters: map[string]string{
					"filename": "Name of the file to download",
				},
				Example: "curl -X GET http://localhost:8080/api/v1/files/document.pdf -o downloaded.pdf",
			},
			{
				Path:        "/api/v1/files/:filename",
				Method:      "DELETE",
				Description: "Delete a file",
				Parameters: map[string]string{
					"filename": "Name of the file to delete",
				},
				Response: map[string]interface{}{
					"status": "success",
				},
				Example: "curl -X DELETE http://localhost:8080/api/v1/files/document.pdf",
			},
			{
				Path:        "/api/v1/filesystem/list",
				Method:      "GET",
				Description: "List contents of a directory",
				Parameters: map[string]string{
					"path": "Directory path to list",
				},
				Response: map[string]interface{}{
					"path": "/path/to/dir",
					"directories": []string{"dir1", "dir2"},
					"files": []map[string]interface{}{
						{"name": "file1.txt", "size": 1024, "modified": "2023-06-01T12:00:00Z"},
					},
				},
				Example: "curl -X GET \"http://localhost:8080/api/v1/filesystem/list?path=/home/user/Documents\"",
			},
			{
				Path:        "/api/v1/filesystem/content",
				Method:      "GET",
				Description: "Get file content",
				Parameters: map[string]string{
					"path": "Path to file",
				},
				Response: map[string]interface{}{
					"content": "file content...",
				},
				Example: "curl -X GET \"http://localhost:8080/api/v1/filesystem/content?path=/home/user/file.txt\"",
			},
		},
	})

	// Shell commands
	apiDocs = append(apiDocs, APICategory{
		Name:        "Shell",
		Description: "Execute shell commands",
		Endpoints: []APIEndpoint{
			{
				Path:        "/api/v1/shell/exec",
				Method:      "POST",
				Description: "Execute a shell command",
				RequestBody: map[string]interface{}{
					"command": "ls -la",
					"timeout": 10, // seconds
				},
				Response: map[string]interface{}{
					"stdout":   "command output",
					"stderr":   "error output if any",
					"exitCode": 0,
				},
				Example: "curl -X POST -H \"Content-Type: application/json\" -d '{\"command\":\"ls -la\"}' http://localhost:8080/api/v1/shell/exec",
			},
			{
				Path:        "/api/v1/shell/stream",
				Method:      "GET",
				Description: "Stream a long-running command (WebSocket)",
				Parameters: map[string]string{
					"command": "Command to execute",
				},
				Example: "Accessible via WebSocket: ws://localhost:8080/api/v1/shell/stream?command=top",
			},
		},
	})

	// System information
	apiDocs = append(apiDocs, APICategory{
		Name:        "System",
		Description: "Access system information",
		Endpoints: []APIEndpoint{
			{
				Path:        "/api/v1/system/info",
				Method:      "GET",
				Description: "Get system information",
				Response: map[string]interface{}{
					"hostname":  "computer-name",
					"platform":  "linux",
					"cpuUsage":  "23.5%",
					"memoryUsage": "45.2%",
					"uptime":    "3d 12h 5m",
				},
				Example: "curl -X GET http://localhost:8080/api/v1/system/info",
			},
			{
				Path:        "/api/v1/system/notify",
				Method:      "POST",
				Description: "Send a system notification",
				RequestBody: map[string]interface{}{
					"title":   "Notification Title",
					"message": "Notification content",
					"type":    "info", // info, warning, error
				},
				Response: map[string]interface{}{
					"status": "success",
				},
				Example: "curl -X POST -H \"Content-Type: application/json\" -d '{\"title\":\"Alert\",\"message\":\"Server restart required\"}' http://localhost:8080/api/v1/system/notify",
			},
		},
	})

	// Media streaming
	apiDocs = append(apiDocs, APICategory{
		Name:        "Media",
		Description: "Stream and control media",
		Endpoints: []APIEndpoint{
			{
				Path:        "/api/v1/media/audio/devices",
				Method:      "GET",
				Description: "List available audio devices",
				Response: map[string]interface{}{
					"devices": []map[string]interface{}{
						{"id": "default", "name": "System Default"},
						{"id": "hw:0,0", "name": "Built-in Audio"},
					},
				},
				Example: "curl -X GET http://localhost:8080/api/v1/media/audio/devices",
			},
			{
				Path:        "/api/v1/media/audio/stream",
				Method:      "GET",
				Description: "Stream system audio output",
				Parameters: map[string]string{
					"device": "Audio device ID (optional)",
				},
				Example: "Accessible via WebSocket: ws://localhost:8080/api/v1/media/audio/stream?device=default",
			},
			{
				Path:        "/api/v1/media/screen",
				Method:      "GET",
				Description: "Stream screen content",
				Parameters: map[string]string{
					"quality": "Stream quality (low, medium, high)",
					"fps":     "Frames per second (1-30)",
				},
				Example: "Accessible via WebSocket: ws://localhost:8080/api/v1/media/screen?quality=medium&fps=15",
			},
		},
	})
	
	// Sort categories alphabetically
	sort.Slice(apiDocs, func(i, j int) bool {
		return apiDocs[i].Name < apiDocs[j].Name
	})
}

// ServeAPIDocsJSON serves the API documentation as JSON
func ServeAPIDocsJSON(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apiVersion": "v1",
		"categories": apiDocs,
	})
}

// ServeAPIDocsUI serves the HTML API documentation page
func ServeAPIDocsUI(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, apiDocsTemplate)
}

const apiDocsTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>NoPlaceLike API Documentation</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        :root {
            --primary: #4444ff;
            --dark: #1a1a2e;
            --light: #f5f5f5;
            --code-bg: #282a36;
            --code-text: #f8f8f2;
            --border: #ddd;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: system-ui, -apple-system, sans-serif;
            line-height: 1.6;
            background: var(--light);
            color: #333;
            padding-bottom: 5rem;
        }
        header {
            background: var(--dark);
            color: white;
            padding: 2rem;
            text-align: center;
        }
        header h1 { margin-bottom: 0.5rem; }
        main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        .api-version {
            display: inline-block;
            background: var(--primary);
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8rem;
            margin-left: 0.5rem;
        }
        .toc {
            margin: 2rem 0;
            padding: 1.5rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .toc h2 {
            margin-bottom: 1rem;
            border-bottom: 1px solid var(--border);
            padding-bottom: 0.5rem;
        }
        .toc-list {
            list-style-type: none;
        }
        .toc-category {
            margin-bottom: 0.5rem;
            font-weight: 600;
        }
        .toc-endpoints {
            list-style-type: none;
            padding-left: 1.5rem;
            margin-bottom: 1rem;
        }
        .toc-endpoint {
            margin: 0.25rem 0;
        }
        .toc-endpoint a {
            color: var(--primary);
            text-decoration: none;
        }
        .toc-endpoint a:hover {
            text-decoration: underline;
        }
        .category {
            margin: 3rem 0;
        }
        .category-header {
            background: var(--dark);
            color: white;
            padding: 1rem;
            border-radius: 8px 8px 0 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .category-description {
            padding: 1rem;
            background: rgba(255,255,255,0.5);
            border: 1px solid var(--border);
            border-top: none;
        }
        .endpoint {
            margin: 1.5rem 0;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .endpoint-header {
            padding: 1rem;
            display: flex;
            align-items: center;
            background: #f8f9fa;
            border-bottom: 1px solid var(--border);
        }
        .http-method {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            color: white;
            font-weight: bold;
            min-width: 60px;
            text-align: center;
            margin-right: 1rem;
        }
        .get { background: #61affe; }
        .post { background: #49cc90; }
        .put { background: #fca130; }
        .delete { background: #f93e3e; }
        .endpoint-path {
            font-family: monospace;
            font-size: 1.1rem;
        }
        .endpoint-content {
            padding: 1rem;
        }
        .endpoint-section {
            margin: 1rem 0;
        }
        .endpoint-section h4 {
            margin-bottom: 0.5rem;
            padding-bottom: 0.25rem;
            border-bottom: 1px solid var(--border);
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 1rem 0;
        }
        th, td {
            padding: 0.5rem;
            text-align: left;
            border: 1px solid var(--border);
        }
        th {
            background: #f5f5f5;
        }
        .code-block {
            background: var(--code-bg);
            color: var(--code-text);
            padding: 1rem;
            border-radius: 4px;
            overflow-x: auto;
            font-family: monospace;
            margin: 1rem 0;
        }
        .try-btn {
            background: var(--primary);
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
        }
        .try-btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <header>
        <h1>NoPlaceLike API Documentation <span class="api-version">v1</span></h1>
        <p>A comprehensive API for accessing system resources across your network</p>
    </header>

    <main>
        <div class="toc">
            <h2>Table of Contents</h2>
            <ul class="toc-list" id="tocList">
                <!-- Dynamically generated -->
            </ul>
        </div>

        <div id="apiContent">
            <!-- Dynamically generated -->
        </div>
    </main>

    <script>
        async function loadDocs() {
            try {
                const response = await fetch('/api/v1/docs/json');
                const data = await response.json();
                
                // Build table of contents
                const tocList = document.getElementById('tocList');
                const apiContent = document.getElementById('apiContent');
                
                data.categories.forEach(category => {
                    // Add to TOC
                    const categoryItem = document.createElement('li');
                    categoryItem.className = 'toc-category';
                    categoryItem.innerHTML = "<a href=\"#" + slugify(category.name) + "\">" + category.name + "</a>";
                    tocList.appendChild(categoryItem);
                    
                    const endpointsList = document.createElement('ul');
                    endpointsList.className = 'toc-endpoints';
                    
                    category.endpoints.forEach(endpoint => {
                        const endpointId = slugify(category.name + '-' + endpoint.method + '-' + endpoint.path);
                        const endpointItem = document.createElement('li');
                        endpointItem.className = 'toc-endpoint';
                        endpointItem.innerHTML = "<a href=\"#" + endpointId + "\">" + endpoint.method + " " + endpoint.path + "</a>";
                        endpointsList.appendChild(endpointItem);
                    });
                    
                    tocList.appendChild(endpointsList);
                    
                    // Add category content
                    const categoryDiv = document.createElement('div');
                    categoryDiv.className = 'category';
                    categoryDiv.id = slugify(category.name);
                    
                    categoryDiv.innerHTML = 
                        '<div class="category-header">' +
                        '<h2>' + category.name + '</h2>' +
                        '</div>' +
                        '<div class="category-description">' +
                        '<p>' + category.description + '</p>' +
                        '</div>';
                    
                    // Add endpoints
                    category.endpoints.forEach(endpoint => {
                        const endpointId = slugify(category.name + '-' + endpoint.method + '-' + endpoint.path);
                        const endpointDiv = document.createElement('div');
                        endpointDiv.className = 'endpoint';
                        endpointDiv.id = endpointId;
                        
                        let endpointContent = 
                            '<div class="endpoint-header">' +
                            '<span class="http-method ' + endpoint.method.toLowerCase() + '">' + endpoint.method + '</span>' +
                            '<span class="endpoint-path">' + endpoint.path + '</span>' +
                            '</div>' +
                            '<div class="endpoint-content">' +
                            '<p>' + endpoint.description + '</p>';
                        
                        // Parameters
                        if (endpoint.parameters && Object.keys(endpoint.parameters).length > 0) {
                            endpointContent += 
                                '<div class="endpoint-section">' +
                                '<h4>Parameters</h4>' +
                                '<table>' +
                                '<thead>' +
                                '<tr>' +
                                '<th>Name</th>' +
                                '<th>Description</th>' +
                                '</tr>' +
                                '</thead>' +
                                '<tbody>';
                            
                            for (const [param, desc] of Object.entries(endpoint.parameters)) {
                                endpointContent += 
                                    '<tr>' +
                                    '<td><code>' + param + '</code></td>' +
                                    '<td>' + desc + '</td>' +
                                    '</tr>';
                            }
                            
                            endpointContent += 
                                '</tbody>' +
                                '</table>' +
                                '</div>';
                        }
                        
                        // Request Body
                        if (endpoint.requestBody && Object.keys(endpoint.requestBody).length > 0) {
                            endpointContent += 
                                '<div class="endpoint-section">' +
                                '<h4>Request Body</h4>' +
                                '<div class="code-block">' + stringify(endpoint.requestBody) + '</div>' +
                                '</div>';
                        }
                        
                        // Response
                        if (endpoint.response && Object.keys(endpoint.response).length > 0) {
                            endpointContent += 
                                '<div class="endpoint-section">' +
                                '<h4>Response</h4>' +
                                '<div class="code-block">' + stringify(endpoint.response) + '</div>' +
                                '</div>';
                        }
                        
                        // Example
                        if (endpoint.example) {
                            endpointContent += 
                                '<div class="endpoint-section">' +
                                '<h4>Example</h4>' +
                                '<div class="code-block">' + endpoint.example + '</div>' +
                                '</div>';
                        }
                        
                        endpointContent += '</div>';
                        endpointDiv.innerHTML = endpointContent;
                        categoryDiv.appendChild(endpointDiv);
                    });
                    
                    apiContent.appendChild(categoryDiv);
                });
            } catch (error) {
                console.error('Error loading API docs:', error);
            }
        }
        
        function slugify(text) {
            return text
                .toLowerCase()
                .replace(/[^\w ]+/g, '')
                .replace(/ +/g, '-');
        }
        
        function stringify(obj) {
            return JSON.stringify(obj, null, 2);
        }
        
        document.addEventListener('DOMContentLoaded', loadDocs);
    </script>
</body>
</html>
`
