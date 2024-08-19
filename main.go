package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
)

// func notFound(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "public/404.gohtml")
// }

func main() {
	mux := http.NewServeMux()

	//static files
	staticHandler := http.FileServer(http.Dir("static"))
	mux.HandleFunc("GET /static*", http.StripPrefix("/static", staticHandler).ServeHTTP)

	//homepage
	mux.HandleFunc("GET /", StaticHandler(template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))))

	//about
	mux.HandleFunc("GET /about", StaticHandler(template.Must(template.ParseFiles("templates/base.gohtml", "templates/about.gohtml"))))

	//posts
	postTemplate := template.Must(template.ParseFiles("post.gohtml"))
	mux.HandleFunc("GET /posts/{slug}", PostHandler(FileReader{}, postTemplate))

	//404
	// todo

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

type Post struct {
	Content template.HTML
	Title   string `toml:"title"`
	Slug    string `toml:"slug"`
	Author  Author `toml:"author"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

func PostHandler(sl SlugReader, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var post Post
		rest, err := frontmatter.Parse(strings.NewReader(postMarkdown), &post)
		if err != nil {
			http.Error(w, "Error parsing frontmatter", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		err = goldmark.Convert(rest, &buf)
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
		post.Content = template.HTML(buf.String())

		err = tpl.Execute(w, post)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

func StaticHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.Execute(w, r)
		if err != nil {
			http.Error(w, "Error executing static template", http.StatusInternalServerError)
			return
		}
	}
}
