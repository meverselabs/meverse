package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ReqData struct {
	req   *jRPCRequest
	resCh *chan *JRPCResponse
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
	s.e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	s.e.POST("/api/endpoints/http", func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()

		var req jRPCRequest
		if err := dec.Decode(&req); err != nil {
			dec := json.NewDecoder(bytes.NewReader(body))
			dec.UseNumber()

			var req2 jRPCRequest2
			if err := dec.Decode(&req2); err != nil {
				return err
			}
			req.JSONRPC = req2.JSONRPC
			req.ID = req2.ID
			req.Method = req2.Method
			req.Params = make([]*json.Number, 0, len(req2.Params))
			for _, v := range req2.Params {
				s := fmt.Sprint(v)
				a := json.Number(s)
				req.Params = append(req.Params, &a)
			}
		}
		resCh := make(chan *JRPCResponse)
		reqCh <- &ReqData{
			req:   &req,
			resCh: &resCh,
		}
		/*
			res := s.handleJRPC(&req)
		*/
		res := <-resCh
		if res == nil {
			return c.NoContent(http.StatusOK)
		} else {
			return c.JSON(http.StatusOK, res)
		}
	})
	s.e.GET("/api/endpoints/websocket", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		Type := strings.ToLower(c.QueryParam("type"))
		switch Type {
		default:
			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					return err
				}
				dec := json.NewDecoder(bytes.NewReader(data))
				dec.UseNumber()

				var req jRPCRequest
				if err := dec.Decode(&req); err != nil {
					dec := json.NewDecoder(bytes.NewReader(data))
					dec.UseNumber()

					var req2 jRPCRequest2
					if err := dec.Decode(&req2); err != nil {
						return err
					}
					req.JSONRPC = req2.JSONRPC
					req.ID = req2.ID
					req.Method = req2.Method
					req.Params = make([]*json.Number, 0, len(req2.Params))
					for _, v := range req2.Params {
						s := fmt.Sprint(v)
						a := json.Number(s)
						req.Params = append(req.Params, &a)
					}
				}
				resCh := make(chan *JRPCResponse)
				reqCh <- &ReqData{
					req:   &req,
					resCh: &resCh,
				}
				/*
					res := s.handleJRPC(&req)
				*/
				res := <-resCh
				if res != nil {
					if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
						return err
					}
					if err := conn.WriteJSON(res); err != nil {
						return err
					}
				}
			}
		}
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
		return nil, ErrExistSubName
	}
	js := NewJRPCSub()
	s.subMap[SubName] = js
	return js, nil //TEMP
}

func (s *APIServer) handleJRPC(req *jRPCRequest) *JRPCResponse {
	ls := strings.SplitN(req.Method, ".", 2)
	if len(ls) != 2 {
		res := &JRPCResponse{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Error:   ErrInvalidMethod.Error(),
		}
		return res
	}

	args := []*string{}
	for _, v := range req.Params {
		if v == nil {
			args = append(args, nil)
		} else {
			args = append(args, (*string)(v))
		}
	}
	s.Lock()
	sub, has := s.subMap[ls[0]]
	s.Unlock()
	if !has {
		res := &JRPCResponse{
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
		if req.ID == nil {
			return nil
		} else {
			res := &JRPCResponse{
				JSONRPC: req.JSONRPC,
				ID:      req.ID,
				Error:   ErrInvalidMethod.Error(),
			}
			return res
		}
	}

	ret, err := fn(req.ID, NewArgument(args))
	if req.ID == nil {
		return nil
	} else {
		res := &JRPCResponse{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
		}
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Result = ret
		}
		return res
	}
}
