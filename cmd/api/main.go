package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var keys []string
var data map[string]string

var users []string

var correctUsername = "puppet"
var correctPassword = "password"

var lock = &sync.RWMutex{}
var tokens map[string]struct{}

var passwordFile string

func main() {
	wordlist := flag.String("wd", "dataW0rdL1st", "wordlist for items")
	userWordlist := flag.String("ud", "userW0rdL1st", "user list")
	logFilename := flag.String("log", "api.log", "filename where to store logs")
	port := flag.Int("port", 8080, "port on which to listen")
	iface := flag.String("iface", "0.0.0.0", "bind to specific interface")
	passFile := flag.String("pw", "passw0rdFile", "path to the password file")
	flag.Parse()

	passwordFile = *passFile

	tokens = make(map[string]struct{})

	if err := populate(*wordlist, *userWordlist); err != nil {
		log.Fatalf("Error populating data: %+v\n", err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("I have different versions"))
	})

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("OK!"))
	})

	api.HandleFunc("/list", list)
	api.HandleFunc("/list/{key}", getKey)

	api.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("You need to pass username and password using JSON"))
	}).Methods("GET")
	api.HandleFunc("/login", login).Methods("POST")

	api.HandleFunc("/n0tesOfABro", serveKey)

	api.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("What did you expect to find here????"))
	})

	api.HandleFunc("/today", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("Today is a beautiful day!"))
	})

	api.HandleFunc("/tool", func(w http.ResponseWriter, r *http.Request) {
		w.Write(message("If you have found me, you are on the right track!"))
	})

	api.HandleFunc("/users", listUsers)

	logFile, err := os.OpenFile(*logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening %s log filename: %+v\n", *logFilename, err)
	}
	defer logFile.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		message := fmt.Sprintf("%s - %s => %s\n", req.RemoteAddr, req.UserAgent(), req.RequestURI)
		logFile.WriteString(message)
		r.ServeHTTP(w, req)
	})

	address := fmt.Sprintf("%s:%d", *iface, *port)
	msg := fmt.Sprintf("Starting the server on port %d", *port)

	log.Print(msg)
	log.Fatal(http.ListenAndServe(address, handler))
}

type creds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.Write(message("I expected content type to be application/json"))
		return
	}

	var credsReceived creds
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000))
	if err != nil {
		log.Printf("Error occurred: %+v\n", err)
		return
	}
	defer r.Body.Close()

	json.Unmarshal(body, &credsReceived)

	if credsReceived.Username == correctUsername && credsReceived.Password == correctPassword {
		w.Header().Add("FIT-Token", generateToken())
		w.Write(message("Good job, correct login!"))
	} else {
		w.Write(message("Wrong username or password"))
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	var ret = struct {
		Keys []string `json:"keys"`
	}{
		Keys: keys,
	}
	b, _ := json.MarshalIndent(ret, "", "    ")
	w.Write(b)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	b, _ := json.MarshalIndent(users, "", "    ")
	w.Write(b)
}

// return the specific value for key passed inside mux.Vars
func getKey(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	key := params["key"]

	val := data[key]
	if _, ok := data[key]; !ok {
		val = "NOT FOUND"
	}

	if key == "flag" {
		authorized := r.Header.Get("FIT-Token")
		if authorized == "" {
			w.Write(message("You need to provide Fit-Token header"))
			return
		}

		lock.RLock()
		_, ok := tokens[authorized]
		defer lock.RUnlock()

		if !ok {
			w.Write(message("Token is not valid"))
			return
		}

		val += " ; Additionally, take a look at the /api/v1/n0tesOfABro"
	}

	var ret = struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   key,
		Value: val,
	}
	b, _ := json.MarshalIndent(ret, "", "    ")
	w.Write(b)
}

func populate(filename, usernameWordlist string) error {
	data = make(map[string]string)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		keys = append(keys, line[0])
		data[line[0]] = line[1]
	}

	userFile, err := os.Open(usernameWordlist)
	if err != nil {
		return err
	}
	defer userFile.Close()

	newScanner := bufio.NewScanner(userFile)
	for newScanner.Scan() {
		users = append(users, newScanner.Text())
	}

	return nil
}

func generateToken() string {
	lock.Lock()
	t := time.Now()
	timeString := t.String()
	hasher := sha1.New()
	hasher.Write([]byte(timeString))
	key := hex.EncodeToString(hasher.Sum(nil))
	tokens[key] = struct{}{}
	defer lock.Unlock()

	return key
}

func message(content string) []byte {
	formattedMsg := fmt.Sprintf("{\"message\":\"%s\"}", content)
	return []byte(formattedMsg)
}

func serveKey(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(passwordFile)
	if err != nil {
		http.Error(w, "Error reading pw data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(data))
}
