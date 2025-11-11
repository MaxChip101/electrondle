package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gorilla/mux"
)

func HtmlStart(w http.ResponseWriter, title string) {
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>%s</title>
	</head>
	<body>
	`, title)
}

func HtmlEnd(w http.ResponseWriter) {
	fmt.Fprint(w, `
	</body>
	</html>
	`)
}

func HtmlInput(w http.ResponseWriter) {
	HtmlStart(w, "Electrondle")
	io.WriteString(w, `
		<form method="POST" action="/">
            <label for="username">Name:</label>
            <input type="text" id="username" name="username" required>
            <br><br>
            <button type="submit">send!</button>
        </form>
		`)
	HtmlEnd(w)
}

var value string

func HandleInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST Only", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Couldn't Parse Form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	value = r.FormValue("username")

	Display(w, r)
}

func Display(w http.ResponseWriter, r *http.Request) {
	HtmlStart(w, "Electrondle")
	HtmlInput(w)
	io.WriteString(w, value)
	HtmlEnd(w)
}

func OpenApp() {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", "http://localhost:3000").Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:3000").Start()
	case "darwin":
		err = exec.Command("open", "http://localhost:3000").Start()
	default:
		err = fmt.Errorf("unsupported operating system, couldn't automatically open")
	}
	if err != nil {
		log.Println(err)
	}
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", Display).Methods("GET")
	router.HandleFunc("/", HandleInput).Methods("POST")

	go func() {
		time.Sleep(time.Second)
		OpenApp()
		fmt.Println("opening app")
	}()

	fmt.Println("running server")
	err := http.ListenAndServe(":3000", router)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else {
		log.Fatal(err)
	}

}
