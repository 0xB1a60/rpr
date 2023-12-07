package transport

import (
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/http"
)

type EditReq struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

func (s *Server) processEdit(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w)

	var req EditReq
	if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error("err while decoding",
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad body"))
		return
	}

	if len(req.Id) == 0 || len(req.Value) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Id and value must not be empty"))
		return
	}

	if len(req.Value) > 10_000 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Max length allowed for value is 10_000"))
		return
	}

	changes, err := s.DB.EditKV(r.Context(), req.Id, req.Value)
	if err != nil {
		s.Logger.Error("err while writing to database",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("err while writing to database"))
		return
	}

	if *changes == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id does not exist"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
