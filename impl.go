package rest

import (
	"github.com/modelhub/core"
	"github.com/modelhub/session"
	"net/http"
	"github.com/robsix/golog"
)

func NewRestApi(coreApi core.CoreApi, getSession session.SessionGetter, log golog.Log) http.ServeMux {
	mux := http.NewServeMux()
	//user
	mux.HandleFunc("/api/v1/user/getCurrent",)
	mux.HandleFunc("/api/v1/user/setProperty",)
	mux.HandleFunc("/api/v1/user/get",)
	mux.HandleFunc("/api/v1/user/getInProjectContext",)
	mux.HandleFunc("/api/v1/user/getInProjectInviteContext",)
	mux.HandleFunc("/api/v1/user/search",)
	//project
	mux.HandleFunc("/api/v1/project/create",)
	mux.HandleFunc("/api/v1/project/setName",)
	mux.HandleFunc("/api/v1/project/setDescription",)
	mux.HandleFunc("/api/v1/project/setImage",)
	mux.HandleFunc("/api/v1/project/addUsers",)
	mux.HandleFunc("/api/v1/project/acceptInvite",)
	mux.HandleFunc("/api/v1/project/declineInvite",)
	mux.HandleFunc("/api/v1/project/getImage/",)
	mux.HandleFunc("/api/v1/project/get",)
	mux.HandleFunc("/api/v1/project/getInUserContext",)
	mux.HandleFunc("/api/v1/project/getInUserInviteContext",)
	mux.HandleFunc("/api/v1/project/search",)
	//treeNode
	mux.HandleFunc("/api/v1/treeNode/createFolder",)
	mux.HandleFunc("/api/v1/treeNode/createDocument",)
	mux.HandleFunc("/api/v1/treeNode/setName",)
	mux.HandleFunc("/api/v1/treeNode/move",)
	mux.HandleFunc("/api/v1/treeNode/get",)
	mux.HandleFunc("/api/v1/treeNode/getChildren",)
	mux.HandleFunc("/api/v1/treeNode/getParents",)
	mux.HandleFunc("/api/v1/treeNode/globalSearch",)
	mux.HandleFunc("/api/v1/treeNode/projectSearch",)
	//documentVersion
	mux.HandleFunc("/api/v1/documentVersion/create",)
	mux.HandleFunc("/api/v1/documentVersion/get",)
	mux.HandleFunc("/api/v1/documentVersion/getForDocument",)
	mux.HandleFunc("/api/v1/documentVersion/getSeedFile/",)
	//sheet
	mux.HandleFunc("/api/v1/sheet/setName",)
	mux.HandleFunc("/api/v1/sheet/getItem/",)
	mux.HandleFunc("/api/v1/sheet/getForDocumentVersion",)
	mux.HandleFunc("/api/v1/sheet/globalSearch",)
	mux.HandleFunc("/api/v1/sheet/projectSearch",)
	return mux
}