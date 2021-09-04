package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//ResultToSlack will send result to slack
func ResultToSlack(outURL, errURL, action, randomID, status, webhook string) {

	m := ComposeSlackMessage(outURL, errURL, action, randomID, status)
	m.PostToSlack(webhook)

}

//UpdateMongodb updates the status of the action.
func UpdateMongodb(s *mgo.Session, actionID string, status string) error {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")
	err := c.Update(bson.M{"actionid": actionID}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return err
	}

	return nil
}

//InsertMongodb updates the status of the action.
func InsertMongodb(s *mgo.Session, actionResponse ActionResponse) {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")
	err := c.Insert(actionResponse)
	if err != nil {
		if mgo.IsDup(err) {
			return
		}
		log.Println("Failed insert action details : ", err)
		return
	}
}

func Filepathjoin(dirPath string, pathElements ...string) (string, error) {
	log.Println("p1", pathElements)
	p := filepath.Join(append([]string{dirPath}, pathElements...)...)
	log.Println("p2", p, pathElements)
	p = filepath.FromSlash(p)

	log.Println("p3", p)
	log.Println("p4", p, dirPath)
	if !strings.HasPrefix(p, dirPath) {
		err := fmt.Errorf("path = %q, should be relative to %q", p, dirPath)
		return p, err
	}
	return p, nil
}
