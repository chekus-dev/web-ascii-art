package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)


// validateInput checks every character is printable ASCII.
func validateInput(input string) error {
	for _, char := range input { // Loop through each character in the input string
		// ASCII characters range from ' ' (32) to '~' (tilde). If the character is outside this range, it is not supported.
		if char < 32 || char > 126 {
			return fmt.Errorf("unsupported character") // Return an error if an unsupported character is found

		}
	}
	return nil // Return nil if all characters are valid
}



// renderWord builds the 8-row ASCII-art block for a single word.
func renderWord(word string, bannerLines []string) (string, error) {
	var out strings.Builder        // Use strings.Builder for efficient string concatenation
	for row := 0; row < 8; row++ { // Loop through each of the 8 rows for the ASCII-art representation
		for _, char := range word {
			index := row + (int(char-' ') * 9) + 1      // Calculate the index in bannerLines for the current character and row
			if index < 0 || index >= len(bannerLines) { // Validate the index to ensure it is within the bounds of bannerLines
				return "", fmt.Errorf("invalid banner file") // Return an error if the index is out of bounds
			}
			out.WriteString(bannerLines[index]) // Append the ASCII-art line for the current character and row
		}
		out.WriteString("\n") // Add a newline after each row to separate the lines of the ASCII-art block
	}
	return out.String(), nil // Return the complete ASCII-art block for the word as a string
}


// loadBanner reads and validates the banner file, returning its lines.
func loadBanner(fileName string) ([]string, error) {
	data, err := os.ReadFile(fileName) // Read the banner file
	if err != nil {
		return nil, fmt.Errorf("reading banner file: %w", err)
	}
	if len(data) == 0 { // Check if the file is empty
		return []string{}, fmt.Errorf("error: banner files empty") // Handle empty file case
	}
	input := strings.ReplaceAll(string(data), "\r\n", "\n") // Normalize line endings to handle different platforms
	lines := strings.Split(input, "\n")                     // Validate the number of lines in the banner file
	if len(data) < 855 {
		return nil, fmt.Errorf("invalid banner file")
	}
	return lines, nil // Return the lines of the banner file
}

// asciiHandler handles the /ascii route
func asciiHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Read query parameters
	text := r.URL.Query().Get("text")
	banner := r.URL.Query().Get("banner")

	// 3. Check if text is provided
	if text == "" {
		http.Error(w, "Please provide text", http.StatusBadRequest)
		return
	}

	// 4. Default banner to standard if not provided
	if banner == "" {
		banner = "standard"
	}

	// 5. Validate the input
	if err := validateInput(text); err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 6. Load the banner file
	bannerLines, err := loadBanner(banner + ".txt")
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 7. Handle newlines in the text
	var result strings.Builder
	words := strings.Split(text, "\\n")
	for i, word := range words {
		if word == "" {
			result.WriteString("\n")
			continue
		}
		// 8. Render each word
		art, err := renderWord(word, bannerLines)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result.WriteString(art)
		// Add newline between words but not after the last one
		if i != len(words)-1 {
			result.WriteString("\n")
		}
	}

	// 9. Send the result back
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, result.String())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ascii", asciiHandler)

	log.Println("Starting server on port :8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
