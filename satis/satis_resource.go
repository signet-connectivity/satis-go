package satis

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ilyes512/satis-go/satis/satisphp"
	"github.com/Ilyes512/satis-go/satis/satisphp/api"
	"github.com/gorilla/mux"
)

// SatisResource struct
type SatisResource struct {
	Host           string
	SatisPhpClient satisphp.SatisClient
	Username       string
	APIToken       string
}

// Add repository in Satis Repo and regenerate static web docs
func (r *SatisResource) addRepo(res http.ResponseWriter, req *http.Request) {
	// unmarshal post body
	apiR := &api.Repo{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(apiR); err != nil {
		log.Print(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	repo := api.NewRepo(apiR.Type, apiR.URL)

	if _, err := r.SatisPhpClient.FindRepo(repo.ID); err == nil || err != satisphp.ErrRepoNotFound {
		res.WriteHeader(http.StatusConflict)
		return
	}

	body, err := r.upsertRepo(repo)
	if err != nil {
		log.Print(err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.Header().Set("Location", fmt.Sprintf("%s/api/repo/%s", r.Host, repo.ID))
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, body)
}

// Add repository in Satis Repo and regenerate static web docs
func (r *SatisResource) saveRepo(res http.ResponseWriter, req *http.Request) {
	repoID := mux.Vars(req)["id"]

	repo := &api.Repo{}

	// unmarshal post body
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(repo); err != nil {
		log.Print(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if repo.ID != "" && repo.ID != repoID {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	existing, err := r.SatisPhpClient.FindRepo(repoID)
	if err != nil {
		switch err {
		case satisphp.ErrRepoNotFound:
			res.WriteHeader(http.StatusNotFound)
		default:
			log.Print(err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	existing.Type = repo.Type
	existing.URL = repo.URL

	body, err := r.upsertRepo(&existing)
	if err != nil {
		log.Print(err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res, body)
}

// save config and regenerate satis-web
func (r *SatisResource) upsertRepo(repo *api.Repo) (string, error) {
	if err := r.SatisPhpClient.SaveRepo(repo, true); err != nil {
		return "", err
	}

	// marshal response
	newRepo := api.NewRepo(repo.Type, repo.URL)
	b, err := json.Marshal(newRepo)
	if err != nil {
		return "", err
	}
	return string(b[:]), nil
}

// Get One Repo
func (r *SatisResource) findRepo(res http.ResponseWriter, req *http.Request) {
	repoID := mux.Vars(req)["id"]

	repo, err := r.SatisPhpClient.FindRepo(repoID)

	if err != nil {
		switch err {
		case satisphp.ErrRepoNotFound:
			res.WriteHeader(http.StatusNotFound)
		default:
			log.Print(err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// marshal response
	b, err := json.Marshal(repo)
	if err != nil {
		log.Print(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res, string(b[:]))
}

// Get All Repos
func (r *SatisResource) findAllRepos(res http.ResponseWriter, req *http.Request) {

	repos, err := r.SatisPhpClient.FindAllRepos()
	if err != nil {
		log.Print(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// marshal response
	b, err := json.Marshal(repos)
	if err != nil {
		log.Print(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res, string(b[:]))
}

func (r *SatisResource) deleteRepo(res http.ResponseWriter, req *http.Request) {
	repoID := mux.Vars(req)["id"]

	if err := r.SatisPhpClient.DeleteRepo(repoID, true); err != nil {
		switch err {
		case satisphp.ErrRepoNotFound:
			res.WriteHeader(http.StatusNotFound)
		default:
			log.Print(err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusNoContent)
}

// Regenerate static web docs
func (r *SatisResource) generateStaticWeb(res http.ResponseWriter, req *http.Request) {
	if err := r.SatisPhpClient.GenerateSatisWeb(); err != nil {
		log.Print(err)

		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "application/json")
}

func (r *SatisResource) generateStaticWebNow() error {
	return r.SatisPhpClient.GenerateSatisWeb()
}

func (r *SatisResource) updatePackage(res http.ResponseWriter, req *http.Request) {
	if r.Username != "" && r.APIToken != "" {
		if req.URL.Query()["username"][0] != r.Username || req.URL.Query()["apiToken"][0] != r.APIToken {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if err := r.SatisPhpClient.GenerateSatisWeb(); err != nil {
		log.Print(err)

		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
	res.Header().Set("Content-Type", "application/json")
}
