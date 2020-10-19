package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/project/helper"
	"github.com/project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var collection = helper.ConnectToMongoDB()

func main() {

	http.HandleFunc("/meeting/", getMeetingByID)
	http.HandleFunc("/meeting", createMeeting)
	http.HandleFunc("/meetings", getMeetings)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getMeetingByID(w http.ResponseWriter, r *http.Request) {

	var meetingID string = r.URL.Path[9:]

	var meeting models.Meeting

	// set header.
	w.Header().Set("Content-Type", "application/json")

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(meetingID)

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := collection.FindOne(context.TODO(), filter).Decode(&meeting)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(meeting)
}

func getMeetings(w http.ResponseWriter, r *http.Request) {
	_, emailOK := r.URL.Query()["participant"]
	_, startOK := r.URL.Query()["start"]
	_, endOK := r.URL.Query()["end"]

	if emailOK {
		getMeetingsByParticipant(w, r)
	} else if startOK && endOK {
		getMeetingsByTime(w, r)
	} else {
		log.Println("Url Param 'key' is missing")
	}

}

func getMeetingsByParticipant(w http.ResponseWriter, r *http.Request) {

	emails, ok := r.URL.Query()["participant"]

	if !ok || len(emails[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	email := emails[0]

	log.Println("Url Param 'key' is: " + string(email))

	var meetings []models.Meeting

	// set header.
	w.Header().Set("Content-Type", "application/json")

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var meeting models.Meeting
		if err = cursor.Decode(&meeting); err != nil {
			log.Fatal(err)
		}

		for _, participant := range meeting.Participants {
			if participant.Email == email {
				meetings = append(meetings, meeting)
			}
		}

	}

	json.NewEncoder(w).Encode(meetings)
}

func getMeetingsByTime(w http.ResponseWriter, r *http.Request) {

	startTimeList, startOK := r.URL.Query()["start"]
	endTimeList, endOK := r.URL.Query()["end"]

	if !startOK || len(startTimeList[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	if !endOK || len(endTimeList[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	startTime := startTimeList[0]
	endTime := endTimeList[0]

	log.Println("Url Param 'key' is: " + string(startTime))
	log.Println("Url Param 'key' is: " + string(endTime))

	var meetings []models.Meeting

	// set header.
	w.Header().Set("Content-Type", "application/json")

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var meeting models.Meeting
		if err = cursor.Decode(&meeting); err != nil {
			log.Fatal(err)
		}
		if checkOverlapTime(meeting, startTime, endTime) {
			meetings = append(meetings, meeting)
		}
	}

	json.NewEncoder(w).Encode(meetings)
}

func createMeeting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var meeting models.Meeting

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&meeting)

	//add timestamp
	t := time.Now()
	meeting.CreationTimeStamp = t.String()

	if !checkRSVPOverlap(meeting) {

		// insert our book model.
		result, err := collection.InsertOne(context.TODO(), meeting)

		if err != nil {
			helper.GetError(err, w)
			return
		}

		json.NewEncoder(w).Encode(result)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[9:])
}

func checkRSVPOverlap(meetingGiven models.Meeting) bool {
	for _, participant := range meetingGiven.Participants {
		var meetings []models.Meeting = getAcceptedMeetingsArrayByParticipant(participant)

		for _, meeting := range meetings {
			if checkOverlapTime(meeting, meetingGiven.StartTime, meetingGiven.EndTime) {
				return true
			}

		}
	}
	return false
}

func checkOverlapTime(meeting models.Meeting, startTime string, endTime string) bool {
	hs, _ := strconv.Atoi(startTime[0:2])
	he, _ := strconv.Atoi(endTime[0:2])
	ms, _ := strconv.Atoi(startTime[len(startTime)-2:])
	me, _ := strconv.Atoi(endTime[len(endTime)-2:])

	mms, _ := strconv.Atoi(meeting.StartTime[len(meeting.StartTime)-2:])
	mme, _ := strconv.Atoi(meeting.EndTime[len(meeting.EndTime)-2:])
	mhs, _ := strconv.Atoi(meeting.StartTime[0:2])
	mhe, _ := strconv.Atoi(meeting.EndTime[0:2])

	if mhe < hs || mhs > he {
		return false
	} else if mhe == hs && mme <= ms {
		return false
	} else if mhs == he && mms >= me {
		return false
	} else {
		return true
	}
}

func getMeetingsArrayByParticipant(participantCheck models.Participant) []models.Meeting {

	var meetings []models.Meeting

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var meeting models.Meeting
		if err = cursor.Decode(&meeting); err != nil {
			log.Fatal(err)
		}

		for _, participant := range meeting.Participants {
			if participant.Email == participantCheck.Email {
				meetings = append(meetings, meeting)
			}
		}

	}

	return meetings
}

func getAcceptedMeetingsArrayByParticipant(participantCheck models.Participant) []models.Meeting {

	var meetings []models.Meeting

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var meeting models.Meeting
		if err = cursor.Decode(&meeting); err != nil {
			log.Fatal(err)
		}

		for _, participant := range meeting.Participants {
			if participant.Email == participantCheck.Email && participant.Rsvp == "Yes" {
				meetings = append(meetings, meeting)
			}
		}

	}

	return meetings
}
