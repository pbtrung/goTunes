package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v2"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
)

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/js", getJS)
	r.HandleFunc("/css", getCSS)
	r.HandleFunc("/album-art-empty", getEmptyAlbumArt)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/itemURL", getItemURL)

	var port string

	if os.Getenv("PORT") == "" {
		port = "3000"
	} else {
		port = os.Getenv("PORT")
	}
	log.Fatalln(http.ListenAndServe(":"+port, r))
}

func home(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth0-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if session.Values["profile"] == nil {
		cwd, _ := os.Getwd()
		t, err := template.ParseFiles(filepath.Join(cwd, "./public/login.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = godotenv.Load()
		if err != nil {
			log.Fatalln(err)
		}
		auth0 := map[string]string{"AUTH0_DOMAIN": os.Getenv("AUTH0_DOMAIN"), "AUTH0_CLIENT_ID": os.Getenv("AUTH0_CLIENT_ID"),
			"AUTH0_CALLBACK_URL": os.Getenv("AUTH0_CALLBACK_URL"), "AUTH0_CLIENT_SECRET": os.Getenv("AUTH0_CLIENT_SECRET")}
		err = t.Execute(w, auth0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		serveFile("./public/goTunes.html", w)
	}
}

func getJS(w http.ResponseWriter, r *http.Request) {
	getResource(w, r, "./public/goTunes.js")
}

func getCSS(w http.ResponseWriter, r *http.Request) {
	getResource(w, r, "./public/goTunes.css")
}

func getEmptyAlbumArt(w http.ResponseWriter, r *http.Request) {
	getResource(w, r, "./public/album-art-empty.png")
}

func getResource(w http.ResponseWriter, r *http.Request, path string) {
	session, err := store.Get(r, "auth0-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if session.Values["profile"] == nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Restricted contents!"))
		return
	} else {
		serveFile(path, w)
	}
}

func serveFile(path string, w http.ResponseWriter) {

	data, err := ioutil.ReadFile(path)

	if err == nil {
		var contentType string

		if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		} else {
			contentType = "text/plain"
		}

		w.Header().Add("Content-Type", contentType)
		w.Write(data)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("404 " + http.StatusText(404)))
	}
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	domain := os.Getenv("AUTH0_DOMAIN")

	conf := &oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + domain + "/authorize",
			TokenURL: "https://" + domain + "/oauth/token",
		},
	}

	code := r.URL.Query().Get("code")

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Getting userInfo
	client := conf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://" + domain + "/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var profile map[string]interface{}
	if err = json.Unmarshal(raw, &profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "auth0-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gob.Register(map[string]interface{}{})
	session.Values["id_token"] = token.Extra("id_token")
	session.Values["access_token"] = token.AccessToken
	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getItemURL(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "auth0-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if session.Values["profile"] == nil {
		serveFile("./public/login.html", w)
	} else {
		var secrets = "./data/tokens.json"

		if _, err := os.Stat(secrets); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Need Goole tokens!"))
			return
		}
		queryValues := r.URL.Query()
		if queryValues.Get("id") == "" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Need item ID!"))
			return
		}
		var fileID = queryValues.Get("id")

		file, err := os.Open(secrets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tok := new(oauth2.Token)
		err = json.NewDecoder(file).Decode(tok)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		ctx := context.Background()

		config, err := getConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tokenSource := config.TokenSource(oauth2.NoContext, tok)
		newTok, err := tokenSource.Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if newTok.AccessToken != tok.AccessToken {
			saveToken(secrets, newTok)
			fmt.Println("Save new token: ", newTok.AccessToken)
		}

		client := config.Client(ctx, tok)
		srv, err := drive.New(client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		driveFile, err := srv.Files.Get(fileID).Do()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var itemURL = driveFile.DownloadUrl + "&access_token=" + newTok.AccessToken
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", itemURL)
	}
}

func getConfig() (*oauth2.Config, error) {
	var configInput = "./data/config.json"
	file, err := ioutil.ReadFile(configInput)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	config, err := google.ConfigFromJSON(file, drive.DriveReadonlyScope)
	return config, err
}

func saveToken(path string, tok *oauth2.Token) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(tok)
}
