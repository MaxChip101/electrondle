package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Attempt struct {
	NobleGas string
	S        int
	D        int
	P        int
}

type PeriodicTable struct {
	Elements []Element
}

type Element struct {
	Name     string `json:"Name"`
	Symbol   string `json:"symbol"`
	NobleGas string `json:"noble-gas-prefix"`
	S_Prefix int    `json:"s-prefix"`
	S_Suffix int    `json:"s-suffix"`
	D_Prefix int    `json:"d-prefix"`
	D_Suffix int    `json:"d-suffix"`
	P_Prefix int    `json:"p-prefix"`
	P_Suffix int    `json:"p-suffix"`
}

var random *rand.Rand
var periodic_table PeriodicTable
var current_element Element
var attempts []Attempt

func HtmlStart(w http.ResponseWriter) {
	io.WriteString(w, `
<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta
      		name="viewport"
      		content="width=device-width,minimum-scale=1,initial-scale=1"
    	/>
		<title>Electrondle</title>
		<link rel="stylesheet" href="/res/style.css">
		<link rel="icon" type="image/x-icon" href="/res/favicon.ico">
	</head>
	<body class="body">
		<h1 class="title">Electrondle</h1>
	`)
}

func StartAttempts(w http.ResponseWriter) {
	io.WriteString(w, `
	<div class="attempts">
	`)
}

func GetNobleGasNumber(noble_gas string) int {
	switch noble_gas {
	case "~":
		return 0
	case "He":
		return 1
	case "Ne":
		return 2
	case "Ar":
		return 3
	case "Kr":
		return 4
	case "Xe":
		return 5
	case "Rn":
		return 6
	default:
		return -1
	}
}

func CheckDifference(difference int) string {
	if difference == 0 {
		return "✅"
	} else if difference >= 4 {
		return "⏫"
	} else if difference <= -4 {
		return "⏬"
	} else if difference >= 1 {
		return "🔼"
	} else if difference <= -1 {
		return "🔽"
	}
	return "~"
}

func CheckNobleGas(attempt Attempt) string {
	return CheckDifference(GetNobleGasNumber(current_element.NobleGas) - GetNobleGasNumber(attempt.NobleGas))
}

func Check_S(attempt Attempt) string {
	return CheckDifference(current_element.S_Suffix - attempt.S)
}

func Check_D(attempt Attempt) string {
	return CheckDifference(current_element.D_Suffix - attempt.D)
}

func Check_P(attempt Attempt) string {
	return CheckDifference(current_element.P_Suffix - attempt.P)
}

// ⏫🔼	✅	🔽⏬
func AddAttempts(w http.ResponseWriter) {

	for _, v := range attempts {
		fmt.Fprintf(w, `
	<div class="attempt">
        <pre>[%v]	s^%v	d^%v	p^%v</pre>
        <pre>%v	%v	%v	%v</pre>
    </div>
	`, v.NobleGas, v.S, v.D, v.P, CheckNobleGas(v), Check_S(v), Check_D(v), Check_P(v))
	}
}

func EndAttempts(w http.ResponseWriter) {
	io.WriteString(w, `
		</div>
	`)
}

func HtmlEnd(w http.ResponseWriter) {
	fmt.Fprint(w, `
	</body>
	</html>
	`)
}

func HtmlInput(w http.ResponseWriter) {
	noble_gas := "~"
	s := 1
	d := 0
	p := 0

	if pos := len(attempts) - 1; pos >= 0 {
		noble_gas = attempts[pos].NobleGas
		s = attempts[pos].S
		d = attempts[pos].D
		p = attempts[pos].P
	}

	fmt.Fprintf(w, `
	<form method="POST" action="/submit">
      <div class="inputs">
    	<div>
          <select id="noble-gas" class="dropdown" name="noble-gas">
            <option value="~" %v>None</option>
            <option value="He" %v>[He]</option>
            <option value="Ne" %v>[Ne]</option>
            <option value="Ar" %v>[Ar]</option>
            <option value="Kr" %v>[Kr]</option>
            <option value="Xe" %v>[Xe]</option>
            <option value="Rn" %v>[Rn]</option>
          </select>
        </div>
        <div>
          <label class="text">s^</label>
          <select id="s-orbital" class="dropdown" name="s-orbital">
            <option value="1" %v>1</option>
            <option value="2" %v>2</option>
          </select>
        </div>
        <div>
          <label class="text">d^</label>
          <select id="d-orbital" class="dropdown" name="d-orbital">
            <option value="0" %v>0</option>
            <option value="1" %v>1</option>
            <option value="2" %v>2</option>
            <option value="3" %v>3</option>
            <option value="4" %v>4</option>
            <option value="5" %v>5</option>
            <option value="6" %v>6</option>
            <option value="7" %v>7</option>
            <option value="8" %v>8</option>
            <option value="9" %v>9</option>
            <option value="10" %v>10</option>
          </select>
        </div>
        <div>
          <label class="text">p^</label>
          <select id="p-orbital" class="dropdown" name="p-orbital">
            <option value="0" %v>0</option>
            <option value="1" %v>1</option>
            <option value="2" %v>2</option>
            <option value="3" %v>3</option>
            <option value="4" %v>4</option>
            <option value="5" %v>5</option>
            <option value="6" %v>6</option>
          </select>
        </div>
      </div>
      <button class="submit-button" type="submit">
        <span>Submit</span>
      </button>
    </form>
	`,
		// remember last played values
		Selected(noble_gas == "~"),
		Selected(noble_gas == "He"),
		Selected(noble_gas == "Ne"),
		Selected(noble_gas == "Ar"),
		Selected(noble_gas == "Kr"),
		Selected(noble_gas == "Xe"),
		Selected(noble_gas == "Rn"),
		Selected(s == 1),
		Selected(s == 2),
		Selected(d == 0),
		Selected(d == 1),
		Selected(d == 2),
		Selected(d == 3),
		Selected(d == 4),
		Selected(d == 5),
		Selected(d == 6),
		Selected(d == 7),
		Selected(d == 8),
		Selected(d == 9),
		Selected(d == 10),
		Selected(p == 0),
		Selected(p == 1),
		Selected(p == 2),
		Selected(p == 3),
		Selected(p == 4),
		Selected(p == 5),
		Selected(p == 6),
	)
}

func Selected(condition bool) string {
	if condition {
		return "selected=\"selected\""
	} else {
		return ""
	}
}

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

	s_value, err := strconv.ParseInt(r.FormValue("s-orbital"), 10, strconv.IntSize)
	if err != nil {
		log.Fatal(err)
	}
	d_value, err := strconv.ParseInt(r.FormValue("d-orbital"), 10, strconv.IntSize)
	if err != nil {
		log.Fatal(err)
	}
	p_value, err := strconv.ParseInt(r.FormValue("p-orbital"), 10, strconv.IntSize)
	if err != nil {
		log.Fatal(err)
	}

	attempts = append(attempts, Attempt{
		NobleGas: r.FormValue("noble-gas"),
		S:        int(s_value),
		D:        int(d_value),
		P:        int(p_value),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Display(w http.ResponseWriter, r *http.Request) {
	HtmlStart(w)
	StartAttempts(w)
	AddAttempts(w)
	EndAttempts(w)
	HtmlInput(w)
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

func Start() {
	data, err := os.ReadFile(filepath.Join("res", "periodic_table.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &periodic_table)
	if err != nil {
		log.Fatal(err)
	}

	num := random.Int() % len(periodic_table.Elements)
	current_element = periodic_table.Elements[num]
}

func main() {
	random = rand.New(rand.NewSource(time.Now().Unix()))
	router := mux.NewRouter()
	router.HandleFunc("/", Display).Methods("GET")
	router.HandleFunc("/submit", HandleInput).Methods("POST")

	router.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(http.Dir("./res"))))

	go func() {
		time.Sleep(time.Second)
		OpenApp()
		fmt.Println("opening app")
	}()

	Start()
	fmt.Println("running server")
	err := http.ListenAndServe(":3000", router)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else {
		log.Fatal(err)
	}

}
