package helper

import (
	"github.com/bitrainforest/filmeta-hic/core/errno"
	"github.com/bitrainforest/filmeta-hic/core/httpx/response"
	"github.com/bitrainforest/pulsar/api/codex"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func WarpMongoErr(err error) error {
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	return err
}

func WarpResp(err error) response.Response {
	if errn, ok := err.(errno.Error); ok {
		return errn
	}
	return codex.ErrService.FormatErrMsg(err)
}
