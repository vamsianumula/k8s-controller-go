package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var inMemoryStore map[uuid.UUID]string = make(map[uuid.UUID]string)

type VM struct {
	Id   uuid.UUID
	Name string
}

func addVM(w http.ResponseWriter, req *http.Request) {
	var vm VM
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&vm)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Println("Received request for vm: " + vm.Name)
	for k, v := range inMemoryStore {
		if v == vm.Name {
			fmt.Println("VM with this name already exists")
			w.WriteHeader(http.StatusConflict)
			jsonString, err := json.Marshal(k)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(jsonString)
			return
		}
	}
	vm.Id = uuid.New()
	jsonString, err := json.Marshal(vm.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inMemoryStore[vm.Id] = vm.Name
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonString)
}

func getStatus(w http.ResponseWriter, req *http.Request) {

	params := mux.Vars(req)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		fmt.Println("Could not get id from URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, ok := inMemoryStore[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cpuUsage := rand.Intn(100)
	jsonString, err := json.Marshal(cpuUsage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonString)
}

func getServers(w http.ResponseWriter, req *http.Request) {
	vmsList := make([]VM, 0)
	for k, v := range inMemoryStore {
		vm := VM{Id: k, Name: v}
		vmsList = append(vmsList, vm)
	}

	jsonString, err := json.Marshal(vmsList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonString)
}

func getServer(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		fmt.Println("Could not get id from URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, ok := inMemoryStore[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	vm := VM{Id: id, Name: inMemoryStore[id]}
	jsonString, err := json.Marshal(vm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonString)
}

func deleteServer(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		fmt.Println("Could not get id from URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, ok := inMemoryStore[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(inMemoryStore, id)
	w.WriteHeader(http.StatusNoContent)
}

func checkName(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	if _, ok := params["name"]; !ok {
		fmt.Println("Unable to get name from URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := params["name"]
	fmt.Println("Received request to check name: " + name)
	forbiddenNamesList := []string{
		"forbiddenName1",
		"forbiddenName2",
		"forbiddenName3",
	}

	for i := range forbiddenNamesList {
		if strings.EqualFold(name, forbiddenNamesList[i]) {
			fmt.Println("Name is forbidden")
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/servers/{id}/status", getStatus).Methods("GET")
	r.HandleFunc("/servers/{id}", getServer).Methods("GET")
	r.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE")
	r.HandleFunc("/servers", addVM).Methods("POST")
	r.HandleFunc("/servers", getServers).Methods("GET")
	r.HandleFunc("/check/{name}", checkName).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8090", nil)
}
