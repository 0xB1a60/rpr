package transport

import (
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/http"
)

type RemoveReq struct {
	Id string `json:"id"`
}

func (s *Server) processRemove(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w)

	var req RemoveReq
	if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error("err while decoding",
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad body"))
		return
	}

	if len(req.Id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Id must not be empty"))
		return
	}

	if err := s.DB.RemoveKV(r.Context(), req.Id); err != nil {
		s.Logger.Error("err while writing to database",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("err while writing to database"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
