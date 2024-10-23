package models

type Users struct {
	Id           int64
	Login        string  `json:"login"`
	PasswordHash string  `json:"-"`
	Balance      float64 `json:"balance"`
}

type Assets struct {
	Id          int64
	Name        string
	Description string
	Price       float64
	CreatorId   int64
}

type AssetsOwners struct {
	OwnerId int64
	AssetId int64
}
