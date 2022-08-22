package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	octopusApiClient "github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/constants"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"testing"
)

func GoBegin[TResult any](action func() TResult) chan TResult {
	c := make(chan TResult)
	go func() {
		c <- action()
	}()
	return c
}

type Pair[T1 any, T2 any] struct {
	Item1 T1
	Item2 T2
}

func GoBegin2[TResult1 any, TResult2 any](action func() (TResult1, TResult2)) chan Pair[TResult1, TResult2] {
	c := make(chan Pair[TResult1, TResult2])
	go func() {
		r1, r2 := action()
		c <- Pair[TResult1, TResult2]{Item1: r1, Item2: r2}
	}()
	return c
}

func ReceivePair[T1 any, T2 any](receiver chan Pair[T1, T2]) (T1, T2) {
	pair := <-receiver
	return pair.Item1, pair.Item2
}

type responseOrError struct {
	response *http.Response
	error    error
}

type MockHttpServer struct {
	// when the client issues a request, we receive it here
	Request chan *http.Request
	// when we want to respond back to the client, we send it here
	Response chan responseOrError

	// so test code can detect unanswered requests or responses at the end.
	// Not strictly neccessary as unanswered req/resp results in a channel deadlock
	// and go panics and kills the process, so we find out about it, but this is a bit
	// less confusing to troubleshoot
	pendingMsgCount int32

	Closed bool
}

// conforms to RoundTripper so we can use it for httpClient.Transport

func (m *MockHttpServer) RoundTrip(r *http.Request) (*http.Response, error) {
	// we're the client here, so we send a request down the request channel
	atomic.AddInt32(&m.pendingMsgCount, 1)
	m.Request <- r
	atomic.AddInt32(&m.pendingMsgCount, -1)
	// then we wait for a response via the response channel

	atomic.AddInt32(&m.pendingMsgCount, 1)
	x := <-m.Response
	atomic.AddInt32(&m.pendingMsgCount, -1)
	return x.response, x.error
}

func NewMockHttpServer() *MockHttpServer {
	return &MockHttpServer{
		Request:  make(chan *http.Request),
		Response: make(chan responseOrError),
	}
}

func (m *MockHttpServer) Close() {
	m.Closed = true
	close(m.Request)
	close(m.Response)
}

func (m *MockHttpServer) GetPendingMessageCount() int {
	return int(m.pendingMsgCount)
}

func (m *MockHttpServer) ReceiveRequest() (*http.Request, bool) {
	atomic.AddInt32(&m.pendingMsgCount, 1)
	request := <-m.Request
	atomic.AddInt32(&m.pendingMsgCount, -1)
	return request, !m.Closed // reading from closed channels works fine and just returns the default
}

func (m *MockHttpServer) Respond(response *http.Response, err error) {
	if m.Closed {
		return // can't respond after closure
	}

	atomic.AddInt32(&m.pendingMsgCount, 1)
	m.Response <- responseOrError{response: response, error: err}
	atomic.AddInt32(&m.pendingMsgCount, -1)
}

// now we build some higher level methods on top of ReceiveRequest

func (m *MockHttpServer) ExpectRequest(t *testing.T, method string, pathAndQuery string) *RequestWrapper {
	r, ok := m.ReceiveRequest()
	if !ok { // this means the channel was closed
		// don't fatal, there'll be some other assertion failure too and we want to let that have a chance to print
		t.Errorf("ExpectRequest %s %s failed; channel closed", method, pathAndQuery)
		return &RequestWrapper{nil, m}
	}

	rPathAndQuery := r.URL.Path
	if r.URL.RawQuery != "" {
		rPathAndQuery = fmt.Sprintf("%s?%s", rPathAndQuery, r.URL.RawQuery)
	}
	assert.Equal(t, method, r.Method)
	assert.Equal(t, pathAndQuery, rPathAndQuery)

	return &RequestWrapper{r, m}
}

type RequestWrapper struct {
	// in case you need it
	Request *http.Request
	Server  *MockHttpServer
}

func (r *RequestWrapper) RespondWith(responseObject any) {
	if responseObject == nil {
		panic("TODO: implement responses with no body")
	}

	body, _ := json.Marshal(responseObject)

	// Regarding response errors:
	// Note that we would use an error here for a low level thing like a network error.
	// An HTTP error like a 404 or 500 would be considered a valid response with an
	// appropriate status code
	r.Server.Respond(&http.Response{
		StatusCode:    http.StatusOK,
		Body:          ioutil.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}, nil)
}

func NewRootResource() *octopusApiClient.RootResource {
	root := octopusApiClient.NewRootResource()
	root.Links[constants.LinkSpaces] = "/api/spaces{/id}{?skip,ids,take,partialName}"

	// Note: all this stuff typically doesn't appear at the root resource level
	// has assigned a default space. We don't like default spaces, so the unit tests
	// should probably not mimic that structure; clean it up one day
	root.Links[constants.LinkProjects] = "/api/Spaces-1/projects{/id}{?name,skip,ids,clone,take,partialName,clonedFromProjectId}"
	root.Links[constants.LinkChannels] = "/api/Spaces-1/channels{/id}{?skip,take,ids,partialName}"
	root.Links[constants.LinkDeploymentProcesses] = "/api/Spaces-1/deploymentprocesses{/id}{?skip,take,ids}"
	root.Links[constants.LinkReleases] = "/api/Spaces-1/releases{/id}{?skip,ignoreChannelRules,take,ids}"
	root.Links[constants.LinkFeeds] = "/api/Spaces-1/feeds{/id}{?skip,take,ids,partialName,feedType,name}"
	return root
}