package models

import "time"

type User struct {
    ID           int    `json:"id"`
    ContactType  string `json:"contact_type"`
    Contact     string `json:"contact"`
}

type UserRequest struct {
    ID           int       `json:"id"`
    UserID       int       `json:"user_id"`
    RequestMethod string   `json:"request_method"`
    TokensIn     int       `json:"tokens_in"`
    TokensOut    int       `json:"tokens_out"`
    DateCreated  time.Time `json:"date_created"`
}
