package cfapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

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
	errs := []error{}
	for _, errData := range cfErr.Errors {
		getErr := func() error {
			//
			rawChain := errData.JSON.ExtraFields["error_chain"]
			if !rawChain.IsMissing() {
				chain, errC := wrapChainMessages(rawChain.Raw())
				if errC == nil {
					return fault.Wrap(chain,
						fmsg.WithDesc(strconv.FormatInt(errData.Code, 10), errData.Message),
					)
				}
			}

			return fault.New(errData.Message)
		}

		//
		errs = append(errs, getErr())
	}

	return fault.Wrap(
		errors.Join(errs...), // TODO(maintainer) I have doubts joining errors is handled by Fault
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

func wrapChainMessages(chainMessagesJSON string) (out error, err error) {
	//
	var links []ErrorChainLink
	err = json.Unmarshal([]byte(chainMessagesJSON), &links)
	if err != nil {
		return
	}

	//
	messages := []string{}
	for _, link := range links {
		messages = append(messages, link.Message)
	}

	for _, msg := range slices.Backward(messages) {
		if out == nil {
			out = fault.New(msg)
			continue
		}
		out = fault.Wrap(out, fmsg.With(msg))
	}

	return
}
