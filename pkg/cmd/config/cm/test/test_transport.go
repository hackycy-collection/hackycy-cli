package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

// cmTestProviderTransport is the command-owned HTTP boundary for config cm test.
type cmTestProviderTransport interface {
	Do(*http.Request) (*http.Response, error)
}

func executeCMTestProvider(ctx context.Context, profile appconfig.ResolvedCMProfile, transport cmTestProviderTransport) (cmTestProviderResult, error) {
	timeout := time.Duration(profile.TimeoutMS * float64(time.Millisecond))
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := newCMTestProviderRequest(requestContext, profile)
	if err != nil {
		return cmTestProviderResult{}, err
	}
	response, err := transport.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestContext.Err(), context.DeadlineExceeded) {
			return cmTestProviderResult{}, fmt.Errorf("Provider request timed out after %sms", strconv.FormatFloat(profile.TimeoutMS, 'f', -1, 64))
		}
		return cmTestProviderResult{}, err
	}
	if response == nil {
		return cmTestProviderResult{}, errors.New("CM test provider returned no response")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return cmTestProviderResult{}, fmt.Errorf("read CM test provider response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return cmTestProviderResult{}, cmTestHTTPError(response, body)
	}
	return decodeCMTestProviderResponse(body)
}

func cmTestHTTPError(response *http.Response, body []byte) error {
	statusText := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(response.Status), strconv.Itoa(response.StatusCode)))
	if statusText == "" {
		statusText = http.StatusText(response.StatusCode)
	}
	message := fmt.Sprintf("%d %s", response.StatusCode, statusText)
	if len(body) > 0 {
		message += ": " + string(body)
	}
	return errors.New(message)
}
