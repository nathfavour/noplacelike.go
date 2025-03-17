// ...existing code...

// redirectToDocumentation redirects to API documentation
func (a *API) redirectToDocumentation(c *gin.Context) {
	c.Redirect(http.StatusFound, "/api/v1/docs")
}

// listFiles lists all files in the uploads directory
func (a *API) listFiles(c *gin.Context) {
	uploadDir := expandPath(a.config.UploadFolder)
	files, err := listFilesInDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// uploadFile handles file upload
func (a *API) uploadFile(c *gin.Context) {
	uploadDir := expandPath(a.config.UploadFolder)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	
	// Ensure filename is safe
	filename := getSafeFilename(file.Filename)
	
	// Save the file
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"filename": filename,
	})
}

// downloadFile serves a file for download
func (a *API) downloadFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",
		})
		return
	}
	
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Serve the file
	c.File(filepath)
}

// deleteFile deletes a file
func (a *API) deleteFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No filename specified",
		})
		return
	}
	
	// Ensure the filename doesn't contain path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	filepath := filepath.Join(expandPath(a.config.UploadFolder), filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Delete the file
	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// listFilesInDir returns a list of files in a directory
func listFilesInDir(dir string) ([]string, error) {
	// Ensure directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		return []string{}, nil // Return empty list for new directory
	}
	
	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	// Extract filenames
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// getSafeFilename ensures a filename is safe for use
func getSafeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)
	
	// Replace potentially problematic characters
	replacer := strings.NewReplacer(
		"../", "",
		"./", "",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	
	return replacer.Replace(filename)
}
