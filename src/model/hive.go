package model

type Hive struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type Sector struct {
	ID int `json:"id"`
	HiveID int `json:"hive_id"`
	Name string `json:"name"`
	IP string `json:"ip"`
	Port int `json:"port"`
}
