package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Linkfj struct {
	URL string `json:"url"`
}
type Status string

const (
	StatusInitialized Status = "initialized"
	StatusInProgress  Status = "in progress"
	StatusReady       Status = "Ready"
)

type Archive struct {
	Arch     bytes.Buffer
	Objs     int
	zpwrr    *zip.Writer
	ArchStat Status
	Link     string
}

var n int
var idcount int
var archs map[int]*Archive = make(map[int]*Archive, 8)

//for i:=0; i<3; i++{
//archs[i].Arch=bytes.newBuffer([]byte{})
//}

func main() {
	http.HandleFunc("/ini_arch", initArchive)
	http.HandleFunc("/add_obj", addObject)
	http.HandleFunc("/arch_stat", getArchStatus)
	http.HandleFunc("/archive/", downloadArch)

	log.Print("Server StuffLoad is listening on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func initArchive(w http.ResponseWriter, r *http.Request) {
	if n == 3 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("You already have 3 not formed archives, you can't have more ones at once. Server is busy."))
		return
	}
	idcount += 1
	n += 1
	archs[idcount] = &Archive{}
	buf := &archs[idcount].Arch
	archs[idcount].zpwrr = zip.NewWriter(buf)
	archs[idcount].ArchStat = StatusInitialized
	w.Write([]byte("The id of created task by formig a zip archive is: " + strconv.Itoa(idcount)))
	r.Body.Close()
}

func addObject(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	if r.Method != "PATCH" {
		http.Error(w, "You should do a PATCH request", http.StatusMethodNotAllowed)
		return
	}
	id, err := helpy1(w, r)
	if err != nil {
		log.Print(err)
		return
	}
	if archs[id].Objs == 3 {
		http.Error(w, "Archive with this id already contains 3 objects and its task was finished",
			http.StatusConflict)
		return
	}
	tmpl := strings.Split("http://"+r.Host+r.URL.Path, "/")
	archlink := strings.Join([]string{tmpl[0], tmpl[1], tmpl[2]}, "/") + "/archive/" + strconv.Itoa(id)
	link := Linkfj{}
	byts, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		return
	}
	if err := json.Unmarshal(byts, &link); err != nil {
		log.Print(err)
		http.Error(w, "Please create the body of the request in json format properly", http.StatusBadRequest)
		return
	}
	resp, err := http.Get(link.URL)
	if err != nil {
		log.Print(err)
		http.Error(w, `Perhaps, your URL from patch request is in a wrong
         format or server where the resource located, is not available`, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	contnt := resp.Header.Get("Content-Type")
	if !(contnt == "application/pdf" || contnt == "image/jpeg") {
		http.Error(w, "Supported for downloading only files of formats \".jpeg\" and \".pdf\"", http.StatusConflict)
		return
	}
	if !(resp.StatusCode == http.StatusNonAuthoritativeInfo || resp.StatusCode == http.StatusOK) {
		http.Error(w, "The resource is unavailable for downloading, but now at this archive left less space",
			http.StatusNotFound)
		archs[id].Objs += 1
		w.Write([]byte("The object added to archive with id:" + strconv.Itoa(id)))
		if archs[id].Objs == 3 {
			archs[id].ArchStat = StatusReady
			archs[id].Link = archlink
			n -= 1
			if archs[id].Arch.Len() == 0 {
				http.Error(w, `All 3 objects of archive are not available for download, this archive 
                is empty and will be deleted`,
					http.StatusCreated)
				delete(archs, id)
			}
			archs[id].zpwrr.Close()
		} else {
			archs[id].ArchStat = StatusInProgress
		}
		return
	}
	urlparts := strings.Split(link.URL, "/")
	filewrr, err := archs[id].zpwrr.Create(urlparts[len(urlparts)-1])
	if err != nil {
		log.Print(err)
		return
	}
	dtresp, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		log.Print(err2)
		return
	}
	_, err1 := filewrr.Write(dtresp)
	if err1 != nil {
		log.Print(err1)
		return
	}
	archs[id].Objs += 1
	w.Write([]byte("The object added to archive with id:" + strconv.Itoa(id)))
	if archs[id].Objs == 3 {
		n -= 1
		archs[id].ArchStat = StatusReady
		archs[id].Link = archlink
		archs[id].zpwrr.Close()
	} else {
		archs[id].ArchStat = StatusInProgress
	}
}

func getArchStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := helpy1(w, r)
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("The status of this task with id: %d by creating zip archive is %s",
		id, archs[id].ArchStat)))
	if archs[id].ArchStat == StatusReady {
		fmt.Fprintf(w, "The link to download the archive is %s", archs[id].Link)
	}

}

func helpy1(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "You made a mistake in transfering parametres", http.StatusBadRequest)
		return 0, errors.New("parameters are in a wrong format")
	}
	if amt := r.Form.Get("id"); amt != "" {
		idt, err := strconv.Atoi(amt)
		if err != nil {
			http.Error(w, "You made a mistake in transfering parametres", http.StatusBadRequest)
			return 0, errors.New("parameters are in a wrong format")
		}
		_, ok := archs[idt]
		if !ok {
			http.Error(w, "There are no task with such id", http.StatusNotFound)
			return 0, errors.New("not existing task id")
		}
		return idt, nil
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`You didn't specify the id of archive, which status you need to know, in parameters 
        of your request`))
		return 0, errors.New("parameters not found")
	}
}

func downloadArch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split("http://"+r.Host+r.URL.Path, "/")[4])
	if err != nil {
		log.Print(err)
		return
	}
	btdata, err1 := io.ReadAll(&archs[id].Arch)
	if err1 != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"archive%d.zip\"", id))
	w.Header().Set("Content-Length", strconv.Itoa(len(btdata)))
	w.Write(btdata)
}
