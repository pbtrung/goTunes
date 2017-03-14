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
	"time"

	"google.golang.org/api/drive/v2"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))

func main() {
	r := mux.NewRouter()
	store.MaxAge(86400 * 5)

	r.HandleFunc("/", home)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/logo", getLogo)
	r.HandleFunc("/lock-css", getLockCSS)

	r.Handle("/js", negroni.New(
		negroni.HandlerFunc(isAuthenticated),
		negroni.Wrap(http.HandlerFunc(getJS)),
	))
	r.Handle("/css", negroni.New(
		negroni.HandlerFunc(isAuthenticated),
		negroni.Wrap(http.HandlerFunc(getCSS)),
	))
	r.Handle("/gdrToken", negroni.New(
		negroni.HandlerFunc(isAuthenticated),
		negroni.Wrap(http.HandlerFunc(getGDRToken)),
	))
	r.Handle("/album-art-empty", negroni.New(
		negroni.HandlerFunc(isAuthenticated),
		negroni.Wrap(http.HandlerFunc(getEmptyAlbumArt)),
	))
	r.Handle("/cloudantAuth", negroni.New(
		negroni.HandlerFunc(isAuthenticated),
		negroni.Wrap(http.HandlerFunc(getCloudantAuth)),
	))

	var port string
	if os.Getenv("PORT") == "" {
		port = "3000"
	} else {
		port = os.Getenv("PORT")
	}
	log.Fatalln(http.ListenAndServe(":"+port, r))
}

func isAuthenticated(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session, err := store.Get(r, "auth0-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := session.Values["profile"]; !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		next(w, r)
	}
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
		serveFile(w, "./public/goTunes.html")
	}
}

func getGDRToken(w http.ResponseWriter, r *http.Request) {
	accessToken := getGDRAccessToken()
	type Token struct {
		AccessToken string
		Expiration  time.Time
	}
	token := &Token{AccessToken: accessToken.AccessToken, Expiration: accessToken.Expiry}
	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(b))
}

func getCloudantAuth(w http.ResponseWriter, r *http.Request) {
	type CloudantAuth struct {
		User     string
		Password string
	}
	auth := &CloudantAuth{User: os.Getenv("CLOUDANT_USER"), Password: os.Getenv("CLOUDANT_PASSWORD")}
	b, err := json.Marshal(auth)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(b))
}

func getJS(w http.ResponseWriter, r *http.Request) {
	serveFile(w, "./public/goTunes.js")
}

func getLogo(w http.ResponseWriter, r *http.Request) {
	serveFile(w, "./public/goTunes.png")
}

func getLockCSS(w http.ResponseWriter, r *http.Request) {
	serveFile(w, "./public/lock.css")
}

func getCSS(w http.ResponseWriter, r *http.Request) {
	serveFile(w, "./public/goTunes.css")
}

func getEmptyAlbumArt(w http.ResponseWriter, r *http.Request) {
	serveFile(w, "./public/album-art-empty.png")
}

func serveFile(w http.ResponseWriter, path string) {
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
		} else if strings.HasSuffix(path, ".jpg") {
			contentType = "image/jpg"
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

func getGDRAccessToken() *oauth2.Token {
	var secrets = "./data/tokens.json"
	if _, err := os.Stat(secrets); os.IsNotExist(err) {
		log.Fatalln(err)
	}
	file, err := os.Open(secrets)
	if err != nil {
		log.Fatalln(err)
	}
	tok := new(oauth2.Token)
	err = json.NewDecoder(file).Decode(tok)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	config, err := getConfig()
	if err != nil {
		log.Fatalln(err)
	}

	tokenSource := config.TokenSource(oauth2.NoContext, tok)
	newTok, err := tokenSource.Token()
	if err != nil {
		log.Fatalln(err)
	}
	if newTok.AccessToken != tok.AccessToken {
		saveToken(secrets, newTok)
		fmt.Println("Save new token: ", newTok.AccessToken)
	}

	return newTok
}
