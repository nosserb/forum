package handlers

import (
	"forum/controller/logging"
	forumDB "forum/model/functions"
	"net/http"
	"strconv"
)

// Serve images
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch ID, redirect if not found
	imageIdStr := r.URL.Query().Get("id")
	if imageIdStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	imageID, err := strconv.Atoi(imageIdStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid image ID")
		return
	}

	// Fetch image
	img, err := forumDB.FetchImage(db, int64(imageID))
	if err != nil {
		ErrorHandler(w, r, http.StatusNotFound, "Image not found")
		logging.Logger.Printf("Image not found : %d", imageID)
		return
	}

	detectedType := http.DetectContentType(img.Data)
	if detectedType != img.Type {
		ErrorHandler(w, r, 500, "DB indicates wrong image type")
		return
	}

	// write image to response
	w.Header().Set("Content-Type", img.Type)
	w.Write(img.Data)

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusOK)
}
