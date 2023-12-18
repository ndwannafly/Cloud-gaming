package app

import (
	"encoding/json"
	"log"
	"net/http"
    "fmt"
    // "io/ioutil"
    
	"coordinator/app/api/response"
    "gorm.io/gorm"
)

type App struct {
    Id          string  `json:"id" gorm:"column:id;type:VARCHAR(255)"`
    Name        string  `json:"name" gorm:"column:name;type:VARCHAR(255)"`
    Type        string  `json:"type" gorm:"column:type;type:VARCHAR(50)"`
    PosterUrl   string  `json:"posterURL" gorm:"column:poster_url;type:VARCHAR(255)"`
    Device      string  `json:"device" gorm:"column:device;type:varchar(255)"`
}

type GetAppListResponse struct {
	Apps []*App `json:"apps"`
}

func GetAppList(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
    var appList []*App
    
    result := db.
        Model(&App{}).
        Find(&appList)
    
    if result.Error != nil {
        fmt.Println("Can't get app list from DB: %s", result.Error);
        return
    
    }

	resp := response.Response{
		Data: GetAppListResponse{Apps: appList},
	}
	
	deviceParams, ok := r.URL.Query()["device"]
	if ok && len(deviceParams[0]) > 0 {
		device := deviceParams[0]

		var filteredAppList []*App
		for _, app := range appList {
			if app.Device == device {
				filteredAppList = append(filteredAppList, app)
			}
		}

		resp = response.Response{
			Data: GetAppListResponse{Apps: filteredAppList},
		}   
	}

	jsonResp, err := json.Marshal(resp)
    if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Couldn't marshall get app list response to JSON", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
