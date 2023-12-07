package transport

import (
	jsoniter "github.com/json-iterator/go"
	gonanoid "github.com/matoous/go-nanoid"
	"go.uber.org/zap"
	"net/http"
)

type AddReq struct {
	Value string `json:"value"`
}

type AddRes struct {
	Id string `json:"id"`
}

func (s *Server) processAdd(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w)

	var req AddReq
	if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error("err while decoding",
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad body"))
		return
	}

	if len(req.Value) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Value must not be empty"))
		return
	}

	if len(req.Value) > 10_000 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Max length allowed for value is 10_000"))
		return
	}

	id, err := gonanoid.Nanoid()
	if err != nil {
		s.Logger.Error("err while generating id",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("err while generating id"))
		return
	}

	if err = s.DB.AddKV(r.Context(), id, req.Value); err != nil {
		s.Logger.Error("err while writing to database",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("err while writing to database"))
		return
	}

	res, err := jsoniter.Marshal(&AddRes{Id: id})
	if err != nil {
		s.Logger.Error("err while marshaling response",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("err while marshaling response"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
