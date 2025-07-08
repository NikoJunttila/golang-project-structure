package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nikojunttila/community/internal/logger"
)

// GetUploadPageHandler renders html page for file upload
func GetUploadPageHandler(w http.ResponseWriter, _ *http.Request) {
	if err := templates.ExecuteTemplate(w, "upload.html", nil); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

// PostFileUploadHandler saves a file locally
func PostFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 10MB. This line saves you from those accidental 100MB uploads!
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		RespondWithError(r.Context(), w, 400, "Failed to parse file", err)
		return
	}

	// Retrieve the file from form data
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error(r.Context(), err, "Failed to close file")
		}
	}()

	_, _ = fmt.Fprintf(w, "Uploaded File: %s\n", handler.Filename)
	_, _ = fmt.Fprintf(w, "File Size: %d\n", handler.Size)
	_, _ = fmt.Fprintf(w, "MIME Header: %v\n", handler.Header)

	// Now let’s save it locally
	dst, err := createFile(handler.Filename)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := dst.Close(); err != nil {
			logger.Error(r.Context(), err, "Failed to destination file")
		}
	}()

	// Copy the uploaded file to the destination file
	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
	}
}
func createFile(filename string) (*os.File, error) {
	// Create an uploads directory if it doesn’t exist
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			return nil, err
		}
	}
	// Build the file path and create it
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		return nil, err
	}

	return dst, nil
}
