package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"coordinator/app/api/response"
	"coordinator/app/client"

	"github.com/redis/go-redis/v9"
)

type Provider struct {
	ID         string  `json:"id"`
	HostName   string  `json:"hostName"`
	Platform   string  `json:"platform"`
	CpuName    string  `json:"cpuName"`
	CpuNum     int     `json:"cpuNum"`
	MemSize    float64 `json:"memSize"`
	CpuPercent float64 `json:"cpuPercent"`
	MemPercent float64 `json:"memPercent"`
    ActivePlayer int64    `json:"activePlayer"`
}

type GetProviderListResp struct {
	Providers []*Provider `json:"providers"`
}

func GetProviderList(ctx context.Context, redisClient *redis.Client, hub *client.Hub, w http.ResponseWriter, r *http.Request) {
	hasOwnerIDParam := r.URL.Query().Has("owner")
	ownerID := r.URL.Query().Get("owner")

	providers := make([]*Provider, 0)

	for _, p := range hub.GetProviders() {
		if !hasOwnerIDParam || p.Provider.OwnerID == ownerID {
            activePlayer, err := redisClient.Get(ctx, p.Provider.OwnerID).Int64()
            if err != nil {
                fmt.Println("err:", err)
                fmt.Printf("Cannot get number of active player on server %s, set to 0 as default\n", ownerID)
                activePlayer = 0;
            }
        
			providers = append(providers, &Provider{
				ID:         p.ID,
				HostName:   p.Provider.HostName,
				Platform:   p.Provider.Platform,
				CpuName:    p.Provider.CpuName,
				CpuNum:     p.Provider.CpuNum,
				MemSize:    p.Provider.MemSize,
				CpuPercent: p.Provider.CpuPercent,
				MemPercent: p.Provider.MemPercent,
                ActivePlayer: activePlayer,
			})
		}
	}

	resp := response.Response{
		Data: GetProviderListResp{Providers: providers},
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Couldn't marshall get provider list response to JSON", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
