package httpx

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-errors/errors"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/session"
)

// HAR is an HTTP Archive.
// https://w3c.github.io/web-performance/specs/HAR/Overview.html
// http://www.softwareishard.com/blog/har-12-spec/
type HAR struct {
	in chan interface{}
}

// NewHAR creates a mew HAR format.
func NewHAR(_ session.Context) (flow.OutputFormat, error) {
	return &HAR{
		in: make(chan interface{}),
	}, nil
}

// In returns the input channel.
func (h *HAR) In() chan<- interface{} { return h.in }

// Init initializes the HAR sink.
func (h *HAR) Init(ctx session.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "har").
		Logger()

	ctx.Logger.Debug().Msg("init")

	c := NewCorrelator(ctx.Config.HTTPTimeout)
	c.Start(ctx)

	ctx.Logger.Debug().Msg("correlator started")

	hh := &har{
		Log: harLog{
			Version: "1.2",
			Creator: harCreator{
				Name:    "vhs",
				Version: "0.0.1",
			},
		},
	}

	for {
		select {
		case n := <-h.in:
			switch m := n.(type) {
			case Message:
				c.Messages <- m
				if ctx.Config.DebugHTTPMessages {
					ctx.Logger.Debug().Interface("m", m).Msg("received message")
				} else {
					ctx.Logger.Debug().Msg("received message")
				}
			}
		case r := <-c.Exchanges:
			h.addRequest(ctx, hh, r)
			if ctx.Config.DebugHTTPMessages {
				ctx.Logger.Debug().Interface("r", r).Msg("received request from correlator")
			} else {
				ctx.Logger.Debug().Msg("received request from correlator")
			}
		case <-ctx.StdContext.Done():
			if err := json.NewEncoder(w).Encode(hh); err != nil {
				ctx.Errors <- errors.Errorf("failed to encode to JSON: %w", err)
			}
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}

func (h *HAR) addRequest(ctx session.Context, hh *har, req *Request) {
	//var headers []harNVP
	//for n, vals := range req.Header {
	//	for _, v := range vals {
	//		headers = append(headers, harNVP{Name: n, Value: v})
	//	}
	//}

	//var queryString []harNVP
	//for n, vals := range req.URL.Query() {
	//	for _, v := range vals {
	//		queryString = append(queryString, harNVP{Name: n, Value: v})
	//	}
	//}

	request := harRequest{
		Method:      req.Method,
		URL:         req.URL.String(),
		HTTPVersion: req.Proto,
		Cookies:     extractCookies(req.Cookies),
		Headers:     mapToHarNVP(req.Header), //headers,
		QueryString: mapToHarNVP(req.URL.Query()), //queryString,
		PostData:    extractPostData(req),
		HeaderSize:  -1,
		BodySize:    len(req.Body),
	}

	var response harResponse
	if req.Response != nil {
		//var resHeaders []harNVP
		//for n, vals := range req.Response.Header {
		//	for _, v := range vals {
		//		resHeaders = append(resHeaders, harNVP{Name: n, Value: v})
		//	}
		//}

		content := harContent{
			Size:     req.Response.ContentLength,
			MimeType: req.Response.Header.Get("Content-Type"),
			Text:     req.Response.Body,
		}

		response = harResponse{
			Status:      req.Response.StatusCode,
			StatusText:  req.Response.Status,
			HTTPVersion: req.Response.Proto,
			Cookies:     extractCookies(req.Response.Cookies),
			Headers:     mapToHarNVP(req.Response.Header),//resHeaders,
			Content:     content,
			RedirectURL: req.Response.Location,
			HeadersSize: -1,
			BodySize:    len(req.Response.Body),
		}
	}

	entry := harEntry{
		StartedDateTime: req.Created.Format(time.RFC3339),
		Time:            req.Response.Created.Sub(req.Created).Milliseconds(),
		Request:         request,
		Response:        response,
		Cache:           harCache{},
		Timings: harEntryTiming{
			Send:    1,
			Wait:    1,
			Receive: 1,
		},
		ServerIPAddress: extractServerIP(req),
		Connection:      req.GetConnectionID(),
	}

	if ctx.Config.DebugHTTPMessages {
		ctx.Logger.Debug().Interface("entry", entry).Msg("adding entry")
	} else {
		ctx.Logger.Debug().Msg("adding entry")
	}

	hh.Log.Entries = append(hh.Log.Entries, entry)
}

// HELPERS, ETC.


func extractCookies(cookies []*http.Cookie) []harCookie {
	harCookies := make([]harCookie, len(cookies))
	for i, cookie := range cookies {
		if cookie != nil {
			harCookies[i] = harCookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				Expires:  cookie.Expires.Format(time.RFC3339),
				HTTPOnly: cookie.HttpOnly,
				Secure:   cookie.Secure,
				Comment:  "",
			}
		}
	}
	return harCookies
}

// extractPostData pulls the POST data out of a Request into a harPOST.
// Data of type application/x-www-form-urlencoded is parsed into name-value pairs and stored in the Params field.
// Data of other types (including multipart/form-data) is not parsed and is included as-is in the Text field.
// Field Params and field Text are mutually exclusive.
func extractPostData(req *Request) harPOST {
	var post harPOST
	if req.Method != http.MethodPost {
		return post
	}

	if req.PostForm != nil {
		post.Params = mapToHarNVP(req.PostForm)
	} else {
		post.Text = req.Body
	}

	return post
}

func extractServerIP(req *Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return ""
	}

	add, err := net.LookupHost(host)
	if err != nil {
		return ""
	}

	return strings.Join(add, ".")
}

func mapToHarNVP(m map[string][]string) []harNVP {
	var nvps []harNVP
	for n, vals := range m {
		for _, v := range vals {
			nvps = append(nvps, harNVP{Name: n, Value: v})
		}
	}
	return nvps
}

// HAR FORMAT DEFINITION STRUCTS
//har is the root of a HTTP Archive (HAR) file.
type har struct {
	Log harLog `json:"log"`
}

//harLog is the topmost object in a HAR file.
type harLog struct {
	Version string     `json:"version"`           //Required
	Creator harCreator `json:"creator"`           //Required
	//Browser harCreator `json:"browser,omitempty"` //Optional
	//Pages   []harPage  `json:"pages,omitempty"`   //Optional
	Entries []harEntry `json:"entries"`           //Required
	Comment string     `json:"comment"`           //Optional
}

//harCreator is the object used for Creator and Browser entries at the harLog level.
type harCreator struct {
	Name    string `json:"name"`    //Required
	Version string `json:"version"` //Required
	Comment string `json:"comment"` //Optional
}

//harPage is the object used for the Pages entry at the harLog level.
//type harPage struct {
//	StartedDateTime string          `json:"startedDateTime,omitempty"` //Required
//	ID              string          `json:"id,omitempty"`              //Required
//	Title           string          `json:"title,omitempty"`           //Required
//	PageTimings     []harPageTiming `json:"pageTimings,omitempty"`     //Optional TODO: Does this depend on browserish things?
//	Comment         string          `json:comment,omitempty`           //Optional
//}

//harPageTiming is the object for the PageTimings entry at the harPage level.
//type harPageTiming struct { //TODO
//	OnContentLoad int64  `json:"onContentLoad,omitempty"` //Optional TODO: Depends on browser events; -1 if unknown
//	OnLoad        int64  `json:"onLoad,omitempty"`        //Optional TODO: Depends on browser events; -1 if unknown
//	Comment       string `json:"comment,omitempty"`       //Optional
//}

//harEntry is the object for the Entries entry at the harLog level.
type harEntry struct {
	Pageref         string         `json:"pageref,omitempty"`         //Optional
	StartedDateTime string         `json:"startedDateTime,omitempty"` //Required; timestamp of request start in ISO 8601 format
	Time            int64          `json:"time,omitempty"`            //Required
	Request         harRequest     `json:"request,omitempty"`         //Required
	Response        harResponse    `json:"response,omitempty"`        //Required
	Cache           harCache       `json:"cache,omitempty"`           //Optional TODO: Probably doesn't make sense in VHS. Leave empty ("{}") or omit if not available.
	Timings         harEntryTiming `json:"timings,omitempty"`         //Required TODO: Not sure if this one makes sense or not.
	ServerIPAddress string         `json:"serverIPAddress,omitempty"` //Optional
	Connection      string         `json:"connection,omitempty"`      //Optional
	Comment         string         `json:"comment,omitempty"`         //Optional
}

//harRequest is the object for the Request entry at the harEntry level.
type harRequest struct {
	Method      string      `json:"method,omitempty"`      //Required
	URL         string      `json:"url,omitempty"`         //Required
	HTTPVersion string      `json:"httpVersion,omitempty"` //Required
	Cookies     []harCookie `json:"cookies,omitempty"`     //Required? If present in the request. TODO: Might not be applicable to VHS
	Headers     []harNVP    `json:"headers,omitempty"`     //Required
	QueryString []harNVP    `json:"queryString,omitempty"` //Required
	PostData    harPOST     `json:"postData,omitempty"`    //Optional TODO: Probably worth implementing
	HeaderSize  int         `json:"headerSize,omitempty"`  //Required; -1 if data unavailable
	BodySize    int         `json:"bodySize,omitempty"`    //Required; -1 if data unavailable
	Comment     string      `json:"comment,omitempty"`     //Optional
}

//harResponse is the object for the Response entry at the harEntry level.
type harResponse struct {
	Status      int         `json:"status,omitempty"`      //Required
	StatusText  string      `json:"statusText,omitempty"`  //Required
	HTTPVersion string      `json:"httpVersion,omitempty"` //Required
	Cookies     []harCookie `json:"cookies,omitempty"`     //Required? If present in the request. TODO: Might not be applicable to VHS
	Headers     []harNVP    `json:"headers,omitempty"`     //Required
	Content     harContent  `json:"content,omitempty"`     //Required
	RedirectURL string      `json:"redirectURL,omitempty"` //Required if applicable?
	HeadersSize int         `json:"headersSize,omitempty"` //Required; -1 if data unavailable.
	BodySize    int         `json:"bodySize,omitempty"`    //Required; -1 if data unavailable
	Comment     string      `json:"comment,omitempty"`     //Optional
}

//harCookie is the object for the Cookies entry at the harRequest and harResponse levels.
type harCookie struct { //TODO
	Name     string `json:"name"`               //Required
	Value    string `json:"value"`              //Required
	Path     string `json:"path,omitempty"`     //Optional
	Domain   string `json:"domain,omitempty"`   //Optional
	Expires  string `json:"expires,omitempty"`  //Optional. Cookie expiration time in ISO 8601 format
	HTTPOnly bool   `json:"httpOnly,omitempty"` //Optional
	Secure   bool   `json:"secure,omitempty"`   //Optional
	Comment  string `json:"comment,omitempty"`  //Optional
}

//harPOST is the object for the PostData entry at the harRequest level.
type harPOST struct { //TODO
	MIMEType string   `json:"mimeType"`          //Required
	Params   []harNVP `json:"params,omitempty"`  //Mutually exclusive with Text
	Text     string   `json:"text,omitempty"`    //Mutually exclusive with Params
	Comment  string   `json:"comment,omitempty"` //Optional
}

//harNVP is a name-value pair and is used at harRequest, harResponse, and harPOST levels.
type harNVP struct {
	Name    string `json:"name"`              //Required
	Value   string `json:"value"`             //Required
	Comment string `json:"comment,omitempty"` //Optional
}

//harContent is the object for the Content entry at the harResponse level.
type harContent struct {
	Size        int64  `json:"size"`                  //Required
	Compression int    `json:"compression,omitempty"` //Optional
	MimeType    string `json:"mimeType"`              //Required
	Text        string `json:"text,omitempty"`        //Optional
	Encoding    string `json:"encoding,omitempty"`    //Optional
	Comment     string `json:"comment,omitempty"`     //Optional
}

//harCache is the object for the Cache entry at the harEntry level.
type harCache struct { //TODO: Probably doesn't make sense in VHS. Leave empty ("{}") or omit if not available.
	//BeforeRequest harCacheStatus `json:"beforeRequest,omitempty"` //Optional
	//AfterRequest  harCacheStatus `json:"afterRequest,omitempty"`  //Optional
	//Comment       string         `json:"comment,omitempty"`       //Optional
}

//harCacheStatus is the object for the BeforeRequest and AfterRequest entries at the harCache level.
//type harCacheStatus struct { //TODO: might not be needed if cache is omitted or left blank.
//	Expires    string `json:"expires,omitempty"`    //Optional
//	LastAccess string `json:"lastAccess,omitempty"` //Required
//	ETag       string `json:"eTag,omitempty"`       //Required
//	HitCount   int64  `json:"hitCount,omitempty"`   //Required
//	Comment    string `json:"comment,omitempty"`    //Optional
//}

//harEntryTiming is the object for the Timings entry at the harEntry level.
type harEntryTiming struct {
	Blocked int64  `json:"blocked,omitempty"` //Optional
	DNS     int64  `json:"dns,omitempty"`     //Optional
	Connect int64  `json:"connect,omitempty"` //Optional
	Send    int64  `json:"send"`              //Required
	Wait    int64  `json:"wait"`              //Required
	Receive int64  `json:"receive"`           //Required
	SSL     int64  `json:"ssl,omitempty"`     //Optional
	Comment string `json:"comment,omitempty"` //Optional
}
