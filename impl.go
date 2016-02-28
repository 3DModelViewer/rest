package rest

import (
	"github.com/modelhub/core"
	"github.com/modelhub/session"
	"net/http"
	"github.com/robsix/golog"
	"encoding/json"
	"errors"
	"io"
)

func NewRestApi(coreApi core.CoreApi, getSession session.SessionGetter, log golog.Log) http.ServeMux {
	mux := http.NewServeMux()
	//user
	mux.HandleFunc("/api/v1/user/getCurrent", handlerWrapper(coreApi, getSession, userGetCurrent, log))
	mux.HandleFunc("/api/v1/user/setProperty", handlerWrapper(coreApi, getSession, userSetProperty, log))
	mux.HandleFunc("/api/v1/user/get", handlerWrapper(coreApi, getSession, userGet, log))
	mux.HandleFunc("/api/v1/user/getInProjectContext", handlerWrapper(coreApi, getSession, userGetInProjectContext, log))
	mux.HandleFunc("/api/v1/user/getInProjectInviteContext", handlerWrapper(coreApi, getSession, userGetInProjectInviteContext, log))
	mux.HandleFunc("/api/v1/user/search", handlerWrapper(coreApi, getSession, userSearch, log))
	//project
	mux.HandleFunc("/api/v1/project/create", handlerWrapper(coreApi, getSession, projectCreate, log))
	mux.HandleFunc("/api/v1/project/setName", handlerWrapper(coreApi, getSession, projectSetName, log))
	mux.HandleFunc("/api/v1/project/setDescription", handlerWrapper(coreApi, getSession, projectSetDescription, log))
	mux.HandleFunc("/api/v1/project/setImage", handlerWrapper(coreApi, getSession, projectSetImage, log))
	mux.HandleFunc("/api/v1/project/addUsers", handlerWrapper(coreApi, getSession, projectAddUsers, log))
	mux.HandleFunc("/api/v1/project/removeUsers", handlerWrapper(coreApi, getSession, projectRemoveUsers, log))
	mux.HandleFunc("/api/v1/project/acceptInvite", handlerWrapper(coreApi, getSession, projectAcceptInvite, log))
	mux.HandleFunc("/api/v1/project/declineInvite", handlerWrapper(coreApi, getSession, projectDeclineInvite, log))
	mux.HandleFunc("/api/v1/project/getImage/", handlerWrapper(coreApi, getSession, projectGetImage, log))
	mux.HandleFunc("/api/v1/project/get", handlerWrapper(coreApi, getSession, projectGet, log))
	mux.HandleFunc("/api/v1/project/getInUserContext", handlerWrapper(coreApi, getSession, projectGetInUserContext, log))
	mux.HandleFunc("/api/v1/project/getInUserInviteContext", handlerWrapper(coreApi, getSession, projectGetInUserInviteContext, log))
	mux.HandleFunc("/api/v1/project/search", handlerWrapper(coreApi, getSession, projectSearch, log))
	//treeNode
	mux.HandleFunc("/api/v1/treeNode/createFolder", handlerWrapper(coreApi, getSession, treeNodeCreateFolder, log))
	mux.HandleFunc("/api/v1/treeNode/createDocument", handlerWrapper(coreApi, getSession, treeNodeCreateDocument, log))
	mux.HandleFunc("/api/v1/treeNode/setName", handlerWrapper(coreApi, getSession, treeNodeSetName, log))
	mux.HandleFunc("/api/v1/treeNode/move", handlerWrapper(coreApi, getSession, treeNodeMove, log))
	mux.HandleFunc("/api/v1/treeNode/get", handlerWrapper(coreApi, getSession, treeNodeGet, log))
	mux.HandleFunc("/api/v1/treeNode/getChildren", handlerWrapper(coreApi, getSession, treeNodeGetChildren, log))
	mux.HandleFunc("/api/v1/treeNode/getParents", handlerWrapper(coreApi, getSession, treeNodeGetParents, log))
	mux.HandleFunc("/api/v1/treeNode/globalSearch", handlerWrapper(coreApi, getSession, treeNodeGlobalSearch, log))
	mux.HandleFunc("/api/v1/treeNode/projectSearch", handlerWrapper(coreApi, getSession, treeNodeProjectSearch, log))
	//documentVersion
	mux.HandleFunc("/api/v1/documentVersion/create", handlerWrapper(coreApi, getSession, documentVersionCreate, log))
	mux.HandleFunc("/api/v1/documentVersion/get", handlerWrapper(coreApi, getSession, documentVersionGet, log))
	mux.HandleFunc("/api/v1/documentVersion/getForDocument", handlerWrapper(coreApi, getSession, documentVersionGetForDocument, log))
	mux.HandleFunc("/api/v1/documentVersion/getSeedFile/", handlerWrapper(coreApi, getSession, documentVersionGetSeedFile, log))
	//sheet
	mux.HandleFunc("/api/v1/sheet/setName", handlerWrapper(coreApi, getSession, sheetSetName, log))
	mux.HandleFunc("/api/v1/sheet/getItem/", handlerWrapper(coreApi, getSession, sheetGetItem, log))
	mux.HandleFunc("/api/v1/sheet/getForDocumentVersion", handlerWrapper(coreApi, getSession, sheetGetForDocumentVersion, log))
	mux.HandleFunc("/api/v1/sheet/globalSearch", handlerWrapper(coreApi, getSession, sheetGlobalSearch, log))
	mux.HandleFunc("/api/v1/sheet/projectSearch", handlerWrapper(coreApi, getSession, sheetProjectSearch, log))

	return mux
}

//START Util

func handlerWrapper(coreApi core.CoreApi, getSession session.SessionGetter, handler handler, log golog.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if session, err := getSession(w, r); err != nil {
			writeError(w, err, log)
		} else if session == nil {
			writeError(w, errors.New("no session found"), log)
		} else if forUser, err := session.User(); err != nil {
			writeError(w, err, log)
		}else if forUser == "" {
			writeError(w, errors.New("no valid user id in session"), log)
		} else if err := handler(coreApi, session, w, r, log); err != nil {
			writeError(w, err, log)
		}
	}
}

type handler func(core.CoreApi, session.Session, http.ResponseWriter, *http.Request, golog.Log) error

func writeJson(w http.ResponseWriter, src interface{}, log golog.Log) {
	if b, err := json.Marshal(src); err != nil {
		writeError(w, err, log)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

func readJson(r *http.Request, dst interface{}) error {
	if r != nil && r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		return decoder.Decode(dst)
	}
	return nil
}

func writeError(w http.ResponseWriter, err error, log golog.Log) {
	le := log.Error("RestApi error: %v", err)
	w.WriteHeader(500)
	w.Write("unexpected error, id: "+le.LogId)
}

//END Util

//START Handlers

func userGetCurrent(coreApi core.CoreApi, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	if res, err := coreApi.User().Get(session.User()); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func userSetProperty(coreApi core.CoreApi, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Property string `json:"property"`
		Value string `json:"value"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.User().SetProperty(session.User(), args.Property, ); err != nil {
		return err
	} else {
		return nil
	}
}

func userGet(coreApi core.CoreApi, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct{
		Ids []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.User().Get(args.Ids); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

//END Handlers