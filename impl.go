package rest

import (
	"encoding/json"
	"errors"
	"github.com/modelhub/core"
	"github.com/modelhub/session"
	"github.com/modelhub/vada"
	"github.com/modelhub/core/user"
	"github.com/modelhub/core/project"
	"github.com/modelhub/core/treenode"
	"github.com/modelhub/core/documentversion"
	"github.com/modelhub/core/sheet"
	"github.com/robsix/golog"
	sj "github.com/robsix/json"
	"io"
	"net/http"
	"strings"
	"github.com/modelhub/core/helper"
)

const (
	sheetGetItemPath = "/api/v1/sheet/getItem/"
)

func NewRestApi(coreApi core.CoreApi, getSession session.SessionGetter, vada vada.VadaClient, log golog.Log) *http.ServeMux {
	mux := http.NewServeMux()
	//user
	mux.HandleFunc("/api/v1/user/getCurrent", handlerWrapper(coreApi, getSession, userGetCurrent, log))
	mux.HandleFunc("/api/v1/user/setProperty", handlerWrapper(coreApi, getSession, userSetProperty, log))
	mux.HandleFunc("/api/v1/user/get", handlerWrapper(coreApi, getSession, userGet, log))
	mux.HandleFunc("/api/v1/user/search", handlerWrapper(coreApi, getSession, userSearch, log))
	//project
	mux.HandleFunc("/api/v1/project/create", handlerWrapper(coreApi, getSession, projectCreate, log))
	mux.HandleFunc("/api/v1/project/setName", handlerWrapper(coreApi, getSession, projectSetName, log))
	mux.HandleFunc("/api/v1/project/setThumbnail", handlerWrapper(coreApi, getSession, projectSetThumbnail, log))
	mux.HandleFunc("/api/v1/project/addUsers", handlerWrapper(coreApi, getSession, projectAddUsers, log))
	mux.HandleFunc("/api/v1/project/removeUsers", handlerWrapper(coreApi, getSession, projectRemoveUsers, log))
	mux.HandleFunc("/api/v1/project/acceptInvite", handlerWrapper(coreApi, getSession, projectAcceptInvite, log))
	mux.HandleFunc("/api/v1/project/declineInvite", handlerWrapper(coreApi, getSession, projectDeclineInvite, log))
	mux.HandleFunc("/api/v1/project/getRole", handlerWrapper(coreApi, getSession, projectGetRole, log))
	mux.HandleFunc("/api/v1/project/getMemberships", handlerWrapper(coreApi, getSession, projectGetMemberships, log))
	mux.HandleFunc("/api/v1/project/getMembershipInvites", handlerWrapper(coreApi, getSession, projectGetMembershipInvites, log))
	mux.HandleFunc("/api/v1/project/getThumbnail/", handlerWrapper(coreApi, getSession, projectGetThumbnail, log))
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
	mux.HandleFunc("/api/v1/documentVersion/getThumbnail/", handlerWrapper(coreApi, getSession, documentVersionGetThumbnail, log))
	//sheet
	mux.HandleFunc("/api/v1/sheet/setName", handlerWrapper(coreApi, getSession, sheetSetName, log))
	mux.HandleFunc(sheetGetItemPath, handlerWrapper(coreApi, getSession, sheetGetItem(vada), log))
	mux.HandleFunc("/api/v1/sheet/get", handlerWrapper(coreApi, getSession, sheetGet, log))
	mux.HandleFunc("/api/v1/sheet/getForDocumentVersion", handlerWrapper(coreApi, getSession, sheetGetForDocumentVersion, log))
	mux.HandleFunc("/api/v1/sheet/globalSearch", handlerWrapper(coreApi, getSession, sheetGlobalSearch, log))
	mux.HandleFunc("/api/v1/sheet/projectSearch", handlerWrapper(coreApi, getSession, sheetProjectSearch, log))
	//helpers
	mux.HandleFunc("/api/v1/helper/getChildrenDocumentsWithLatestVersionAndFirstSheetInfo", handlerWrapper(coreApi, getSession, helperGetChildrenDocumentsWithLatestVersionAndFirstSheetInfo, log))
	mux.HandleFunc("/api/v1/helper/getDocumentVersionsWithFirstSheetInfo", handlerWrapper(coreApi, getSession, helperGetDocumentVersionsWithFirstSheetInfo, log))

	return mux
}

//START Util

func handlerWrapper(coreApi core.CoreApi, getSession session.SessionGetter, handler handler, log golog.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r != nil && r.Body != nil {
			defer r.Body.Close()
		}
		if session, err := getSession(w, r); err != nil {
			writeError(w, err, log)
		} else if session == nil {
			writeError(w, errors.New("no session found"), log)
		} else if forUser, err := session.User(); err != nil {
			writeError(w, err, log)
		} else if forUser == "" {
			writeError(w, errors.New("no valid user id in session"), log)
		} else if err := handler(coreApi, forUser, session, w, r, log); err != nil {
			writeError(w, err, log)
		}
	}
}

type handler func(core.CoreApi, string, session.Session, http.ResponseWriter, *http.Request, golog.Log) error

func writeJson(w http.ResponseWriter, src interface{}, log golog.Log) {
	if b, err := json.Marshal(src); err != nil {
		writeError(w, err, log)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

func writeOffsetJson(w http.ResponseWriter, res interface{}, totalResults int, log golog.Log) {
	if res == nil {
		res = []interface{}{}
	}
	writeJson(w, &struct {
		TotalResults int `json:"totalResults"`
		Results      interface{} `json:"results"`
	}{
		TotalResults: totalResults,
		Results:      res,
	}, log)
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
	w.Write([]byte(le.LogId))
}

//END Util

//START Handlers

func userGetCurrent(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	if res, err := coreApi.User().GetCurrent(forUser); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func userSetProperty(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Property string `json:"property"`
		Value    string `json:"value"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if prop, err := user.Property(args.Property); err != nil {
		return err
	} else if err := coreApi.User().SetProperty(forUser, prop, args.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func userGet(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
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

func userSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Search string `json:"search"`
		Role   string `json:"role"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.User().Search(args.Search, args.Offset, args.Limit, user.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func projectCreate(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	thumbnail, _, err := r.FormFile("thumbnail")
	if thumbnail != nil {
		defer thumbnail.Close()
	}
	if err != nil && err != http.ErrMissingFile {
		return err
	}
	if res, err := coreApi.Project().Create(forUser, r.FormValue("name"), r.FormValue("thumbnailType"), thumbnail); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func projectSetName(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Project().SetName(forUser, args.Id, args.Name); err != nil {
		return err
	} else {
		return nil
	}
}

func projectSetThumbnail(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	thumbnail, _, err := r.FormFile("thumbnail")
	if thumbnail != nil {
		defer thumbnail.Close()
	}
	if err != nil && err != http.ErrMissingFile {
		return err
	}
	if err := coreApi.Project().SetThumbnail(forUser, r.FormValue("id"), r.FormValue("thumbnailType"), thumbnail); err != nil {
		return err
	} else {
		return nil
	}
}

func projectAddUsers(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id    string   `json:"id"`
		Role  string   `json:"role"`
		Users []string `json:"users"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Project().AddUsers(forUser, args.Id, project.Role(args.Role), args.Users); err != nil {
		return err
	} else {
		return nil
	}
}

func projectRemoveUsers(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id    string   `json:"id"`
		Users []string `json:"users"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Project().RemoveUsers(forUser, args.Id, args.Users); err != nil {
		return err
	} else {
		return nil
	}
}

func projectAcceptInvite(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id string `json:"id"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Project().AcceptInvite(forUser, args.Id); err != nil {
		return err
	} else {
		return nil
	}
}

func projectDeclineInvite(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id string `json:"id"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Project().DeclineInvite(forUser, args.Id); err != nil {
		return err
	} else {
		return nil
	}
}

func projectGetRole(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id string `json:"id"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.Project().GetRole(forUser, args.Id); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func projectGetMemberships(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id   string `json:"id"`
		Role   string `json:"role"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Project().GetMemberships(forUser, args.Id, project.Role(args.Role), args.Offset, args.Limit, project.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func projectGetMembershipInvites(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id   string `json:"id"`
		Role   string `json:"role"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Project().GetMembershipInvites(forUser, args.Id, project.Role(args.Role), args.Offset, args.Limit, project.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func projectGetThumbnail(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	pathSegments := strings.Split(r.URL.Path, "/")
	id := pathSegments[len(pathSegments)-3]
	mimeType := pathSegments[len(pathSegments)-2]
	mimeSubtype := pathSegments[len(pathSegments)-1]
	var res *http.Response
	var err error
	if res, err = coreApi.Project().GetThumbnail(forUser, id); res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	} else if _, err := io.Copy(w, res.Body); err != nil {
		return err
	} else {
		w.Header().Add("Content-Type", mimeType+"/"+mimeSubtype)
		return nil
	}
}

func projectGet(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Ids []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.Project().Get(forUser, args.Ids); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func projectGetInUserContext(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		User   string `json:"user"`
		Role   string `json:"role"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Project().GetInUserContext(forUser, args.User, project.Role(args.Role), args.Offset, args.Limit, project.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func projectGetInUserInviteContext(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		User   string `json:"user"`
		Role   string `json:"role"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Project().GetInUserInviteContext(forUser, args.User, project.Role(args.Role), args.Offset, args.Limit, project.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func projectSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Search string `json:"search"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Project().Search(forUser, args.Search, args.Offset, args.Limit, project.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func treeNodeCreateFolder(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Parent string `json:"parent"`
		Name   string `json:"name"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.TreeNode().CreateFolder(forUser, args.Parent, args.Name); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func treeNodeCreateDocument(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	file, header, err := r.FormFile("file")
	if file != nil {
		defer file.Close()
	}
	fileName := ""
	if header != nil {
		fileName = header.Filename
	}
	if err != nil {
		return err
	}

	thumbnail, _, err := r.FormFile("thumbnail")
	if thumbnail != nil {
		defer thumbnail.Close()
	}
	if err != nil && err != http.ErrMissingFile {
		return err
	}

	if res, err := coreApi.TreeNode().CreateDocument(forUser, r.FormValue("parent"), r.FormValue("name"), r.FormValue("uploadComment"), r.FormValue("fileType"), fileName, file, r.FormValue("thumbnailType"), thumbnail); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func treeNodeSetName(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.TreeNode().SetName(forUser, args.Id, args.Name); err != nil {
		return err
	} else {
		return nil
	}
}

func treeNodeMove(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Parent string   `json:"parent"`
		Ids    []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.TreeNode().Move(forUser, args.Parent, args.Ids); err != nil {
		return err
	} else {
		return nil
	}
}

func treeNodeGet(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Ids []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.TreeNode().Get(forUser, args.Ids); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func treeNodeGetChildren(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id       string `json:"id"`
		NodeType string `json:"nodeType"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
		SortBy   string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.TreeNode().GetChildren(forUser, args.Id, treenode.NodeType(args.NodeType), args.Offset, args.Limit, treenode.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func treeNodeGetParents(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id string `json:"id"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.TreeNode().GetParents(forUser, args.Id); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func treeNodeGlobalSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Search   string `json:"search"`
		NodeType string `json:"nodeType"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
		SortBy   string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.TreeNode().GlobalSearch(forUser, args.Search, treenode.NodeType(args.NodeType), args.Offset, args.Limit, treenode.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func treeNodeProjectSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Project  string `json:"project"`
		Search   string `json:"search"`
		NodeType string `json:"nodeType"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
		SortBy   string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.TreeNode().ProjectSearch(forUser, args.Project, args.Search, treenode.NodeType(args.NodeType), args.Offset, args.Limit, treenode.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func documentVersionCreate(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	file, header, err := r.FormFile("file")
	if file != nil {
		defer file.Close()
	}
	fileName := ""
	if header != nil {
		fileName = header.Filename
	}
	if err != nil {
		return err
	}

	thumbnail, _, err := r.FormFile("thumbnail")
	if thumbnail != nil {
		defer thumbnail.Close()
	}
	if err != nil && err != http.ErrMissingFile {
		return err
	}

	if res, err := coreApi.DocumentVersion().Create(forUser, r.FormValue("document"), r.FormValue("uploadComment"), r.FormValue("fileType"), fileName, file, r.FormValue("thumbnailType"), thumbnail); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func documentVersionGet(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Ids []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.DocumentVersion().Get(forUser, args.Ids); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func documentVersionGetForDocument(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Document string `json:"document"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
		SortBy   string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.DocumentVersion().GetForDocument(forUser, args.Document, args.Offset, args.Limit, documentversion.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func documentVersionGetSeedFile(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	pathSegments := strings.Split(r.URL.Path, "/")
	id := pathSegments[len(pathSegments)-1]
	var res *http.Response
	var err error
	if res, err = coreApi.DocumentVersion().GetSeedFile(forUser, id); res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	} else if _, err := io.Copy(w, res.Body); err != nil {
		return err
	} else {
		return nil
	}
}

func documentVersionGetThumbnail(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	pathSegments := strings.Split(r.URL.Path, "/")
	id := pathSegments[len(pathSegments)-3]
	mimeType := pathSegments[len(pathSegments)-2]
	mimeSubtype := pathSegments[len(pathSegments)-1]
	var res *http.Response
	var err error
	if res, err = coreApi.DocumentVersion().GetThumbnail(forUser, id); res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	} else if _, err := io.Copy(w, res.Body); err != nil {
		return err
	} else {
		w.Header().Add("Content-Type", mimeType+"/"+mimeSubtype)
		return nil
	}
}

func sheetSetName(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if err := coreApi.Sheet().SetName(forUser, args.Id, args.Name); err != nil {
		return err
	} else {
		return nil
	}
}

func sheetGetItem(vada vada.VadaClient) handler {
	return func(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
		id := r.URL.Path[len(sheetGetItemPath):]
		path := ""
		if slashIdx := strings.Index(id, "/"); slashIdx == -1 {
			return errors.New("can't find item path in sheetGetItem call")
		} else {
			path = id[slashIdx:]
			id = id[:slashIdx]
		}
		baseUrn := ""
		var res *http.Response
		var err error
		if baseUrn, err = session.GetSheetBaseUrn(id); err != nil {
			if res, baseUrn, err = coreApi.Sheet().GetItem(forUser, id, path); err == nil {
				session.SetAccessedSheet(id, baseUrn)
			}
		} else {
			log.Info("RestApi session recentlyAccessSheet was found :)") //TODO delete this once verified it works
			res, err = vada.GetSheetItem(baseUrn + path)
		}
		if res != nil && res.Body != nil {
			defer res.Body.Close()
			contentType := strings.Join(res.Header["Content-Type"], ",")
			w.Header().Add("Content-Type", contentType)
			if contentType == "application/json" {
				log.Info("RestApi making json lmv safe") //TODO delete this once verifieid it works
				if js, err := sj.FromReadCloser(res.Body); err != nil {
					return err
				} else {
					unsafeJsonStr, _ := js.ToString()
					log.Info("unsafe json %v", unsafeJsonStr) //TODO delete this once verifieid it works
					safeJsonStr := strings.Replace(unsafeJsonStr, baseUrn+"/", sheetGetItemPath, -1)
					log.Info("safe json %v", safeJsonStr) //TODO delete this once verifieid it works
					writeJson(w, safeJsonStr, log)
				}
			} else {
				_, err = io.Copy(w, res.Body)
			}
		}
		return err
	}
}

func sheetGet(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Ids []string `json:"ids"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, err := coreApi.Sheet().Get(forUser, args.Ids); err != nil {
		return err
	} else {
		writeJson(w, res, log)
		return nil
	}
}

func sheetGetForDocumentVersion(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		DocumentVersion string `json:"documentVersion"`
		Offset          int    `json:"offset"`
		Limit           int    `json:"limit"`
		SortBy          string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Sheet().GetForDocumentVersion(forUser, args.DocumentVersion, args.Offset, args.Limit, sheet.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func sheetGlobalSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Search string `json:"search"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Sheet().GlobalSearch(forUser, args.Search, args.Offset, args.Limit, sheet.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func sheetProjectSearch(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Project string `json:"project"`
		Search  string `json:"search"`
		Offset  int    `json:"offset"`
		Limit   int    `json:"limit"`
		SortBy  string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Sheet().ProjectSearch(forUser, args.Project, args.Search, args.Offset, args.Limit, sheet.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func helperGetChildrenDocumentsWithLatestVersionAndFirstSheetInfo(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Folder     string `json:"folder"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Helper().GetChildrenDocumentsWithLatestVersionAndFirstSheetInfo(forUser, args.Folder, args.Offset, args.Limit, helper.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

func helperGetDocumentVersionsWithFirstSheetInfo(coreApi core.CoreApi, forUser string, session session.Session, w http.ResponseWriter, r *http.Request, log golog.Log) error {
	args := &struct {
		Document     string `json:"document"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		SortBy string `json:"sortBy"`
	}{}
	if err := readJson(r, args); err != nil {
		return err
	} else if res, totalResults, err := coreApi.Helper().GetDocumentVersionsWithFirstSheetInfo(forUser, args.Document, args.Offset, args.Limit, helper.SortBy(args.SortBy)); err != nil {
		return err
	} else {
		writeOffsetJson(w, res, totalResults, log)
		return nil
	}
}

//END Handlers
