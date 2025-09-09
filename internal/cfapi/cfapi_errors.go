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

// check if errors notified an HTTP 404 error, meaning resource that we interacted with do not exist
func (a *API) Is404(err error) bool {
	var cfErr *cloudflare.Error
	isNotFound := errors.As(err, &cfErr) && cfErr.StatusCode == 404
	return isNotFound
}

// prettify and add context to complex errors that might arise using CloudFlare API
func (a *API) wrapPrettyForAPI(err error) error {
	if err == nil {
		return nil
	}

	var cfErr *cloudflare.Error
	isAPIError := errors.As(err, &cfErr)

	// We do not want to wrap:
	// - All errors which are not API related
	// - 404 HTTP from CF API (which we might track); they would otherwise not be detected using errors.As(). They are simple enought for prettifying not being useful.
	if !isAPIError || cfErr.StatusCode == 404 {
		return err
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
		// TODO(maintainer) We do not quite want joining here since errors are at the "same level". But, how well, that'd be good enough.
		errors.Join(errs...),
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
