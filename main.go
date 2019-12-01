package main

import (
    "html/template"
	"net/http"
	"fmt"
    "log"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"encoding/json"
)

type BoardMeeting struct { 
	Name string
	Date    string    
	Purpose  string   
	Details string   
}

type BoardMeetings []BoardMeeting

func main(){
	http.HandleFunc("/", landingHandler)
	http.HandleFunc("/getMeetingDetails", getMeetingDetails)
	http.HandleFunc("/getSuggestions", getSuggestions)
	http.ListenAndServe(":8080",nil)
}

type Names []Name

func getInitialValues() BoardMeetings{
	//meeting := BoardMeetings{{"SBIN","22-May-2015" ,"Results/Dividend" ,"To approve the Audited financial results"},{"SBIN","19-May-2009" ,"Results/Dividend" ,"To approve the something somethings"}}
	meeting := BoardMeetings{}
	return meeting
}

func getMeetingValues(companyName string) BoardMeetings{
	//scraping https://www.nseindia.com to get the list of meeting for the requested company
	resp, err := http.Get("https://www.nseindia.com/marketinfo/companyTracker/boardMeeting.jsp?symbol=" + companyName)
    if err != nil{
		fmt.Println("error getting the meeting list")
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200{
		fmt.Println("error getting the meeting list 2")
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil{
		fmt.Println("error parsing the body for meeting list")
    }
	
	//variable to hold the links for the board meeting details 
    var links []string
    doc.Find("td").Each(func(i int, s *goquery.Selection){
        href, has_attr := s.Find("a").First().Attr("onclick")
        if has_attr{
			jsRemoved := strings.ReplaceAll(href,"javascript:window.open(","")
			convertedToObject := strings.Split(jsRemoved, "'")
			meetingLinks := "https://www.nseindia.com" + convertedToObject[1]
			links = append(links, meetingLinks)
        }
	})

	//meeting holds the value for the board meeting details 
	meeting := BoardMeetings { }
	for i := 0; i < len(links); i++ {
		boardMeetingDate := strings.ReplaceAll(strings.Split(links[i], "&")[1],"bmdate=","")
		resp, err := http.Get(links[i])
        if err != nil{
        	fmt.Println("Error coming in getting links : ",err)
    	}
   		defer resp.Body.Close()

    	if resp.StatusCode != 200{
        	fmt.Println("Status code error in getting meeting details ", resp.StatusCode, resp.Status)
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil{
			fmt.Println("Error coming in parsing : ",err)
		}

		var meetingDetails []string
		doc.Find("tr td .t0").Each(func(i int, s *goquery.Selection){
			meetingDetails = append(meetingDetails, strings.TrimSpace(s.Text()))
		})

		//meeting = BoardMeetings { BoardMeeting{boardMeetingDate,meetingDetails[1],meetingDetails[3]},}
		meeting = append(meeting, BoardMeeting{companyName,boardMeetingDate,meetingDetails[1],meetingDetails[len(meetingDetails)-1]})
		fmt.Println("=============================")
		}
		fmt.Println(meeting)
		return meeting

}

func landingHandler(w http.ResponseWriter, r *http.Request){
	meeting := getInitialValues()
	template, _ := 	template.ParseFiles("index.html")
	template.Execute(w,meeting)
	fmt.Println(meeting)
}

func getMeetingDetails(w http.ResponseWriter, r *http.Request){
	fmt.Println("Reached here")
	
	companyNames, ok := r.URL.Query()["companyName"]
    
    if !ok || len(companyNames[0]) < 1 {
        log.Println("Url Param 'companyName' is missing")
        return
    }

    // Query()["companyName"] will return an array of items, 
    // we only want the single item.
	companyName := companyNames[0]
	fmt.Println(companyName)
	
	//companyName := "SBIN"
	meeting := getMeetingValues(companyName)
	template, _ := 	template.ParseFiles("index.html")
	template.Execute(w,meeting)
	fmt.Println(meeting)
}



type Name struct {
	FName string
	LName string
}

func getSuggestions(w http.ResponseWriter, r *http.Request) {
	queries, ok := r.URL.Query()["query"]
    
    if !ok || len(queries[0]) < 1 {
        log.Println("Url Param 'query' is missing")
        return
    }

    // Query()["query"] will return an array of items, 
    // we only want the single item.
	query := queries[0]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	url := "https://www.nseindia.com/corporates/common/getCompanyListMktTracker.jsp"

	payload := strings.NewReader("query=" + query)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	

	res, _ := http.DefaultClient.Do(req)

	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(string(body))
	s, _ := getResponse([]byte(body))

	//fmt.Println(s.Rows1)
	var options []string
	for i :=0 ; i < len(s.Rows1); i++{
		options = append(options,s.Rows1[i].CompanyValues)
		options = append(options,",")
	}
	fmt.Fprintf(w, "%+v", options)
}


type rows struct{
	CompanyValues string `json:"CompanyValues"`
	CompanyNames string `json:"CompanyNames"`
}

type response struct{
	Rows1 []rows `json:"rows1"`
	Success string `json:"success"`
	Results int `json:"results"`
}

func getResponse(body []byte) (*response, error) {
    var s = new(response)
    err := json.Unmarshal(body, &s)
    if(err != nil){
        fmt.Println("whoops:", err)
    }
    return s, err
}