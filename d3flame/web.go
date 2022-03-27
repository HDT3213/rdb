package d3flame

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

var flameTmplData = &struct {
	D3Css        template.CSS
	D3Js         template.JS
	D3Flame      template.JS
	D3Tip        template.JS
	BootstrapCss template.CSS
}{
	D3Css:        template.CSS(d3Css),
	D3Js:         template.JS(d3Js),
	D3Flame:      template.JS(d3FlameGraphJs),
	D3Tip:        template.JS(d3TipJs),
	BootstrapCss: template.CSS(bootstrapCss),
}

func flamegraph(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("flamegraph").Parse(html))
	err := tmpl.Execute(w, flameTmplData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Internal Error"))
	}
}

// FlameItem is an Element in flamegraph
type FlameItem struct {
	Name     string   `json:"n"`
	Value    int      `json:"v"`
	Children children `json:"c,omitempty"`
}

type children map[string]*FlameItem

func (ch children) MarshalJSON() ([]byte, error) {
	list := make([]*FlameItem, 0, len(ch))
	for _, v := range ch {
		list = append(list, v)
	}
	return json.Marshal(list)
}

// Web starts a web server to render flamegraph
func Web(data []byte, port int) chan<- struct{} {
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
	}
	http.HandleFunc("/flamegraph", flamegraph)
	http.HandleFunc("/stacks.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	fmt.Printf("see http://localhost:%d/flamegraph\n", port)
	stop := make(chan struct{})
	go func() {
		<-stop
		_ = server.Close()
	}()
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return stop
}
