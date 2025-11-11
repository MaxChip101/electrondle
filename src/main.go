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

func Display(w http.ResponseWriter, r *http.Request) {
	HtmlStart(w, "Electrondle")
	io.WriteString(w, "real")
	HtmlEnd(w)
}

func OpenApp() {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", "localhost:3000").Start()
	case "windows":
		err = exec.Command("rundll32", "localhost:3000").Start()
	case "darwin":
		err = exec.Command("open", "localhost:3000").Start()
	default:
		err = fmt.Errorf("unsupported operating system, couldn't automatically open")
	}
	if err != nil {
		log.Println(err)
	}
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", Display)

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
