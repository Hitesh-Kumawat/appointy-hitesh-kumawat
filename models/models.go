package models

import "go.mongodb.org/mongo-driver/bson/primitive"

//Meeting create struct
type Meeting struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title             string             `json:"title,omitempty" bson:"title,omitempty"`
	Participants      []Participant      `json:"participants" bson:"participants,omitempty"`
	StartTime         string             `json:"startTime,omitempty" bson:"startTime,omitempty"`
	EndTime           string             `json:"endTime,omitempty" bson:"endTime,omitempty"`
	CreationTimeStamp string             `json:"creationTimeStamp,omitempty" bson:"creationTimeStamp,omitempty"`
}

//Participant create struct
type Participant struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	Rsvp  string `json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}
