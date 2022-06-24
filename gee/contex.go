package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int

	// middleware
	handlers []HandlerFunc
	index    int
	engine   *Engine //添加这个变量，之后就可以通过Context 访问Engine中的HTML模板
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Path:    r.URL.Path,
		Method:  r.Method,
		index:   -1,
	}
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//获取动态路由中的参数
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// 获取表单参数
func (c *Context) PostFrom(key string) string {
	return c.Request.FormValue(key)
}

// 获取url参数
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// 设置状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Context-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) Fail(code int, err string) {
	c.String(code, err)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Context-Type", "application/json")
	c.Status(code)
	//指定一个输入流，获得一个encoder
	encoder := json.NewEncoder(c.Writer)
	//将内容序列化到指定数据流中
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500) //帮我们返回错误
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Context-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
