package models

type Tags struct {
	ID     string `json:"_id" bson:"_id"`
	Name   string `json:"name" bson:"name"`
	Color  string `json:"color" bson:"color"`
	Tenant string `json:"tenant_id" bson:"tenant_id"`
}
