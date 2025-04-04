package utils

import "net/http"

func GetPlatformFromRequest(w http.ResponseWriter, r *http.Request) string {

	platform := r.URL.Query().Get("platform")

	if platform == "" {
		http.Error(w, "A valid platform is required.", http.StatusBadRequest)
		return ""
	}

	return platform

}

func GetCodeFromRequest(w http.ResponseWriter, r *http.Request) string {

	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return ""
	}

	return code

}

func GetStateString(w http.ResponseWriter) string {

	state, err := GenerateRandomString()

	if err != nil {
		http.Error(w, "Failed to generate state string", http.StatusInternalServerError)
		return ""
	}

	return state

}
