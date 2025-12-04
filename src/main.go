package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	NobleGas string `json:"noble-gas"`
	S        int    `json:"s"`
	D        int    `json:"d"`
	P        int    `json:"p"`
}

var (
	GameStatePlaying = 0
	GameStateWon     = 1
	GameStateLost    = 2
)

var random *rand.Rand
var periodic_table PeriodicTable
var current_element Element
var attempts []Attempt
var maxAttempts int

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

// Converts the noble gas into a number able to tell the direction of how correct the response was
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

// returns the character to indicate the direction
func CheckNobleGas(attempt Attempt) string {
	return CheckDifference(GetNobleGasNumber(current_element.NobleGas) - GetNobleGasNumber(attempt.NobleGas))
}

func Check_S(attempt Attempt) string {
	return CheckDifference(current_element.S - attempt.S)
}

func Check_D(attempt Attempt) string {
	return CheckDifference(current_element.D - attempt.D)
}

func Check_P(attempt Attempt) string {
	return CheckDifference(current_element.P - attempt.P)
}

func CheckWinState() int {
	// win (all orbitals equal)
	if pos := len(attempts) - 1; pos >= 0 && current_element.NobleGas == attempts[pos].NobleGas && current_element.S == attempts[pos].S && current_element.D == attempts[pos].D && current_element.P == attempts[pos].P {
		return GameStateWon
	} else if len(attempts) >= maxAttempts { // loss (out of attempts)
		return GameStateLost
	} else { // continue
		return GameStatePlaying
	}
}

func AddAttempts(w http.ResponseWriter) {

	state := CheckWinState()

	switch state {
	case GameStatePlaying:
		fmt.Fprintf(w, `
		<div class="attempt">
			<span>%v attempts left</span>
		</div>
		`,
			maxAttempts-len(attempts),
		)
	case GameStateWon:
		fmt.Fprintf(w, `
		<div class="attempt">
			<div class="attempt-row">
				<span>You Won, the element was: %v (%v)</span>
			</div>
			<div class="attempt-row">
				<span>([%v] s<sup>%v</sup> d<sup>%v</sup> p<sup>%v</sup>)</span>
			</div>
		</div>
		`,
			current_element.Name,
			current_element.Symbol,
			current_element.NobleGas,
			current_element.S,
			current_element.D,
			current_element.P,
		)
	case GameStateLost:
		fmt.Fprintf(w, `
		<div class="attempt">
			<div class="attempt-row">
				<span>You Lost, the element was: %v (%v)</span>
			</div>
			<div class="attempt-row">
				<span>([%v] s<sup>%v</sup> d<sup>%v</sup> p<sup>%v</sup>)</span>
			</div>
		</div>
		`,
			current_element.Name,
			current_element.Symbol,
			current_element.NobleGas,
			current_element.S,
			current_element.D,
			current_element.P,
		)
	}

	for _, v := range attempts {
		fmt.Fprintf(w, `
	<div class="attempt">
        <div class="attempt-row">
        <span>[%v]</span>
        <span>s<sup>%v</sup></span>
        <span>d<sup>%v</sup></span>
        <span>p<sup>%v</sup></span>
        </div>
        <div class="attempt-row">
        <span>%v</span>
        <span>%v</span>
        <span>%v</span>
        <span>%v</span>
        </div>
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

	if CheckWinState() != 0 {
		io.WriteString(w, `
		<form action="/try-again">
			<button class="button" type="submit">
       			<span>Try Again</span>
      		</button>
		</form>
		`)
		return
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
      <button class="button" type="submit">
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

// resets the game
func HandleRestart(w http.ResponseWriter, r *http.Request) {
	Start()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleSubmit(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Couldn't Parse Form: "+err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
	d_value, err := strconv.ParseInt(r.FormValue("d-orbital"), 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Couldn't Parse Form: "+err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
	p_value, err := strconv.ParseInt(r.FormValue("p-orbital"), 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Couldn't Parse Form: "+err.Error(), http.StatusInternalServerError)
		log.Println(err)
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

func OpenApp(port int) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", fmt.Sprintf("http://localhost:%v", port)).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", fmt.Sprintf("http://localhost:%v", port)).Start()
	case "darwin":
		err = exec.Command("open", fmt.Sprintf("http://localhost:%v", port)).Start()
	default:
		err = fmt.Errorf("unsupported operating system, couldn't automatically open")
	}
	if err != nil {
		log.Println(err)
	}
}

func Start() {
	num := random.Int() % len(periodic_table.Elements)
	current_element = periodic_table.Elements[num]
	if attempts != nil {
		attempts = slices.Delete(attempts, 0, len(attempts))
	}

}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func main() {
	maxAttempts = 5
	data, err := os.ReadFile(filepath.Join("res", "periodic_table.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &periodic_table)
	if err != nil {
		log.Fatal(err)
	}

	random = rand.New(rand.NewSource(time.Now().Unix()))
	router := mux.NewRouter()
	router.HandleFunc("/", Display)
	router.HandleFunc("/submit", HandleSubmit).Methods("POST")
	router.HandleFunc("/try-again", HandleRestart)

	router.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(http.Dir("./res"))))

	port, err := GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		OpenApp(port)
		fmt.Println("opening app")
	}()

	Start()
	fmt.Println("running server on port: ", port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), router)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else {
		log.Fatal(err)
	}

}
