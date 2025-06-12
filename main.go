package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
)

var tpl *template.Template
var outputString string

func main() {
	var err error
	tpl, err = template.ParseGlob("webpages/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/imagescan", imageScanHandler)
	mux.HandleFunc("/scanresult", scanResultHandler)

	go func() {
		log.Println("Starting server on :8888")
		if err := http.ListenAndServe(":8888", mux); err != nil && err != http.ErrServerClosed {
			log.Printf("Server on :8888 encountered an error: %v", err)
		}
		log.Println("Server on :8888 stopped.")
	}()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil && err != http.ErrServerClosed {
		log.Printf("Server on :8080 encountered an error: %v", err)
	}
	log.Println("Server on :8080 stopped.")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func scanResultHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "scanresults.html", outputString)
	if err != nil {
		log.Printf("Error executing scan result template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func imageScanHandler(w http.ResponseWriter, r *http.Request) {
	imageName := r.FormValue("imageName")
	optionSelected := r.FormValue("outputFormat")

	err := tpl.ExecuteTemplate(w, "loadpage.html", nil)
	if err != nil {
		log.Printf("Error executing load page template: %v", err)
		return
	}

	cmd := exec.Command("trivy", "image", "-f", optionSelected, imageName)
	output, cmdErr := cmd.Output()

	if cmdErr != nil {
		log.Printf("Trivy command error for image '%s': %v", imageName, cmdErr)
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			log.Printf("Trivy stderr: %s", string(exitErr.Stderr))
			if exitErr.ExitCode() == 1 {
				outputString = fmt.Sprintf("Error scanning image '%s'. Please make sure the image exists and is accessible. Trivy stderr: %s", imageName, string(exitErr.Stderr))
			} else {
				outputString = fmt.Sprintf("Trivy command failed for image '%s' with exit code %d. Stderr: %s", imageName, exitErr.ExitCode(), string(exitErr.Stderr))
			}
		} else {
			outputString = fmt.Sprintf("Error executing Trivy for image '%s': %v", imageName, cmdErr)
		}
	} else {
		outputString = string(output)
	}
}
