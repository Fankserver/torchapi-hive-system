package model

type Sector struct {
	ID     int    `json:"id"`
	HiveID int    `json:"hive_id"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Port   int    `json:"port"`
}
