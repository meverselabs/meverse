package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ReqData struct {
	req   *jRPCRequest
	resCh *chan interface{}
}

// Run starts web service of the apiserver
func (s *APIServer) Run(BindAddress string) error {
	reqCh := make(chan *ReqData)

	s.e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err, c.Request().URL.String())
		c.HTML(code, err.Error())
	}
	_callCount := 0
	s.e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	s.e.POST("/", func(c echo.Context) error {
		_callCount++
		// callCount := _callCount
		body, err := ioutil.ReadAll(c.Request().Body)
		// log.Println("body", callCount, string(body))
		if err != nil {
			return errors.WithStack(err)
		}
		defer c.Request().Body.Close()

		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()

		var req jRPCRequest
		if err := dec.Decode(&req); err != nil {
			return errors.WithStack(err)
		}
		resCh := make(chan interface{})
		reqCh <- &ReqData{
			req:   &req,
			resCh: &resCh,
		}
		res := <-resCh
		if res == nil {
			return c.NoContent(http.StatusOK)
		} else {
			// log.Println("response result", callCount, res)
			return c.JSON(http.StatusOK, res)
		}
	})
	s.e.GET("/", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			return errors.WithStack(err)
		}
		defer conn.Close()

		// _type := strings.ToLower(c.QueryParam("type"))
		// switch _type {
		// default:
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return errors.WithStack(err)
			}
			dec := json.NewDecoder(bytes.NewReader(data))
			dec.UseNumber()

			var req jRPCRequest
			if err := dec.Decode(&req); err != nil {
				return errors.WithStack(err)
			}
			resCh := make(chan interface{})
			reqCh <- &ReqData{
				req:   &req,
				resCh: &resCh,
			}
			select {
			case res := <-resCh:
				if res != nil {
					if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
						return errors.WithStack(err)
					}
					if err := conn.WriteJSON(res); err != nil {
						return errors.WithStack(err)
					}
				}
			}
		}
		// }
	})

	s.e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	for i := 0; i < 50; i++ {
		go func() {
			for r := range reqCh {
				res := s.handleJRPC(r.req)
				(*r.resCh) <- res
			}
		}()
	}
	return s.e.Start(BindAddress)
}

// JRPC provides the json rpc feature as a SubName.FunctionName methods
func (s *APIServer) JRPC(SubName string) (*JRPCSub, error) {
	s.Lock()
	defer s.Unlock()

	if _, has := s.subMap[SubName]; has {
		return nil, errors.WithStack(ErrExistSubName)
	}
	js := NewJRPCSub()
	s.subMap[SubName] = js
	return js, nil //TEMP
}

func (s *APIServer) handleJRPC(req *jRPCRequest) interface{} {
	method := req.Method
	if !strings.Contains(method, ".") {
		method = "eth." + method
	}
	ls := strings.SplitN(method, ".", 2)
	if len(ls) != 2 {
		log.Printf("failJRPCResponse %v %+v\n", method, ErrInvalidMethod)
		res := &JRPCResponseWithError{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Error:   ErrInvalidMethod.Error(),
		}
		return res
	}

	args := []interface{}{}
	for _, v := range req.Params {
		if v == nil {
			args = append(args, nil)
		} else {
			args = append(args, v)
		}
	}
	s.Lock()
	sub, has := s.subMap[ls[0]]
	s.Unlock()
	if !has {
		log.Printf("failJRPCResponse %v, %+v\n", method, ErrInvalidMethod)
		res := &JRPCResponseWithError{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Error:   ErrInvalidMethod.Error(),
		}
		return res
	}

	sub.Lock()
	fn, has := sub.funcMap[ls[1]]
	sub.Unlock()
	if !has {
		log.Println("failJRPCResponse", method, req.Method, printParam(req.Params))
		if req.ID == nil {
			return nil
		} else {
			res := &JRPCResponseWithError{
				JSONRPC: req.JSONRPC,
				ID:      req.ID,
				Error:   ErrInvalidMethod.Error(),
			}
			return res
		}
	}

	ret, err := fn(req.ID, NewArgument(args))
	if req.ID == nil {
		log.Println("failJRPCResponse err not fount id", err, req.Method, printParam(req.Params))
		return &JRPCResponseWithError{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Error:   "not fount id",
		}
	} else {
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return &JRPCResponse{
					JSONRPC: req.JSONRPC,
					ID:      req.ID,
					Result:  nil,
				}
			} else {
				log.Printf("failJRPCResponse err %v %v %+v\n", req.Method, printParam(req.Params), err)
				return &JRPCResponseWithError{
					JSONRPC: req.JSONRPC,
					ID:      req.ID,
					Error:   err.Error(),
				}
			}
		} else {
			return &JRPCResponse{
				ID:      req.ID,
				JSONRPC: req.JSONRPC,
				Result:  ret,
			}
		}
	}
}

// add link to apiserver
func (s *APIServer) AddGETPath(path string, con echo.HandlerFunc) error {
	s.e.GET(path, con)
	return nil
}

func printParam(param []interface{}) (r string) {
	for _, v := range param {
		r = r + ", " + fmt.Sprintf("%v", v)
	}
	return
}
