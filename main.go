package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}

var tpl = template.Must(template.ParseFiles("index.html"))

// Serves Home Path
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

type ResponseData struct {
	Success bool `json:"success"`
	Results int  `json:"results"`
	Rows    []struct {
		Symbol           string `json:"Symbol"`
		CompanyName      string `json:"CompanyName"`
		ISIN             string `json:"ISIN"`
		Ind              string `json:"Ind"`
		Purpose          string `json:"Purpose"`
		BoardMeetingDate string `json:"BoardMeetingDate"`
		DisplayDate      string `json:"DisplayDate"`
		SeqID            string `json:"seqId"`
		Details          string `json:"Details"`
	} `json:"rows"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	// fmt.Println(u)

	fmt.Println("Search Query is: ", searchKey)

	client := &http.Client{}
	url := "https://www1.nseindia.com/corporates/corpInfo/equities/getBoardMeetings.jsp?"
	// url := "https://api.github.com/repos/dotcloud/docker"
	req, err := http.NewRequest("GET", url, nil)

	// NSE India website required these headers
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")

	q := req.URL.Query() // Get a copy of the query values.
	q.Add("symbol", searchKey)
	q.Add("period", "Latest Announced")
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.

	// client.Do(req)
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	str := strings.TrimSpace(string(body))

	str = strings.ReplaceAll(str, "success:", "\"success\":")
	str = strings.ReplaceAll(str, "results:", "\"results\":")
	str = strings.ReplaceAll(str, "rows:", "\"rows\":")
	str = strings.ReplaceAll(str, "Symbol:", "\"Symbol\":")
	str = strings.ReplaceAll(str, "CompanyName:", "\"CompanyName\":")
	str = strings.ReplaceAll(str, "ISIN:", "\"ISIN\":")
	str = strings.ReplaceAll(str, "Ind:", "\"Ind\":")
	str = strings.ReplaceAll(str, "Purpose:", "\"Purpose\":")
	str = strings.ReplaceAll(str, "BoardMeetingDate:", "\"BoardMeetingDate\":")
	str = strings.ReplaceAll(str, "DisplayDate:", "\"DisplayDate\":")
	str = strings.ReplaceAll(str, "seqId:", "\"seqId\":")
	str = strings.ReplaceAll(str, "Details:", "\"Details\":")

	rd := ResponseData{}
	err = json.NewDecoder(strings.NewReader(str)).Decode(&rd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// fmt.Println(rd.Rows[0])

	tpl.Execute(w, rd.Rows[0])
}
