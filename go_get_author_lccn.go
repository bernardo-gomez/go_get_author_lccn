package main

import(
  "fmt"
  "net/http"
  "time"
  "net/url"
  "log"
  "os"
  "io/ioutil"
  "encoding/xml"
  "regexp"
)
  type Root struct {
      XMLName xml.Name `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
      Work Workt `xml:"http://id.loc.gov/ontologies/bibframe/ Work"`
  }
  type  Ccontributiont struct{
     Contribution Contributiont `xml:"http://id.loc.gov/ontologies/bibframe/ Contribution"`
  }

  type Contributiont struct{
     Contrib_type Typet `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# type"`
     Ch_agent Ch_agentt `xml:"http://id.loc.gov/ontologies/bibframe/ agent"`

  }
  type Typet struct{
     Resource string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# resource,attr"`
  }

  type  Ch_agentt struct {
       Ch_Agent Ch_Agentt `xml:"http://id.loc.gov/ontologies/bibframe/ Agent"`
  }
  type  Ch_Agentt struct {
       LCCN string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
  }
  type Workt struct{
     XMLName xml.Name `xml:"http://id.loc.gov/ontologies/bibframe/ Work"`
     About string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
//  <bf:contribution>
     Ccontributions []Ccontributiont `xml:"http://id.loc.gov/ontologies/bibframe/ contribution"`
  }

// worldcat_identities_link is a function that performs an API request to
// retrieve the BIBFRAME record for a given ALMA mms_id.
// It parses the BIBFRAME RDF object to extract the author's LCCN.
// The author is identified as the "PrimaryContribution" property.
// The LCCN is used to form the worldcat identities link (WCI).
// worldcat identities link base uri is https://worldcat.org/identities/lccn- 
// this function returns the WCI link and an integer (return code).


func worldcat_identities_link(doc_id string ,linked_data_host string) (string , int) {
    link:=""
    outcome:=1
    worldcat_uri := "https://worldcat.org/identities/lccn-"

//  we send error messages to stderr via the logger package.

    l := log.New(os.Stderr, "go_get_author_lccn ",  1|2)

// connection timeout is 10 seconds.
    client := &http.Client{
         Timeout: 10* time.Second,
    }
    myURL:=linked_data_host+doc_id

    // l.Println(myURL)
// API request to get BIBFRAME record
    resp, err := client.Get(myURL)

    if err != nil {
         if _, ok := err.(*url.Error); ok {
             l.Println(err.(*url.Error))
             return link,outcome
         }
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
      l.Println("Unknown HTTP response")
      return link,outcome
    }

// read API response an save it as an array of bytes.
    body, err := ioutil.ReadAll(resp.Body)


// parse XML string 

    r:=Root{}
    err = xml.Unmarshal(body, &r)
    if err != nil {
       l.Println("xml parsing error:",err)
       return link,outcome
    }

//  extract LCCN 
    re := regexp.MustCompile(`http://id.loc.gov/authorities/names/(.*)`)

    list:=r.Work.Ccontributions
    if len(list) > 0{
      for i:=0; i < len(list); i++{
         if list[i].Contribution.Contrib_type.Resource == "http://id.loc.gov/ontologies/bflc/PrimaryContribution" {
            group:=re.FindStringSubmatch(list[i].Contribution.Ch_agent.Ch_Agent.LCCN)
            if len(group) == 2{
                worldcat_link:=worldcat_uri+group[1]
//                fmt.Printf("%q\n",worldcat_link)
                return worldcat_link,0
            }
         }
      }
    }
    return link,outcome
}

func main(){

   //  read ALMA mms_id from the command line
   if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "usage: %s mms_id\n",os.Args[0]);
        return
   }
   doc_id:= os.Args[1]
   
    linked_data_host:="https://open-na.hosted.exlibrisgroup.com/alma/01GALI_EMORY/bf/entity/instance/"
    identities_link,outcome := worldcat_identities_link(doc_id,linked_data_host)
    if outcome == 0{
      fmt.Printf("%q\n",identities_link)
    }
}
