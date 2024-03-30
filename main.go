package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /posts/{slug}", PostHandler(FileReader{}))

	fmt.Println("server running on " + "http://localhost:3030")
	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

func (fsr FileReader) Read(slug string) (string, error) {
	f, err := os.Open(slug + ".md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func PostHandler(sl SlugReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		var buf bytes.Buffer
		err = goldmark.Convert([]byte(postMarkdown), &buf)
		if err != nil {
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}
		// TODO: parse template once and not on every page load
		tpl, err := template.ParseFiles("post.gohtml")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		// TODO: stop hardcoding post data. parse from frontmatter
		err = tpl.Execute(w, PostData{
			Title:   "Title",
			Content: template.HTML(buf.String()),
			Author:  "Me",
		})
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
		}
		// io.Copy(w, &buf)
	}
}

type PostData struct {
	Title   string
	Content template.HTML
	Author  string
}
