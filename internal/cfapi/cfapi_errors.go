package cfapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
)

const (
	cfAPIReqFailed = "CloudFlare API request failed"
)

func (a *API) wrapPretty(err error) error {
	if err == nil {
		return nil
	}

	var cfErr *cloudflare.Error
	isAPIError := errors.As(err, &cfErr)
	if !isAPIError {
		return a.wrapPretty(err)
	}

	//
	messages := []string{}
	for _, err := range cfErr.Errors {
		//
		msg := fmt.Sprintf("%d |> %s", err.Code, err.Message)

		//
		extra := err.JSON.ExtraFields["error_chain"]
		if !extra.IsMissing() {
			extraTxt, errC := gatherErrorChainMessages(extra.Raw())
			if errC == nil {
				msg = fmt.Sprintf("%s |>> %s", msg, extraTxt)
			}
		}

		//
		messages = append(messages, msg)
	}

	//
	failures := strings.Join(messages, ";")

	return fault.Wrap(
		fault.New(failures),
		fmsg.With(cfAPIReqFailed),
		fctx.With(a.ctx,
			"method", cfErr.Request.Method,
			"url", cfErr.Request.URL.String(),
			"statusCode", fmt.Sprintf("%d(%s)", cfErr.StatusCode, http.StatusText(cfErr.StatusCode)),
		),
	)
}

type ErrorChainLink struct {
	Message string `json:"message"`
}

func gatherErrorChainMessages(extraFieldsJSON string) (out string, err error) {
	//
	var links []ErrorChainLink
	err = json.Unmarshal([]byte(extraFieldsJSON), &links)
	if err != nil {
		return
	}

	//
	messages := []string{}
	for _, link := range links {
		messages = append(messages, link.Message)
	}

	//
	out = strings.Join(messages, " -> ")
	return
}
