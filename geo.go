// This Go program implements a web server that serves HTML and JS files.
// Created with assistance from GPT by OpenAI.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Nominatim APIのレスポンス形式
type NominatimResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// レスポンス形式
type DistanceResponse struct {
	Address1 string  `json:"address1"`
	Address2 string  `json:"address2"`
	Distance float64 `json:"distance_km"`
}

// ジオコーディング：Nominatimを使って住所から緯度・経度を取得する関数
func getLatLon(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")
	endpoint := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// リクエストにUser-Agentヘッダーが必要
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "OpenStreetMapGeocoder")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var nominatimResponse NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&nominatimResponse); err != nil {
		return 0, 0, err
	}

	if len(nominatimResponse) == 0 {
		return 0, 0, fmt.Errorf("住所が見つかりません")
	}

	// 緯度と経度を文字列からfloat64に変換
	lat := parseStringToFloat(nominatimResponse[0].Lat)
	lon := parseStringToFloat(nominatimResponse[0].Lon)

	return lat, lon, nil
}

// 文字列をfloat64に変換する関数
func parseStringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		log.Fatalf("数値変換エラー: %v", err)
	}
	return f
}

// ハバースインの公式を使った大円距離の計算
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // 地球の半径 (km)

	// 度をラジアンに変換
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// 緯度と経度の差
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// ハバースインの公式
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// 距離 (km)
	distance := R * c
	return distance
}

// 距離を計算するハンドラー
func distanceHandler(w http.ResponseWriter, r *http.Request) {
	address1 := r.URL.Query().Get("address1")
	address2 := r.URL.Query().Get("address2")

	if address1 == "" || address2 == "" {
		http.Error(w, "住所1と住所2を指定してください", http.StatusBadRequest)
		return
	}

	// 住所1の緯度・経度を取得
	lat1, lon1, err := getLatLon(address1)
	if err != nil {
		http.Error(w, fmt.Sprintf("住所1の取得エラー: %v", err), http.StatusInternalServerError)
		return
	}

	// 住所2の緯度・経度を取得
	lat2, lon2, err := getLatLon(address2)
	if err != nil {
		http.Error(w, fmt.Sprintf("住所2の取得エラー: %v", err), http.StatusInternalServerError)
		return
	}

	// 大円距離を計算
	distance := haversine(lat1, lon1, lat2, lon2)

	// レスポンスを返す
	response := DistanceResponse{
		Address1: address1,
		Address2: address2,
		Distance: distance,
	}

	// レスポンスを返す前にログ出力
	log.Printf("Response: %+v\n", response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 静的ファイルを提供するハンドラー
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/", http.FileServer(http.Dir("./"))).ServeHTTP(w, r)
}

func main() {
	http.HandleFunc("/distance", distanceHandler)
	http.HandleFunc("/", staticFileHandler)

	fmt.Println("サーバーを起動中: http://localhost:8080/")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("サーバーの起動エラー: %v", err)
	}
}
