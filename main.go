package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	uuid "src/github.com/satori/go.uuid"
	"strings"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	c := getCookie(w, req)

	if req.Method == http.MethodPost {
		f, h, err := req.FormFile("nf")
		handleError(err)
		defer f.Close()

		//	Create SHA1 hash through filename
		ext := strings.Split(h.Filename, ".")[1]
		hash := sha1.New()
		io.Copy(hash, f)
		fname := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext

		// Create new file
		wd, err := os.Getwd()
		handleError(err)
		path := filepath.Join(wd, "public", fname)
		nf, err := os.Create(path)
		handleError(err)
		nf.Close()
		f.Seek(0, 0)
		io.Copy(nf, f)
		c = appendValue(w, c, fname)
	}

	xs := strings.Split(c.Value, "|")
	tpl.ExecuteTemplate(w, "index.gohtml", xs)
}

func getCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("session")

	if err != nil {
		sID, err := uuid.NewV4()
		handleError(err)

		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

		http.SetCookie(w, c)
	}
	return c
}

func appendValue(w http.ResponseWriter, c *http.Cookie, fname string) *http.Cookie {
	s := c.Value
	if !strings.Contains(s, fname) {
		s += "|" + fname
	}
	c.Value = s
	http.SetCookie(w, c)
	return c
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
		return
	}
}
