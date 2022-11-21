package do

import (
	"bytes"
	"encoding/json"
	// "errors"
	"log"
	"fmt"
	"net/http"
)

// "bytes"
// "fmt"
// "log"

// "log"
// "net/http"

const serverPort = "8090"

const reqUrl = "http://127.0.0.1:"+serverPort

func Create(name string) (string, error){
	values := map[string]string{"name": name}
    json_data, err := json.Marshal(values)

    if err != nil {
        log.Printf(err.Error())
    }

    resp, err := http.Post(reqUrl+"/servers", "application/json",bytes.NewBuffer(json_data))

    if err != nil {
        log.Printf(err.Error())
    }

    var res map[string]string

    json.NewDecoder(resp.Body).Decode(&res)

    fmt.Println(res)
	return res["id"], nil
}

func Delete(id string) (int,error){

    client := &http.Client{}

    req, err := http.NewRequest("DELETE", reqUrl+"/servers/"+"id",nil)
    if err != nil {
		fmt.Printf("Failed-1")
        fmt.Printf(err.Error())
        return 500, err
    }
    // Fetch Request
    resp, err := client.Do(req)
    if err != nil {
		fmt.Printf("Failed-2")
        fmt.Printf(err.Error())
        return 500, err
    }
   return resp.StatusCode, nil
}

func Check(name string) (int, error){
	resp, err := http.Get(reqUrl+"/check/"+name)
	if err!=nil{
		fmt.Printf("Checking name API failed with error: %s",err.Error())
		return 500, err
	}
	return resp.StatusCode, nil
}

func GetCpu(id string) (int32, error){
	resp, err := http.Get(reqUrl+"/servers/"+id+"/status")
	if err!=nil{
		fmt.Printf(err.Error())
		return -1, err
	}
	var res map[string]int

    json.NewDecoder(resp.Body).Decode(&res)

    // fmt.Println(res["cpu"])
	
	return int32(res["cpu"]), nil
}
