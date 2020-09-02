package main

import (
    "context"
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
        "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"encoding/hex"

)

//var client *mongo.Client

type Participant struct{
	Name      string        `json:"name" bson:"name"`
	Email_ID  string        `json:"email_id" bson:"email_id"`
	RSVP      string        `json:"rsvp" bson:"rsvp"`
}

type Meeting struct {
	//ID      primitive.ObjectID   `json:"_id" bson:"_id"`
	ID            string          `json:"id" bson:"id"`
	Title         string		  `json:"title" bson:"title"`
	Participants  []Participant   `json:"Participants" bson:"Participants"`
	Start_time    string          `json:"start_time" bson:"start_time"`
	End_time      string		  `json:"end_time" bson:"end_time"`	
	Timestamp     time.Time		  `json:"Timestamp" bson:"Timestamp"`
}

/*var CNX = Connection()

func Connection() *mongo.Client {
    
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(context.TODO(), nil)

    if err != nil {
    log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    return client
}
*/

var Meetings []Meeting 
var Meeting_by_Id Meeting
var Meetings_by_Time bson.M
var Meetings_Of_Participant bson.M

//Scheduling a meeting
func schedule_meetings(w http.ResponseWriter, r *http.Request) {

	if r.Method == POST{
		if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
		}
	}
	id, _ := randomHex(6)
	meetingTitle := r.FormValue("title")
	participant := r.FormValue("participant")
	email := r.FormValue("email")
	rsvp := r.FormValue("rsvp")
	startTime := r.FormValue("start")
	endTime := r.FormValue("end")

	if possible(email , startTime , endTime ) == true:
	insertDB(id, meetingTitle, participant, email, rsvp, startTime, endTime)
	else
	fmt.Fprintf(w, "Meeting not Possible")

	
}
//insert meeting details in database
func insertDB(id string, title string, name string, email string, rsvp string, start_time int, end_ime int) {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("db").Collection("meetings")
	//collection := db.CNX.Database("db").Collection("meetings")

	meet := Meeting{
		ID:    id,
		Title: title,
		Participants: []Participant{
			Participant{
				Name:  name,
				Email: email,
				RSVP:  rsvp}},
		Start_time:    start_time,
		End_time:      end_time,
		CreationTimestamp: time.Now()}

	insert, err := collection.InsertOne(ctx, meet)
	if err != nil {
		panic(err)
	}
	fmt.Println(insert.InsertedID, "inserted")

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

}
//generate id
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
//possibility of scheduling meeting
func possible(email_id string, start_time int) bool {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("db").Collection("meetings")

	cursor, err := collection.Find(ctx, bson.D{Start_time: {$gte: start_time, $lt: end_time}},RSVP: "yes", Participants: bson.D{Email: email_id}})

	if err != nil {
		fmt.Println(" error: ", err)
		defer cursor.Close(ctx)

	} else {
		for cursor.Next(ctx) {
			err := cursor.Decode(&Meetings_Of_Participant)
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			else{
				return false
			}
		}
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	return true
}

//return meeting when id is given
func meeting_from_id(w http.ResponseWriter, r *http.Request) {

	id :=r.URL.Path[len("/meeting/"):]

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("db").Collection("meetings")

	err = collection.FindOne(ctx, bson.D{ID: id}).Decode(Meeting_by_Id)
	if err != nil {
		fmt.Println("No meeting %v:", err)
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	json.NewEncoder(w).Encode(Meeting_by_Id)
}
//return meetings in a given time range
func meetings_during_time(w http.ResponseWriter, r *http.Request) {

	u, _ := url.Parse(r.URL.String())
	q, _ := url.ParseQuery(u.RawQuery)
	start_time, err := strconv.ParseInt(q.Get("start"), 10, 32)
	end_time, err := strconv.ParseInt(q.Get("end"), 10, 32)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("db").Collection("meetings")

	cursor, err := meetingCollection.Find(ctx, bson.D{{Start_time: {$gte: start_time, $lte: end_time}}, {End_time: {$gte: start_time, $lte: end_time}}}) 

	if err != nil {
		fmt.Println(" ERROR: %v", err)
		defer cursor.Close(ctx)
	} 
	else {
		for cursor.Next(ctx) {
			err := cursor.Decode(&Meetings_by_Time)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	json.NewEncoder(w).Encode(Meetings_by_Time)
}
//return meetings of a participant
func  meetings_of_participant(w http.ResponseWriter, r *http.Request) {

	email :=r.URL.Path[len("/meetings?participant="):]
	
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("db").Collection("meetings")

	cursor, err := collection.Find(ctx, bson.D{Email: email})

	if err != nil {
		fmt.Println(" error: %v", err)
		defer cursor.Close(ctx)

	} else {
		for cursor.Next(ctx) {
			err := cursor.Decode(&Meetings_Of_Participant)
			if err != nil {
				fmt.Println(" error:", err)
				os.Exit(1)
			}
		}
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	
	json.NewEncoder(w).Encode(Meetings_Of_Participant)
}

func main() {

    http.HandleFunc("/meetings",schedule_meetings)
	http.HandleFunc("/meeting/{meeting_id}", meeting_from_id)
	http.HandleFunc("/meetings?start={start_time}&end={end_time}", meetings_during_time)
	http.HandleFunc("/meetings?participant={email_id}", meetings_of_participant)
	
    http.ListenAndServe(":8080", nil)
}
