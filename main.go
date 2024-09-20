package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/session"
	_ "github.com/BenFaruna/url-shortener/internal/session/providers/memory"
)

var globalSessions *session.Manager

func main() {
	mux := http.NewServeMux()

	mux.Handle("/session", http.HandlerFunc(Count))
	mux.Handle("/", controller.HomeHandler())
	mux.Handle("/api/v1/", APIMux())

	styles := http.FileServer(http.Dir("./static/css/"))
	mux.Handle("/styles/", http.StripPrefix("/styles/", styles))

	script := http.FileServer(http.Dir("./static/js/"))
	mux.Handle("/scripts/", http.StripPrefix("/scripts/", script))

	if err := http.ListenAndServe(":8000", controller.IncomingRequest(mux)); err != nil {
		panic(err)
	}
}

func APIMux() http.Handler {
	shortenerMux := http.NewServeMux()

	shortenerMux.Handle("/shorten", controller.ShortenHandler(controller.GenerateShortString))
	shortenerMux.Handle("/address/", controller.GetFullAddressHandler())

	return http.StripPrefix("/api/v1", shortenerMux)
}

func Count(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	createtime := sess.Get("createtime")
	if createtime == nil {
		sess.Set("createtime", time.Now().Unix())
	} else if (createtime.(int64) + 360) < (time.Now().Unix()) {
		globalSessions.SessionDestroy(w, r)
		sess = globalSessions.SessionStart(w, r)
	}
	ct := sess.Get("countnum")
	if ct == nil {
		sess.Set("countnum", 1)
	} else {
		sess.Set("countnum", (ct.(int) + 1))
	}
	// t, _ := template.ParseFiles("login.html")
	// w.Header().Set("Content-Type", "text/html")
	// t.Execute(w, sess.Get("countnum"))
	fmt.Fprintf(w, "%d count...", sess.Get("countnum"))
}

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}
