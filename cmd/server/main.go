package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	pb "taggy-gateway/proto"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	LEGACY_BACKEND_PORT = 3030
	TIME_OUT_SECOND     = 15
)

func main() {

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	fc := pb.NewFetcherClient(conn)

	r := gin.Default()
	r.Use(cors.Default())

	r.PATCH("/v1/rss/all", func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), TIME_OUT_SECOND*time.Second)
		defer cancel()
		rsp, err := fc.FetchAllRSS(ctx, &pb.FetchAllRSSRequest{})

		if err != nil {
			log.Println(err)
		}
		c.JSON(200, gin.H{
			"message": rsp.GetMessage(),
		})
	})

	//Legacy backend end point
	r.POST("/route/article/import", ReverseProxy())
	r.GET("/route/rss", ReverseProxy())
	r.POST("/route/rss/search", ReverseProxy())
	r.POST("/route/rss/fetch", ReverseProxy())
	r.GET("/route/rss/userfeeds", ReverseProxy())
	r.GET("/route/rss/feedtags", ReverseProxy())

	r.Run(":3000")
}

const (
	defaultName = "kevin"
)

func ReverseProxy() gin.HandlerFunc {

	target := fmt.Sprintf("localhost:%d", LEGACY_BACKEND_PORT)

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			// r := c.Request
			req.Header = c.Request.Header
			req.URL.Scheme = "http"
			req.URL.Host = target
			// req.Header["my-header"] = []string{r.Header.Get("my-header")}

			// delete(req.Header, "My-Header")
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
