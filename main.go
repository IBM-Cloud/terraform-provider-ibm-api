package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	//"github.com/terrform-schematics-demo/terraform-provider-ibm-api/utils"
	"github.com/terraform-provider-ibm-api/utils"
	mgo "gopkg.in/mgo.v2"
)

func main() {

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	var port int
	flag.IntVar(&port, "p", 9080, "Port on which this server listens")
	flag.Parse()
	r := mux.NewRouter()

	r.HandleFunc("/configuration", utils.ConfHandler)
	r.HandleFunc("/configuration/{repo_name}", utils.ConfDeleteHandler)

	r.HandleFunc("/configuration/{repo_name}/plan", utils.PlanHandler(session)).Methods("POST")

	r.HandleFunc("/configuration/{repo_name}/show", utils.ShowHandler(session)).Methods("POST")

	r.HandleFunc("/configuration/{repo_name}/apply", utils.ApplyHandler(session)).Methods("POST")

	r.HandleFunc("/configuration/{repo_name}/destroy", utils.DestroyHandler(session)).Methods("POST")

	r.HandleFunc("/configuration/{repo_name}/{action}/{actionID}/log", utils.LogHandler)

	r.HandleFunc("/configuration/{repo_name}/{action}/{actionID}/status", utils.StatusHandler)

	r.HandleFunc("/configuration/{repo_name}/{action}/{log_file}", utils.ViewLogHandler)

	r.HandleFunc("/configuration/{repo_name}/{action}", utils.GetActionDetailsHandler(session)).Methods("GET")

	fmt.Println("Server will listen at port", port)
	muxWithMiddlewares := http.TimeoutHandler(r, time.Second*60, "Timeout!")
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), muxWithMiddlewares)
	if err != nil {
		fmt.Printf("Couldn't start the server %v", err)
	}
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")

	index := mgo.Index{
		Key:        []string{"actionid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}
