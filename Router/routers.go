package router

import (
	C "example.com/m/Controller"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/user/login", C.UserLogin).Methods("POST")
	router.HandleFunc("/api/user/signup", C.UserSignup).Methods("POST")
	router.HandleFunc("/api/posts", C.GetMyAllPosts).Methods("GET")
	router.HandleFunc("/api/group", C.CreateGroup).Methods("POST")
	router.HandleFunc("/api/groups", C.GetMyAllGroups).Methods("GET")
	router.HandleFunc("/api/update", C.UpdatePost).Methods("PATCH")
	router.HandleFunc("/api/updateGrp", C.CreatePost).Methods("PATCH")
	router.HandleFunc("/api/viewdetails", C.ViewDetails).Methods("GET")

	return router
}
