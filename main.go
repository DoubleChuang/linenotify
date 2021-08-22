// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"html/template"

	"github.com/DoubleChuang/linenotify/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var clientID string
var clientSecret string
var callbackURL string
var db *mongo.Mongo

func main() {
	var err error

	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/auth", authHandler)
	clientID = os.Getenv("ClientID")
	clientSecret = os.Getenv("ClientSecret")
	callbackURL = os.Getenv("CallbackURL")

	dbUser := os.Getenv("DatabaseUser")
	dbPassword := os.Getenv("DatabasePassword")
	dbHost := os.Getenv("DatabaseHost")

	port := os.Getenv("PORT")
	fmt.Printf("ENV port:%s, cid:%s csecret:%s\n", port, clientID, clientSecret)
	dbUrl := fmt.Sprintf("mongodb+srv://%s:%s@%s/?authSource=admin", dbUser, dbPassword, dbHost)

	db, err = mongo.New(dbUrl, "stock")
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // Populates request.Form
	code := r.Form.Get("code")
	state := r.Form.Get("state")
	fmt.Printf("Get code=%s, state=%s \n", code, state)

	data := url.Values{}
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", callbackURL)
	data.Add("client_id", clientID)
	data.Add("client_secret", clientSecret)

	byt, err := apiCall("POST", apiToken, data, "")
	fmt.Println("ret:", string(byt), " err:", err)

	res := newTokenResponse(byt)
	fmt.Println("result:", res)
	if res.Status != strconv.Itoa(http.StatusOK) {
		w.Write(byt)
		return
	}

	token := res.AccessToken
	e := db.InsertOne(
		bson.M{"token": token,
			"created_at": time.Now(),
		},
		"line_tokens")
	if e != nil {
		fmt.Println("mongo err:", e)
	}

	w.Write(byt)
}
func authHandler(w http.ResponseWriter, r *http.Request) {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(authTmpl)
	check(err)
	noItems := struct {
		ClientID    string
		CallbackURL string
	}{
		ClientID:    clientID,
		CallbackURL: callbackURL,
	}

	err = t.Execute(w, noItems)
	check(err)
}
